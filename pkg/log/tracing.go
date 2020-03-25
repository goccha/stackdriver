package log

import (
	"context"
	"github.com/goccha/envar"
	"github.com/goccha/stackdriver/pkg/trace"
	"github.com/rs/zerolog"
)

const (
	Client      = "client"
	ErrorReport = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
)

func init() {
	if Service == "" {
		Service = envar.String("GAE_SERVICE", "K_SERVICE")
	}
}

var Service string

func WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	value := ctx.Value(trace.TracingKey)
	if value != nil {
		tracing := value.(trace.Tracing)
		event = tracing.WithTrace(ctx, event)
	}
	return event
}

func Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event {
	value := ctx.Value(trace.TracingKey)
	if value == nil {
		return log
	}
	tc := value.(*trace.TracingContext)
	return tc.Dump(ctx, log)
}
