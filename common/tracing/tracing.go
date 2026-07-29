package tracing

import (
	"context"
	"fmt"

	"github.com/MiLab-Bit/OpenFastToken/common"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer() (*trace.TracerProvider, error) {
	opts := []trace.TracerProviderOption{
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("FastToken"),
		)),
	}

	// Try to connect to Jaeger (optional)
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		common.SysError(fmt.Sprintf("[WARN] Failed to create Jaeger exporter: %v", err))
		common.SysError("[WARN] OpenTelemetry traces will not be exported (running without Jaeger)")
	} else {
		opts = append(opts, trace.WithBatcher(exporter))
	}

	tp := trace.NewTracerProvider(opts...)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func ShutdownTracer(tp *trace.TracerProvider) {
	if err := tp.Shutdown(context.Background()); err != nil {
		common.SysError(fmt.Sprintf("Error shutting down tracer: %v", err))
	}
}
