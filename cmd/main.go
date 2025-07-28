package main

import (
	"github.com/Doom-z/RepClient/cmd/app"
	"github.com/Doom-z/RepClient/cmd/app/args"
	"github.com/Doom-z/RepClient/cmd/app/cfg"
	"github.com/Doom-z/RepClient/cmd/app/log"
	"github.com/alexflint/go-arg"
)

func main() {
	var args args.Args
	var conf cfg.Conf
	var defaultConf = cfg.GetDefaultConf()
	args = LoadArgsValid()
	conf = cfg.LoadConfValid(args.Config, defaultConf, "config.toml")
	log.InitLogger(conf.Log, args.Verbose)
	app.Handle(args, conf)
}

func LoadArgsValid() args.Args {
	var args args.Args
	arg.MustParse(&args)
	return args
}
