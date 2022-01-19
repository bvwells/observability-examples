package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

/*
Start up Jaegar.

docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.30

  View traces on Jaeger UI endpoint http://localhost:16686
*/

func main() {
	// Port details: https://www.jaegertracing.io/docs/getting-started/
	collectorEndpointURI := "http://localhost:14268/api/traces"

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(collectorEndpointURI)))
	if err != nil {
		log.Fatalf("failed to create the Jaeger exporter: %v", err)
	}
	je := tracesdk.NewTracerProvider(
		// Always be sure to batch in production. Syncer is used here for demo
		// purpose.
		tracesdk.WithSyncer(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("demo"),
			attribute.String("environment", "demo"),
			attribute.Int64("ID", 1),
		)),
	)
	otel.SetTracerProvider(je)

	ctx, span := otel.Tracer("").Start(context.Background(), "Main")
	defer span.End()
	something(ctx)
}

func something(ctx context.Context) {
	_, span := otel.Tracer("").Start(ctx, "Something")
	defer span.End()
}
