package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	bot, err := NewChineseBot(config)
	if err != nil {
		return fmt.Errorf("creating bot: %w", err)
	}
	defer bot.Stop()

	if err := bot.Start(); err != nil {
		return fmt.Errorf("starting bot: %w", err)
	}

	if err := bot.RegisterCommands(); err != nil {
		return fmt.Errorf("registering commands: %w", err)
	}

	log.Println("Daily Chinese Teacher bot is now running. Press CTRL-C to exit.")

	return waitForShutdown()
}

func waitForShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gracefully...")
	return nil
}
