package app

import (
	"strings"
	"sync"

	"github.com/Doom-z/RepClient/client"
	"github.com/Doom-z/RepClient/client/model"
	"github.com/Doom-z/RepClient/cmd/app/cfg"
	"github.com/Doom-z/RepClient/pkg/logger"
	"github.com/Doom-z/RepClient/pkg/utils"
)

func Run(args Args, cfg cfg.Conf) {
	c, err := client.NewClient(
		cfg.Api.Host,
		client.WithPageSize(args.PageSize),
		client.WithApiKey(cfg.Api.Apikey),
	)
	if err != nil {
		logger.Fatalf("Client init error: %v", err)
	}

	max := args.MaxTotalOutputIp

	switch {
	case args.Ipv6 != "" && args.ModeFull:
		processAAAARecords(c, args.Ipv6)
	case args.Ipv4 != "" && args.ModeFull:
		processARecords(c, args.Ipv4)
	case args.ListFile != "":
		stream := StreamFile(args.ListFile)
		var wg sync.WaitGroup
		jobs := make(chan string)

		for i := 0; i < args.Threads; i++ {
			logger.WithGID().Tracef("Starting worker %d", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				var workerWg sync.WaitGroup
				for line := range jobs {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					workerWg.Add(1)
					go func(target string) {
						defer workerWg.Done()
						processGenericStream(c, line, args.PageSize, max)
					}(line)
				}
				workerWg.Wait()
			}()
		}
		// feed jobs to ch to proccess based on the worker aka threads
		for line := range stream {
			line = strings.TrimSpace(line)
			if line != "" {
				jobs <- line
			}
		}
		close(jobs)
		wg.Wait()
	default:
		if args.Ipv6 != "" {
			logger.Fatal("You must use --full, -f to query ipv6")
		}
		argMap := map[string]string{
			"ip":    args.Ipv4,
			"ns":    args.Ns,
			"cname": args.Cname,
			"txt":   args.Txt,
			"mx":    args.Mx,
		}

		var (
			param, target string
			found         bool
		)

		for k, v := range argMap {
			if v != "" {
				param = k
				target = v
				found = true
				break
			}
		}

		if !found {
			logger.Fatal("You must provide at least one argument: --ip, --ns, --cname, --txt, or --mx")
		}

		processStream(c, param, target, args.PageSize, max)
	}
}

func processAAAARecords(c *client.Client, ipv6 string) {
	recordsChan, errChan := client.FetchDNSRecords[model.AAAARecord](c, "aaaa", ipv6)
	count := 0
	for {
		select {
		case record, ok := <-recordsChan:
			if !ok {
				recordsChan = nil
			} else {
				count++
				logger.WithFields(map[string]any{
					"domain":   record.DomainID,
					"ip":       record.IP,
					"asn":      record.ASN,
					"asn_name": record.ASNName,
					"country":  record.Country,
					"city":     record.City,
					"latlong":  record.LatLong,
				}).Info("Record found")
			}

		case err, ok := <-errChan:
			if ok && err != nil {
				logger.Fatalf("Client fetch error: %v", err)
			}
			errChan = nil
		}

		if recordsChan == nil && errChan == nil {
			break
		}
	}

	logger.Infof("Total AAAA records for %s: %d", ipv6, count)
}

func processARecords(c *client.Client, ipv4 string) {
	recordsChan, errChan := client.FetchDNSRecords[model.ARecord](c, "a", ipv4)

	count := 0
	for {
		select {
		case record, ok := <-recordsChan:
			if !ok {
				recordsChan = nil
			} else {
				count++
				logger.WithFields(map[string]any{
					"domain":   record.DomainID,
					"ip":       record.IP,
					"asn":      record.ASN,
					"asn_name": record.ASNName,
					"country":  record.Country,
					"city":     record.City,
					"latlong":  record.LatLong,
				}).Info("Record found")
			}

		case err, ok := <-errChan:
			if ok && err != nil {
				logger.Fatalf("Client fetch error: %v", err)
			}
			errChan = nil
		}

		if recordsChan == nil && errChan == nil {
			break
		}
	}

	logger.Infof("Total A records for %s: %d", ipv4, count)
}

func processGenericStream(c *client.Client, input string, pageSize int, max int) {
	param := utils.DetectRecordType(input)
	if param == "" {
		logger.Warnf("Could not detect record type for: %s", input)
		return
	}

	processStream(c, param, input, pageSize, max)
}

func processStream(c *client.Client, param, target string, pageSize int, max int) {
	logger.Tracef("Fetching (%s) records for %s with max records: %d", param, target, max)
	recordsChan, errChan := c.FetchRecordsStream(param, target)
	var count int

	for {
		select {
		case record, ok := <-recordsChan:
			if !ok {
				recordsChan = nil
			} else {
				count++
				logger.WithGID().Tracef("%s -> %s (%s) at %d", record.IP, record.DomainID, record.RecordType, record.Timestamp)
				if count%pageSize == 0 {
					logger.WithGID().Debugf("Fetched %d (%s) records for %s", count, param, target)
				}
				if max > 0 && count >= max {
					recordsChan = nil
					errChan = nil
				}
			}

		case err, ok := <-errChan:
			if ok && err != nil {
				logger.Printf("Client fetch error: %v", err)
			}
			errChan = nil
		}

		if recordsChan == nil && errChan == nil {
			break
		}
	}

	logger.WithFields(map[string]any{
		"param": param,
		"type":  target,
		"total": count,
	}).Infof("Successfully fetched all records")
}
