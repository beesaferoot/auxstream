package main

import (
	"auxstream/config"
	"auxstream/internal/cache"
	"auxstream/internal/indexer"
	"auxstream/internal/logger"
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	intervalHours := flag.Int("interval", 24, "Indexing interval in hours")
	configPath := flag.String("config", ".", "Path to config directory")
	runOnce := flag.Bool("once", false, "Run indexing once and exit")
	flag.Parse()

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	environment := "production"
	if conf.GinMode == "debug" || conf.GinMode == "development" {
		environment = "development"
	}
	if err := logger.InitLogger(environment); err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	logger.Info("Starting indexer worker",
		zap.Int("interval_hours", *intervalHours),
		zap.Bool("run_once", *runOnce),
	)

	redisCache := cache.NewRedis(&redis.Options{
		Addr: conf.RedisAddr,
	})

	if _, err := redisCache.Exists(context.Background(), "test_connection"); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	logger.Info("Connected to Redis cache",
		zap.String("address", conf.RedisAddr),
	)

	indexingService := indexer.NewIndexingService(redisCache)

	interval := time.Duration(*intervalHours) * time.Hour
	worker := indexer.NewIndexerWorker(indexingService, interval)

	if err := worker.LoadPopularTracks(); err != nil {
		logger.Fatal("Failed to load external sources", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if *runOnce {
		logger.Info("Running indexing job once")
		worker.RunIndexingOnce(ctx)
		logger.Info("Indexing job completed, exiting")
		return
	}

	go worker.Start(ctx)
	logger.Info("Indexer worker started, waiting for signals...")

	<-sigChan
	logger.Info("Received shutdown signal, stopping worker...")
	cancel()

	// Give the worker some time to finish
	time.Sleep(2 * time.Second)
	logger.Info("Indexer worker stopped")
}
