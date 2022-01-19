package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

/*
  For information on Prometheus see https://prometheus.io/

  Run and view metrics on http://localhost:8888/metrics
*/

func main() {
	pe, err := prometheus.New(prometheus.Config{}, controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
	))
	if err != nil {
		log.Fatalf("failed to create Prometheus exporter: %v", err)
	}
	global.SetMeterProvider(pe.MeterProvider())

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		fmt.Println("running metrics on :8888")
		if err := http.ListenAndServe(":8888", mux); err != nil {
			log.Fatalf("failed to run Prometheus /metrics endpoint: %v", err)
		}
	}()
	ctx := context.Background()

	meter := global.Meter("my-meter")
	histogram := metric.Must(meter).NewFloat64Histogram("myDuration", metric.WithDescription("The distribution of the duration"))
	labels := []attribute.KeyValue{attribute.String("myTag", "my-tag-value")}

	for {
		fmt.Println("exporting...")
		startTime := time.Now()
		time.Sleep(time.Second)

		histogram.Record(ctx, float64(time.Since(startTime)), labels...)
	}
}

/*
# HELP myDuration
# TYPE myDuration histogram
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="0.005"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="0.01"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="0.025"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="0.05"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="0.1"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="0.25"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="0.5"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="1"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="2.5"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="5"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="10"} 0
myDuration_bucket{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0",le="+Inf"} 15
myDuration_sum{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0"} 1.5010590908e+10
myDuration_count{myTag="my-tag-value",service_name="unknown_service:main",telemetry_sdk_language="go",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.3.0"} 15
*/
