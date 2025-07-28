package app

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Doom-z/RepClient/client"
	"github.com/Doom-z/RepClient/client/model"
	"github.com/Doom-z/RepClient/cmd/app/cfg"
	"github.com/Doom-z/RepClient/pkg/fileutil"
	"github.com/Doom-z/RepClient/pkg/logger"
	"github.com/Doom-z/RepClient/pkg/utils"
)

type Run struct {
	Client *client.Client
	Args   Args
	Cfg    cfg.Conf
}

func NewRun(args Args, cfg cfg.Conf) (*Run, error) {
	c, err := client.NewClient(
		cfg.Api.Host,
		client.WithPageSize(args.PageSize),
		client.WithApiKey(cfg.Api.Apikey),
	)
	if err != nil {
		return nil, fmt.Errorf("client init error: %w", err)
	}

	if args.Output {
		if err := fileutil.EnsureDir(cfg.Output.Dir); err != nil {
			return nil, err
		}
	}

	return &Run{
		Client: c,
		Args:   args,
		Cfg:    cfg,
	}, nil
}

func (r *Run) Start() {
	args := r.Args

	switch {
	case args.Trial && args.ListFile == "":
		r.handleSingleQueryFree()
	case args.Trial && args.ListFile != "":
		r.processListFileFree()

	case args.Ipv6 != "" && args.ModeFull:
		r.fetchAAAARecords(args.Ipv6)
	case args.Ipv4 != "" && args.ModeFull:
		r.fetchARecords(args.Ipv4)

	case args.ListFile != "" && !args.Trial:
		r.processListFile()

	default:
		r.handleSingleQuery()
	}
}

func processTypedRecords[T any](c *client.Client, recordType string, ip string, logFn func(T)) {
	recordsCh, errCh := client.FetchDNSRecords[T](c, recordType, ip)
	count := 0

	for {
		select {
		case record, ok := <-recordsCh:
			if !ok {
				recordsCh = nil
			} else {
				count++
				logFn(record)
			}

		case err, ok := <-errCh:
			if ok && err != nil {
				logger.Fatalf("Client fetch error: %v", err)
			}
			errCh = nil
		}

		if recordsCh == nil && errCh == nil {
			break
		}
	}
	logger.Infof("Total %s records for %s: %d", strings.ToUpper(recordType), ip, count)
}

func (r *Run) fetchAAAARecords(ipv6 string) {
	outputPath := fmt.Sprintf("%s/aaaa.%s", r.Cfg.Output.Dir, r.Cfg.Output.Format)

	processTypedRecords(r.Client, "aaaa", ipv6, func(record model.AAAARecord) {
		if r.Args.Output {
			if r.Cfg.Output.Format == "txt" {
				fileutil.SaveData(record.DomainID, outputPath, "append")
			} else {
				fileutil.SaveData(record, outputPath, "append")
			}
		}
		logger.WithFields(map[string]any{
			"domain":   record.DomainID,
			"ip":       record.IP,
			"asn":      record.ASN,
			"asn_name": record.ASNName,
			"country":  record.Country,
			"city":     record.City,
			"latlong":  record.LatLong,
		}).Info("AAAA Record found")
	})
}

func (r *Run) fetchARecords(ipv4 string) {
	processTypedRecords(r.Client, "a", ipv4, func(record model.ARecord) {
		if r.Args.Output {
			outputPath := fmt.Sprintf("%s/a.%s", r.Cfg.Output.Dir, r.Cfg.Output.Format)
			if r.Cfg.Output.Format == "txt" {
				fileutil.SaveData(record.DomainID, outputPath, "append")
			} else {
				fileutil.SaveData(record, outputPath, "append")
			}
		}
		logger.WithFields(map[string]any{
			"domain":   record.DomainID,
			"ip":       record.IP,
			"asn":      record.ASN,
			"asn_name": record.ASNName,
			"country":  record.Country,
			"city":     record.City,
			"latlong":  record.LatLong,
		}).Info("A Record found")
	})
}

func (r *Run) processListFile() {
	stream := StreamFile(r.Args.ListFile)
	jobs := make(chan string, r.Args.Threads*2)

	var wg sync.WaitGroup
	for i := 0; i < r.Args.Threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			logger.WithGID().Tracef("Worker %d started", workerID)
			var innerWg sync.WaitGroup

			for line := range jobs {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				innerWg.Add(1)

				go func(target string) {
					defer innerWg.Done()
					r.processGenericStream(target, func(param string, target string) {
						r.processStream(param, target)
					})
				}(line)
			}

			innerWg.Wait()
		}(i)
	}

	// Feed jobs
	for line := range stream {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			jobs <- trimmed
		}
	}

	close(jobs)
	wg.Wait()
}

