package log

import (
	"github.com/xuperchain/xuperbench/log/colorlog"
)

var (
	logger = colorlog.New(nil, nil, new(colorlog.ColouredFormatter))

	DEBUG   = logger[colorlog.DEBUG]
	INFO    = logger[colorlog.INFO]
	WARNING = logger[colorlog.WARNING]
	WARN    = WARNING
	ERROR   = logger[colorlog.ERROR]
)
