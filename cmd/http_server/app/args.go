package app

import (
	"flag"
	"log"

	"github.com/nimaeskandary/go-realworld/pkg/util"
)

type Args struct {
	ConfigPath string `validate:"required"`
}

func ParseArgs() Args {
	args := Args{}
	flag.StringVar(&args.ConfigPath, "config-path", "", "path to the config file")
	flag.Parse()

	validator := util.NewValidator()
	err := validator.Struct(args)
	if err != nil {
		flag.Usage()
		log.Fatalf("invalid arguments: %v", err)
	}

	return args
}
