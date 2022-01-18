package main

import (
	"fmt"
	"log"
	"net/http"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

/*
  For information on Prometheus see https://prometheus.io/

  Run and view metrics on http://localhost:8888/metrics
*/

func main() {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "demo",
	})
	if err != nil {
		log.Fatalf("failed to create Prometheus exporter: %v", err)
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		fmt.Println("running metrics on :8888")
		if err := http.ListenAndServe(":8888", mux); err != nil {
			log.Fatalf("failed to run Prometheus /metrics endpoint: %v", err)
		}
	}()

	// Port details: https://www.jaegertracing.io/docs/getting-started/
	agentEndpointURI := "localhost:6831"
	collectorEndpointURI := "http://localhost:14268/api/traces"

	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     agentEndpointURI,
		CollectorEndpoint: collectorEndpointURI,
		ServiceName:       "demo",
	})
	if err != nil {
		log.Fatalf("failed to create the Jaeger exporter: %v", err)
	}

	view.RegisterExporter(pe)
	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	err = view.Register(ochttp.DefaultServerViews...)
	if err != nil {
		log.Fatalf("failed to register views: %v", err)
	}

	http.Handle("/hello", ochttp.WithRouteTag(http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "hello world")
		}), "get.Hello"))

	log.Fatal(http.ListenAndServe(":8080", &ochttp.Handler{}))
}
