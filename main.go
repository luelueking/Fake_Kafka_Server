package main

import (
	"context"
	"fmt"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"os"
	"os/signal"
	"syscall"

	"Fake_Kafka_Server/server"
	"Fake_Kafka_Server/server/config"
)

var (
	brokerCfg = config.DefaultConfig()
)

func init() {
	brokerCfg.Addr = "0.0.0.0:9092"
	brokerCfg.DataDir = "/tmp/jocko"
	brokerCfg.RaftAddr = "127.0.0.1:9093"

}

func main() {
	fmt.Println(brokerCfg)
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	tracer, closer, _ := cfg.New(
		"jocko",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)

	srv := server.NewServer(brokerCfg, nil, nil, tracer, closer.Close)
	if err := srv.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %v\n", err)
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signals:
		fmt.Printf("Received signal %s, shutting down...\n", sig)
	}

	if err := srv.Shutdown(); err != nil {
		fmt.Fprintf(os.Stderr, "error shutting down server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server shutdown gracefully.")
}
