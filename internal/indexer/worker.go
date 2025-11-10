package indexer

import (
	"auxstream/internal/logger"
	"auxstream/internal/metrics"
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type IndexerWorker struct {
	service  *IndexingService
	interval time.Duration

	mu       sync.RWMutex
	urlLists map[string][]string
}

type ExternalSourcesConfig struct {
	Sources map[string][]string `yaml:"sources"`
}

// NewIndexerWorker creates a new indexer with a given indexing interval.
func NewIndexerWorker(service *IndexingService, interval time.Duration) *IndexerWorker {
	return &IndexerWorker{
		service:  service,
		interval: interval,
		urlLists: make(map[string][]string),
	}
}

// AddURLList safely registers a source with its URLs.
func (w *IndexerWorker) AddURLList(source string, urls []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.urlLists[source] = urls
}

// Start runs the indexing process until the context is cancelled.
func (w *IndexerWorker) Start(ctx context.Context) {
	logger.Info("Indexer worker started",
		zap.Duration("interval", w.interval),
	)

	// Initial indexing before the first tick
	w.runIndexing(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Optional jitter to prevent sync workloads
			time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
			w.runIndexing(ctx)
		case <-ctx.Done():
			logger.Warn("Indexer worker stopping due to context cancel")
			return
		}
	}
}

// RunIndexingOnce runs the indexing process once and returns.
func (w *IndexerWorker) RunIndexingOnce(ctx context.Context) {
	w.runIndexing(ctx)
}

// runIndexing concurrently processes all URL lists.
func (w *IndexerWorker) runIndexing(ctx context.Context) {
	start := time.Now()

	w.mu.RLock()
	sources := make(map[string][]string, len(w.urlLists))
	for k, v := range w.urlLists {
		sources[k] = append([]string(nil), v...) // copy slice
	}
	w.mu.RUnlock()

	logger.Info("Starting indexing job", zap.Int("sources", len(sources)))

	var (
		wg           sync.WaitGroup
		totalSuccess int64
		totalFail    int64
		mu           sync.Mutex
	)

	for source, urls := range sources {
		if len(urls) == 0 {
			continue
		}

		wg.Add(1)
		go func(source string, urls []string) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					logger.Error("Recovered from panic in source",
						zap.String("source", source),
						zap.Any("error", r),
					)
				}
			}()

			logger.Info("Indexing source",
				zap.String("source", source),
				zap.Int("url_count", len(urls)),
			)

			success, fail := w.service.IndexBatch(ctx, urls)

			mu.Lock()
			totalSuccess += int64(success)
			totalFail += int64(fail)
			mu.Unlock()

			metrics.RecordIndexerJobSource(source, float64(success), float64(fail))
			logger.Info("Source indexing done",
				zap.String("source", source),
				zap.Int("succeeded", success),
				zap.Int("failed", fail),
			)
		}(source, urls)
	}

	wg.Wait()
	duration := time.Since(start)
	metrics.RecordIndexerJob(duration.Seconds(), int(totalSuccess), int(totalFail))

	logger.Info("Indexing job completed",
		zap.Int64("total_succeeded", totalSuccess),
		zap.Int64("total_failed", totalFail),
		zap.Duration("duration", duration),
	)
}

// LoadPopularTracks loads external sources from the YAML config file.
func (w *IndexerWorker) LoadPopularTracks() error {
	configPath := "config/ext_sources.yaml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error("Failed to read external sources config",
			zap.String("path", configPath),
			zap.Error(err),
		)
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config ExternalSourcesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		logger.Error("Failed to parse external sources config",
			zap.String("path", configPath),
			zap.Error(err),
		)
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if len(config.Sources) == 0 {
		logger.Warn("No sources found in config file",
			zap.String("path", configPath),
		)
		return fmt.Errorf("no sources found in config")
	}

	for source, urls := range config.Sources {
		if len(urls) > 0 {
			w.AddURLList(source, urls)
			logger.Info("Loaded external source",
				zap.String("source", source),
				zap.Int("url_count", len(urls)),
			)
		}
	}

	logger.Info("Successfully loaded external sources",
		zap.Int("source_count", len(config.Sources)),
	)

	return nil
}
