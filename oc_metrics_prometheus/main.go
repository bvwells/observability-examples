package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
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

	myDurationMeasure := stats.Float64("myDuration", "Duration of something in milli seconds", stats.UnitMilliseconds)

	myTag, err := tag.NewKey("myTag")
	if err != nil {
		log.Fatalf("failed to create new key: %v", err)
	}

	v := &view.View{
		Name:        "myView",
		Measure:     myDurationMeasure,
		Description: "The distribution of the duration",
		TagKeys:     []tag.Key{myTag},
		Aggregation: view.Distribution(100, 500, 1000, 2000, 4000, 8000, 16000),
	}

	err = view.Register(v)
	if err != nil {
		log.Fatalf("failed to register view: %v", err)
	}
	view.RegisterExporter(pe)

	ctx := context.Background()
	ctx, err = tag.New(ctx, tag.Insert(myTag, "my-tag-value"))
	if err != nil {
		log.Fatalf("failed to tag context: %v", err)
	}

	for {
		fmt.Println("exporting...")
		startTime := time.Now()
		time.Sleep(time.Duration(rand.Float64()) * time.Millisecond)
		stats.Record(ctx, myDurationMeasure.M(float64(time.Since(startTime).Nanoseconds())/float64(time.Millisecond)))
	}
}