func (r *Run) processListFileFree() {
	stream := StreamFile(r.Args.ListFile)
	jobs := make(chan string, r.Args.Threads*2)

	var wg sync.WaitGroup
	for i := 0; i < r.Args.Threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			logger.WithGID().Tracef("Worker %d started", workerID)
			var innerWg sync.WaitGroup

			for line := range jobs {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				innerWg.Add(1)

				go func(target string) {
					defer innerWg.Done()
					r.processGenericStream(target, func(param string, target string) {
						r.processFetch(param, target)
					})
				}(line)
			}

			innerWg.Wait()
		}(i)
	}

	// Feed jobs
	for line := range stream {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			jobs <- trimmed
		}
	}

	close(jobs)
	wg.Wait()
}

func (r *Run) processGenericStream(input string, handler func(param string, target string)) {
	param := utils.DetectRecordType(input)
	if param == "" {
		logger.Warnf("Could not detect record type for: %s", input)
		return
	}
	handler(param, input)
	// r.processStream(param, input)
}

func (r *Run) processStream(param, target string) {
	logger.Tracef("Fetching (%s) records for %s with max records: %d", param, target, r.Args.MaxTotalOutputIp)

	recordsCh, errCh := r.Client.FetchRecordsStream(param, target)
	count := 0
	pageSize := r.Args.PageSize
	max := r.Args.MaxTotalOutputIp
	outputPath := fmt.Sprintf("%s/stream.%s", r.Cfg.Output.Dir, r.Cfg.Output.Format)

	for {
		select {
		case record, ok := <-recordsCh:
			if !ok {
				if count > 0 {
					logger.WithGID().Debugf("Fetched %d (%s) records for %s", count, param, target)
				}
				recordsCh = nil
			} else {
				count++
				if r.Args.Output {
					if r.Cfg.Output.Format == "txt" {
						fileutil.SaveData(record.DomainID, outputPath, "append")
					} else {
						fileutil.SaveData(record, outputPath, "append")
					}
				}

				logger.WithGID().Tracef("%s -> %s (%s) at %d", record.IP, record.DomainID, record.RecordType, record.Timestamp)

				if count%pageSize == 0 {
					logger.WithGID().Debugf("Fetched %d (%s) records for %s", count, param, target)
				}

				if max > 0 && count >= max {
					recordsCh = nil
					errCh = nil
				}
			}

		case err, ok := <-errCh:
			if ok && err != nil {
				logger.Printf("Client fetch error: %v", err)
			}
			errCh = nil
		}

		if recordsCh == nil && errCh == nil {
			break
		}
	}

	logger.WithFields(map[string]any{
		"param": param,
		"type":  target,
		"total": count,
	}).Infof("Successfully fetched all records")
}

func (r *Run) processFetch(param, target string) {
	logger.Tracef("Fetching (%s) records for %s with max records: %d", param, target, r.Args.MaxTotalOutputIp)

	records, err := r.Client.FetchRecords(param, target)
	if err != nil {
		logger.Warnf("Client fetch error: %v", err)
	}
	outputPath := fmt.Sprintf("%s/stream.%s", r.Cfg.Output.Dir, r.Cfg.Output.Format)

	if (len(records) > 0) && r.Args.Output {
		for _, record := range records {
			if r.Cfg.Output.Format == "txt" {
				fileutil.SaveData(record.DomainID, outputPath, "append")
			} else {
				fileutil.SaveData(record, outputPath, "append")
			}
		}
	}

	logger.WithFields(map[string]any{
		"param": param,
		"type":  target,
		"total": len(records),
	}).Infof("Successfully fetched all records")
}

func (r *Run) handleSingleQueryFree() {
	args := r.Args
	if args.Ipv6 != "" {
		logger.Fatal("This features only work in paid plans")
		return
	}
	argMap := map[string]string{
		"ip":    args.Ipv4,
		"ns":    args.Ns,
		"cname": args.Cname,
		"txt":   args.Txt,
		"mx":    args.Mx,
	}

	for k, v := range argMap {
		if v != "" {
			r.processFetch(k, v)
			return
		}
	}
	logger.Fatal("You must provide at least one of the following: --ip, --ns, --cname, --txt, --mx")
}

func (r *Run) handleSingleQuery() {
	args := r.Args

	if args.Ipv6 != "" {
		if !args.ModeFull {
			logger.Fatal("You must use --full, -f to query ipv6")
		}
		r.fetchAAAARecords(args.Ipv6)
		return
	}

	if args.Ipv4 != "" && args.ModeFull {
		r.fetchARecords(args.Ipv4)
		return
	}

	argMap := map[string]string{
		"ip":    args.Ipv4,
		"ns":    args.Ns,
		"cname": args.Cname,
		"txt":   args.Txt,
		"mx":    args.Mx,
	}

	for k, v := range argMap {
		if v != "" {
			r.processStream(k, v)
			return
		}
	}

	logger.Fatal("You must provide at least one of the following: --ip, --ns, --cname, --txt, --mx")
}
