package run

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Doom-z/RepClient/client"
	"github.com/Doom-z/RepClient/pkg/fileutil"
	"github.com/Doom-z/RepClient/pkg/logger"
	"github.com/Doom-z/RepClient/pkg/utils"
)

func StreamFile(file string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		f, err := os.Open(file)
		if err != nil {
			logger.Fatalf("failed to open file %s: %v", file, err)
			return
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				out <- line
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Fatalf("error reading file %s: %v", file, err)
		}
	}()

	return out
}

func (r *Run) handleStreamInput(input string, handler func(param string, target string)) {
	param := utils.DetectRecordType(input)
	if param == "" {
		logger.Warnf("Could not detect record type for: %s", input)
		return
	}
	handler(param, input)
}

func (r *Run) processStreamRecords(param, target string) {
	logger.Tracef("Fetching (%s) records for %s with max records: %d", param, target, r.Args.MaxTotalOutputIp)

	recordsCh, errCh := r.Client.FetchRecordsStream(param, target)
	count := 0
	pageSize := r.Args.PageSize
	max := r.Args.MaxTotalOutputIp
	outputPath := fmt.Sprintf("%s/stream.%s", r.Cfg.Output.Dir, r.Cfg.Output.Format)

	saveCh := make(chan SaveTask, r.Args.PageSize)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range saveCh {
			switch task.Format {
			case "txt":
				if s, ok := task.Data.(string); ok {
					fileutil.SaveData(s, task.Path, "append")
				}
			default:
				fileutil.SaveData(task.Data, task.Path, "append")
			}
		}
	}()

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
					var data any
					if r.Cfg.Output.Format == "txt" {
						data = record.DomainID
					} else {
						data = record
					}

					select {
					case saveCh <- SaveTask{Data: data, Format: r.Cfg.Output.Format, Path: outputPath}:
					default:
						logger.Warnf("Save channel full, dropping record: %v", data)
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
				logger.Warnf("Client fetch error: %v", err)
			}
			errCh = nil
		}

		if recordsCh == nil && errCh == nil {
			break
		}
	}

	close(saveCh)
	wg.Wait()

	logger.WithFields(map[string]any{
		"param": param,
		"type":  target,
		"total": count,
	}).Infof("Successfully fetched all records")
}

func processTypedStream[T HasDomainID](
	c *client.Client,
	recordType, ip string,
	logFn func(T),
	saveCh chan<- SaveTask,
	outputPath, format string,
	shouldSave bool,
) {
	recordsCh, errCh := client.FetchDNSRecords[T](c, recordType, ip)

	count := 0
	for {
		select {
		case record, ok := <-recordsCh:
			if !ok {
				recordsCh = nil
				continue
			}
			count++
			logFn(record)

			if shouldSave {
				var data any
				if format == "txt" {
					data = record.GetDomainID()
				} else {
					data = record
				}

				saveCh <- SaveTask{
					Data:   data,
					Path:   outputPath,
					Format: format,
				}
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
