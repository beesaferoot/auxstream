package indexer

import (
	"auxstream/internal/cache"
	"auxstream/internal/logger"
	"auxstream/internal/metrics"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type IndexingService struct {
	registry *ScraperRegistry
	cache    cache.Cache
}

func NewIndexingService(cache cache.Cache) *IndexingService {
	return &IndexingService{
		registry: NewScraperRegistry(),
		cache:    cache,
	}
}

func (s *IndexingService) IndexURL(ctx context.Context, url string) (*ScrapedMetadata, error) {
	cacheKey := fmt.Sprintf("indexed_track:%s", url)
	var cachedMetadata ScrapedMetadata
	if err := s.cache.Get(cacheKey, &cachedMetadata); err == nil {
		metrics.RecordCacheHit("indexed_track")
		return &cachedMetadata, nil
	}

	metrics.RecordCacheMiss("indexed_track")

	metadata, err := s.registry.ScrapeURL(ctx, url)
	if err != nil {
		logger.Error("Failed to scrape URL",
			zap.String("url", url),
			zap.Error(err),
		)
		metrics.IndexerTracksIndexed.WithLabelValues(DetectSourceFromURL(url), "failed").Inc()
		return nil, fmt.Errorf("failed to scrape URL: %w", err)
	}

	_ = s.cache.Set(cacheKey, metadata, 24*time.Hour)

	s.cacheInSearchIndex(metadata)

	logger.Debug("Track indexed",
		zap.String("artist", metadata.Artist),
		zap.String("title", metadata.Title),
		zap.String("source", metadata.Source),
		zap.String("url", url),
	)

	metrics.IndexerTracksIndexed.WithLabelValues(metadata.Source, "success").Inc()

	return metadata, nil
}

func (s *IndexingService) IndexBatch(ctx context.Context, urls []string) (int, int) {
	successCount := 0
	failCount := 0

	for _, url := range urls {
		if _, err := s.IndexURL(ctx, url); err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	return successCount, failCount
}

func (s *IndexingService) cacheInSearchIndex(metadata *ScrapedMetadata) {
	searchKey := fmt.Sprintf("indexed_search_all:%s", metadata.Source)

	var existingTracks []*ScrapedMetadata
	_ = s.cache.Get(searchKey, &existingTracks)

	existingTracks = append([]*ScrapedMetadata{metadata}, existingTracks...)

	if len(existingTracks) > 1000 {
		existingTracks = existingTracks[:1000]
	}

	_ = s.cache.Set(searchKey, &existingTracks, 24*time.Hour)
}

func (s *IndexingService) GetIndexedTracks(source string, limit int) ([]*ScrapedMetadata, error) {
	searchKey := fmt.Sprintf("indexed_search_all:%s", source)

	var tracks []*ScrapedMetadata
	if err := s.cache.Get(searchKey, &tracks); err != nil {
		return []*ScrapedMetadata{}, nil
	}

	if len(tracks) > limit {
		tracks = tracks[:limit]
	}

	return tracks, nil
}

func (s *IndexingService) SearchIndexedTracks(source, query string, limit int) []*ScrapedMetadata {
	searchKey := fmt.Sprintf("indexed_search_all:%s", source)

	var allTracks []*ScrapedMetadata
	if err := s.cache.Get(searchKey, &allTracks); err != nil {
		return []*ScrapedMetadata{}
	}

	results := make([]*ScrapedMetadata, 0)
	queryLower := toLower(query)

	for _, track := range allTracks {
		if contains(toLower(track.Title), queryLower) || contains(toLower(track.Artist), queryLower) {
			results = append(results, track)
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
