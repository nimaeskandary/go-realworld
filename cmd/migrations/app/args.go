package app

import (
	"flag"
	"log"

	"github.com/nimaeskandary/go-realworld/pkg/util"
)

const (
	ActionApply      string = "apply"
	ActionApplyAll   string = "apply-all"
	ActionRollback   string = "rollback"
	ActionRollbackTo string = "rollback-to"
	ActionStatus     string = "status"
)

type Args struct {
	ConfigPath     string `validate:"required"`
	TargetDatabase string `validate:"required,oneof=realworld_app"`
	Action         string `validate:"required,oneof=apply apply-all rollback rollback-to status"`
	Version        *int64 `validate:"required_if=Action apply,required_if=Action rollback,required_if=Action rollback-to"`
}

func ParseArgs() Args {
	args := Args{}
	var version_ int64

	flag.StringVar(&args.ConfigPath, "config-path", "", "path to the config file")
	flag.StringVar(&args.TargetDatabase, "target-database", "", "target database for migrations")
	flag.StringVar(&args.Action, "action", "", "migration action to perform: apply, apply-all, rollback, rollback-to, status")
	flag.Int64Var(&version_, "version", 0, "target version for up-to and down-to actions")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "version" {
			args.Version = &version_
		}
	})

	validator := util.NewValidator()
	err := validator.Struct(args)
	if err != nil {
		flag.Usage()
		log.Fatalf("invalid arguments: %v", err)
	}

	return args
}
