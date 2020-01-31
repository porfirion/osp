package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"

	"github.com/porfirion/osp/front"
	"github.com/porfirion/osp/processor"
)

type ospConfig struct {
	Host          string
	Port          string
	UnlabeledPath string
	LabeledPath   string
}

var logger *log.Logger = log.New(os.Stdout, "OSP: ", 0)

func main() {
	var config ospConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		logger.Fatal("error decoding config", err)
	}

	logger.Printf("config: %v", config)

	var p processor.Processor
	p, err := processor.NewImageProcessor(config.UnlabeledPath, config.LabeledPath)
	if err != nil {
		logger.Fatalf("error creating ImageProcessor %v\n", err)
	}

	srv, err := front.NewServer(config.Host, config.Port, config.UnlabeledPath, p)
	if err != nil {
		logger.Fatalf("error creating server: %v\n", err)
	}

	srv.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill)

	select {
	case <-interrupt:
	}

	logger.Printf("FINISHED")
}
