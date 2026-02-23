package obs

import (
	"github.com/nimaeskandary/go-realworld/pkg/observability/internal"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

func NewSlogLoggerModule() fx.Option {
	return util.NewFxModule[obs_types.Logger](
		"slog_logger",
		internal.NewSlogLogger,
	)
}

var NewSlogLogger = internal.NewSlogLogger
