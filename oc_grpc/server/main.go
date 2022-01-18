package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/bvwells/observability-examples/oc_grpc/helloworld"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

/*
  For information on Prometheus see https://prometheus.io/

  Run and view metrics on http://localhost:8888/metrics
*/

const (
	address = "127.0.0.1:50000"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
	message := "Hello " + in.Name
	fmt.Println(message)
	return &helloworld.HelloResponse{Message: message}, nil
}

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

	err = view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		log.Fatalf("failed to register views: %v", err)
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	helloworld.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
