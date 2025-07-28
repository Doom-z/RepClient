package app

import (
	"github.com/Doom-z/RepClient/cmd/app/args"
	"github.com/Doom-z/RepClient/cmd/app/cfg"
	"github.com/Doom-z/RepClient/internal/run"
	"github.com/Doom-z/RepClient/pkg/logger"
)

func Handle(args args.Args, conf cfg.Conf) {
	run, err := run.NewRun(args, conf)
	if err != nil {
		logger.Fatal(err)
	}
	run.Start()
}
