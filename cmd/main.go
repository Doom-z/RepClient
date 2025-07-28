package main

import (
	"github.com/Doom-z/RepClient/cmd/app"
	"github.com/Doom-z/RepClient/cmd/app/cfg"
	"github.com/Doom-z/RepClient/cmd/app/log"
	"github.com/alexflint/go-arg"
)

func main() {
	var args app.Args
	var conf cfg.Conf
	var defaultConf = cfg.GetDefaultConf()
	args = LoadArgsValid()
	conf = cfg.LoadConfValid(args.Config, defaultConf, "config.toml")
	log.InitLogger(conf.Log, args.Verbose)
	app.Run(args, conf)
}

func LoadArgsValid() app.Args {
	var args app.Args
	arg.MustParse(&args)
	return args
}
