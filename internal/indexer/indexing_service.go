package indexer

import (
	"auxstream/internal/cache"
	"auxstream/internal/logger"
	"auxstream/internal/metrics"
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// IndexingService scrapes track URLs and caches both the per-URL metadata and a
// per-source search index built from it.
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

// IndexURL scrapes a single URL and caches the result, returning the cached
// copy on a hit without re-scraping. On success it also adds the track to the
// per-source search index so SearchIndexedTracks can find it.
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

// IndexBatch indexes urls sequentially and returns (succeeded, failed) counts;
// a failure on one URL does not stop the rest.
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

// cacheInSearchIndex prepends the track to its source's cached list so the
// newest tracks rank first, capping the list at 1000 to bound memory and
// dropping the oldest entries beyond that.
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

// GetIndexedTracks returns up to limit most-recently-indexed tracks for source.
// A cache miss yields an empty slice and nil error, not a failure.
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

// SearchIndexedTracks does a case-insensitive substring match on title and
// artist over the cached tracks for source, returning at most limit hits.
func (s *IndexingService) SearchIndexedTracks(source, query string, limit int) []*ScrapedMetadata {
	searchKey := fmt.Sprintf("indexed_search_all:%s", source)

	var allTracks []*ScrapedMetadata
	if err := s.cache.Get(searchKey, &allTracks); err != nil {
		return []*ScrapedMetadata{}
	}

	results := make([]*ScrapedMetadata, 0)
	queryLower := strings.ToLower(query)

	for _, track := range allTracks {
		if strings.Contains(strings.ToLower(track.Title), queryLower) || strings.Contains(strings.ToLower(track.Artist), queryLower) {
			results = append(results, track)
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

