package run

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Doom-z/RepClient/client"
	"github.com/Doom-z/RepClient/cmd/app/args"
	"github.com/Doom-z/RepClient/cmd/app/cfg"
	"github.com/Doom-z/RepClient/pkg/fileutil"
	"github.com/Doom-z/RepClient/pkg/logger"
)

type Run struct {
	Client *client.Client
	Args   args.Args
	Cfg    cfg.Conf
}

func NewRun(args args.Args, cfg cfg.Conf) (*Run, error) {
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
		r.runTrialSingleIP()
	case args.Trial && args.ListFile != "":
		r.runTrialFromFile() // a.k.a bulk scan from file

	case args.Ipv6 != "" && args.ModeFull:
		r.runFullIPv6Scan(args.Ipv6)
	case args.Ipv4 != "" && args.ModeFull:
		r.runFullIPv4Scan(args.Ipv4)

	case args.ListFile != "" && !args.Trial:
		r.runBulkScanFromFile()

	default:
		r.runSingleIPScan()
	}
}

func (r *Run) runTrialSingleIP() {
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
			r.fetchAndSaveRecords(k, v)
			return
		}
	}
	logger.Fatal("You must provide at least one of the following: --ip, --ns, --cname, --txt, --mx")
}

func (r *Run) runTrialFromFile() {
	stream := StreamFile(r.Args.ListFile)
	jobs := make(chan string, r.Args.Threads*2)

	var wg sync.WaitGroup
	for i := 0; i < r.Args.Threads; i++ {
		wg.Add(1)
		go r.runWorker(jobs, &wg, i, func(param, target string) {
			r.fetchAndSaveRecords(param, target)
		})
	}

	for line := range stream {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			jobs <- trimmed
		}
	}

	close(jobs)
	wg.Wait()
}

func (r *Run) runFullIPv6Scan(ipv6 string) {
	r.fetchAAAARecordStream(ipv6)
}

func (r *Run) runFullIPv4Scan(ipv4 string) {
	r.fetchARecordStream(ipv4)
}

func (r *Run) runBulkScanFromFile() {
	stream := StreamFile(r.Args.ListFile)
	jobs := make(chan string, r.Args.Threads*2)

	var wg sync.WaitGroup
	for i := 0; i < r.Args.Threads; i++ {
		wg.Add(1)
		go r.runWorker(jobs, &wg, i, func(param, target string) {
			r.processStreamRecords(param, target)
		})
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

func (r *Run) runSingleIPScan() {
	args := r.Args

	if args.Ipv6 != "" {
		if !args.ModeFull {
			logger.Fatal("You must use --full, -f to query ipv6")
		}
		r.fetchAAAARecordStream(args.Ipv6)
		return
	}

	if args.Ipv4 != "" && args.ModeFull {
		r.fetchARecordStream(args.Ipv4)
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
			r.processStreamRecords(k, v)
			return
		}
	}

	logger.Fatal("You must provide at least one of the following: --ip, --ns, --cname, --txt, --mx")
}
