package trace

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"fmt"
	"github.com/goccha/envar"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/rs/zerolog"
	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"net"
	"net/http"
	"strings"
)

const (
	TracingKey = "tracingContext"
	Uid        = "uid"
)

func init() {
	if Service == "" {
		Service = envar.String("GAE_SERVICE", "K_SERVICE")
	}
}

var Service string

func NewExporter(options *stackdriver.Options) (*stackdriver.Exporter, error) {
	// Create and register a OpenCensus Stackdriver Trace exporter.
	if options == nil {
		options = &stackdriver.Options{
			ProjectID: projectID,
		}
	} else if options.ProjectID == "" {
		options.ProjectID = projectID
	}
	exporter, err := stackdriver.NewExporter(*options)
	if err != nil {
		return nil, err
	}
	trace.RegisterExporter(exporter)
	return exporter, nil
}

func Wrap(handler http.Handler) http.Handler {
	if envar.Has("GAE_SERVICE", "K_SERVICE") { // GAE or CloudRun
		return &ochttp.Handler{
			Propagation: &propagation.HTTPFormat{},
			Handler:     handler,
		}
	}
	return &ochttp.Handler{
		Handler: handler,
	}
}

func NewClient(gcp bool) *http.Client {
	if gcp {
		return &http.Client{
			Transport: &ochttp.Transport{
				Propagation: &propagation.HTTPFormat{},
			},
		}
	}
	return &http.Client{
		Transport: &ochttp.Transport{},
	}
}

type Tracing interface {
	WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event
}

func With(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, TracingKey, &TracingContext{
		Path:      req.URL.Path,
		ClientIP:  ClientIP(req),
		RequestID: req.Header.Get(headers.RequestID),
		Service:   Service,
	})
}

type TracingContext struct {
	Path      string
	ClientIP  string
	RequestID string
	Service   string
}

func (tc *TracingContext) Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event {
	span := trace.FromContext(ctx)
	if span != nil {
		spanCtx := span.SpanContext()
		log = log.Str("trace_id", spanCtx.TraceID.String()).Str("span_id", spanCtx.SpanID.String()).
			Bool("sampled", spanCtx.IsSampled())
	}
	if tc.Service != "" {
		log = log.Dict("serviceContext", zerolog.Dict().Str("service", tc.Service))
	}
	return log.Str("client_ip", tc.ClientIP).
		Str("request_id", tc.RequestID)
}

func ClientIP(req *http.Request) string {
	clientIP := req.Header.Get(headers.XForwardedFor)
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(req.Header.Get(headers.XRealIp))
	}
	if clientIP != "" {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr)); err != nil && ip != "" {
		return ip
	}
	return req.Header.Get(headers.XEnvoyExternalAddress)
}

func (tc *TracingContext) WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	if projectID != "" {
		span := trace.FromContext(ctx)
		if span != nil {
			spanCtx := span.SpanContext()
			event = event.Str("logging.googleapis.com/trace", fmt.Sprintf("project/%s/traces/%s", projectID, spanCtx.TraceID.String())).
				Str("logging.googleapis.com/spanId", spanCtx.SpanID.String())
		}
	}
	if tc.RequestID != "" {
		event = event.Str("request_id", tc.RequestID)
	}
	if v := ctx.Value(Uid); v != nil {
		if str, ok := v.(string); ok {
			event = event.Str("uid", str)
		}
	}
	return event
}

var projectID = envar.String("GCP_PROJECT", "GOOGLE_CLOUD_PROJECT")

func SetHeader(ctx context.Context, req *http.Request) {
	value := ctx.Value(TracingKey)
	if value == nil {
		return
	}
	tracing := value.(*TracingContext)
	if tracing.ClientIP != "" {
		req.Header.Set(headers.XRealIp, tracing.ClientIP)
	}
	if tracing.RequestID != "" {
		req.Header.Set(headers.RequestID, tracing.RequestID)
	}
}
