package run

import (
	"fmt"
	"sync"

	"github.com/Doom-z/RepClient/client/model"
	"github.com/Doom-z/RepClient/pkg/fileutil"
	"github.com/Doom-z/RepClient/pkg/logger"
)

func (r *Run) fetchAndSaveRecords(param, target string) {
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

func (r *Run) fetchARecordStream(ipv4 string) {
	outputPath := fmt.Sprintf("%s/a.%s", r.Cfg.Output.Dir, r.Cfg.Output.Format)
	saveTasks := make(chan SaveTask, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go startSaveWorker(&wg, saveTasks)

	processTypedStream(r.Client, "a", ipv4, func(record model.ARecord) {
		logger.WithFields(map[string]any{
			"domain":   record.DomainID,
			"ip":       record.IP,
			"asn":      record.ASN,
			"asn_name": record.ASNName,
			"country":  record.Country,
			"city":     record.City,
			"latlong":  record.LatLong,
		}).Info("A Record found")
	}, saveTasks, outputPath, r.Cfg.Output.Format, r.Args.Output)

	close(saveTasks)
	wg.Wait()
}

func (r *Run) fetchAAAARecordStream(ipv6 string) {
	outputPath := fmt.Sprintf("%s/aaaa.%s", r.Cfg.Output.Dir, r.Cfg.Output.Format)
	saveTasks := make(chan SaveTask, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go startSaveWorker(&wg, saveTasks)

	processTypedStream(r.Client, "aaaa", ipv6, func(record model.AAAARecord) {
		logger.WithFields(map[string]any{
			"domain":   record.DomainID,
			"ip":       record.IP,
			"asn":      record.ASN,
			"asn_name": record.ASNName,
			"country":  record.Country,
			"city":     record.City,
			"latlong":  record.LatLong,
		}).Info("AAAA Record found")
	}, saveTasks, outputPath, r.Cfg.Output.Format, r.Args.Output)

	close(saveTasks)
	wg.Wait()
}
