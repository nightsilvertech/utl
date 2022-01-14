package preparation

import (
	oczipkin "contrib.go.opencensus.io/exporter/zipkin"
	"github.com/afex/hystrix-go/hystrix"
	gokitlog "github.com/go-kit/kit/log"
	"github.com/nightsilvertech/utl/console"
	"github.com/openzipkin/zipkin-go"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

type Data struct {
	LoggingFilePath            string
	Debug                      bool
	ServiceName                string
	TracerUrl                  string
	ZipkinEndpointPort         string
	FractionProbabilitySampler float64
	CircuitBreakerTimeout      int
}

func (prep Data) CircuitBreaker() {
	hystrix.ConfigureCommand(prep.ServiceName, hystrix.CommandConfig{Timeout: prep.CircuitBreakerTimeout})
}

func (prep Data) Logger() gokitlog.Logger {
	return console.CreateStdGoKitLog(prep.ServiceName, prep.Debug, prep.LoggingFilePath)
}

func (prep Data) Tracer() trace.Tracer {
	reporter := httpreporter.NewReporter(prep.TracerUrl)
	localEndpoint, _ := zipkin.NewEndpoint(prep.ServiceName, prep.ZipkinEndpointPort)
	exporter := oczipkin.NewExporter(reporter, localEndpoint)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(prep.FractionProbabilitySampler)})
	trace.RegisterExporter(exporter)
	return trace.DefaultTracer
}
