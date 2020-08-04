package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"mutating-trace-admission-controller/pkg/config"
	"mutating-trace-admission-controller/pkg/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"go.opencensus.io/trace"
)

func main() {
	// read configuration location from command line arg
	var configPath string
	flag.StringVar(&configPath, "configPath", config.DefaultConfigPath, "Path that points to the YAML configuration for this webhook.")
	flag.Parse()

	// parse and validate configuration
	cfg := config.Config{}

	ok, err := config.ParseConfigFromPath(&cfg, configPath)
	if !ok {
		glog.Errorf("configuration parse failed with error: %v", err)
		return
	}

	ok, err = cfg.Validate()
	if !ok {
		glog.Errorf("configuration validation failed with error: %v", err)
		return
	}

	// configure global tracer
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(cfg.Trace.SampleRate)})

	// configure certificates
	pair, err := tls.LoadX509KeyPair("/etc/webhook/certs/cert.pem", "/etc/webhook/certs/key.pem")
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	whsvr := &server.WebhookServer{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", 443),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.Serve)
	whsvr.Server.Handler = mux

	// begin webhook server
	go func() {
		if err := whsvr.Server.ListenAndServeTLS("", ""); err != nil {
			glog.Fatalf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Errorf("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsvr.Server.Shutdown(context.Background())
}
