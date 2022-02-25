package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nginxinc/nginx-ns1-gslb/internal/agent"
)

var configFile = flag.String("config-file", "", "Path to the agent configuration file")

func main() {
	flag.Parse()

	if *configFile == "" {
		log.Fatalf("config-file must be specified")
	}

	globalConfig, err := agent.ParseConfig(configFile)
	if err != nil {
		log.Fatalf("error generating the config: %v", err)
	}

	a, err := agent.New(globalConfig)
	if err != nil {
		log.Fatalf("error creating the agent: %v", err)
	}

	go handleTermination()
	a.Run()
}

// handleTermination makes the agent exit on SIGTERM or SIGINT
func handleTermination() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	s := <-sigChan
	log.Printf("%v signal received, shutting down the agent", s)
	os.Exit(0)
}
