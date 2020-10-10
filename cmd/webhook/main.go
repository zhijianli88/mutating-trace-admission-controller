package main

import (
	"context"
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mutating-trace-admission-controller/pkg/config"
	"mutating-trace-admission-controller/pkg/server"

	"github.com/golang/glog"
)

func main() {
	// read configuration location from command line arg
	var configPath string
	flag.StringVar(&configPath, "configPath", "", "Path that points to the YAML configuration for this webhook.")
	flag.Parse()
	// parse configuration
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		glog.Errorf("parse configuration failed: %v", err)
		return
	}
	// load certificates
	pair, err := cfg.LoadX509KeyPair()
	if err != nil {
		glog.Errorf("load X509 key pair failed: %v", err)
		return
	}
	// config webhook server
	whsvr := &server.WebhookServer{
		Server: &http.Server{
			Addr:      ":443",
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.Serve)
	whsvr.Server.Handler = mux
	// begin webhook server
	go func() {
		err := whsvr.Server.ListenAndServeTLS("", "")
		if err != nil {
			glog.Errorf("listen and serve webhook server failed: %v", err)
			return
		}
	}()
	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	// shutdown webhook server
	whsvr.Server.Shutdown(context.Background())
}
