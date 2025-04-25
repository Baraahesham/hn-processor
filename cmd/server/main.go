package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Baraahesham/hn-processor/internal/config"
	"github.com/Baraahesham/hn-processor/internal/db"
	hnprocessor "github.com/Baraahesham/hn-processor/internal/hn_processor"
	"github.com/Baraahesham/hn-processor/internal/nats"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func main() {
	setupEnv()
	logger := zerolog.New(os.Stdout).With().Logger()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	//intialize DB client
	DbClient, err := db.NewClient(db.NewDBClientParams{
		Logger: &logger,
		DbUrl:  viper.GetString(config.DBUrl),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize DB client")
		return
	}
	natsClient, err := nats.New(nats.NewNatsClientParams{
		Logger:  &logger,
		NatsUrl: viper.GetString(config.NatsUrl),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to NATS")
		return
	}
	// Initialize the HN processor client
	hnProcessor := hnprocessor.NewHnProcessor(hnprocessor.NewHnProcessorParams{
		Ctx:        ctx,
		Logger:     &logger,
		DbClient:   DbClient,
		NatsClient: natsClient,
		Brands:     config.Brands,
	})
	// Start the hnProcessor listener
	go func() {
		hnProcessor.Init()
	}()
	logger.Info().Msg("hn-processor is running.")
	<-ctx.Done()
	logger.Warn().Msg("Interrupt received. Cleaning up...")
	// Gracefully shutdown clients
	natsClient.Close()
	if err := DbClient.Close(); err != nil {
		logger.Error().Err(err).Msg("Error closing DB")
	}
}
func setupEnv() {
	viper.SetDefault(config.Port, "8080")
	viper.SetDefault(config.DBUrl, "postgres://hnuser:hnpass@localhost:5432/hackernews?sslmode=disable")
	viper.SetDefault(config.NatsUrl, "nats://localhost:4222")
	viper.SetDefault(config.NatsTopStoriesSubject, "hnfetcher.topstories")
	viper.SetDefault(config.MaxWorkers, 10)
	viper.SetDefault(config.MaxCapacity, 100)

}
