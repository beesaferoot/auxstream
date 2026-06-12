package search

import (
	"auxstream/internal/cache"
	"auxstream/internal/external"
	"auxstream/internal/logger"
	"auxstream/internal/metrics"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Service fronts the external Aggregator with a read-through cache; identical
// queries within cacheTTL are served from cache rather than re-hitting sources.
type Service struct {
	aggregator *external.Aggregator
	cache      cache.Cache
	cacheTTL   time.Duration
}

// SearchRequest represents a search query
type SearchRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
	Source     string `json:"source,omitempty"` // Optional: "local", "youtube", or empty for all
}

// SearchResponse represents the search results with metadata
type SearchResponse struct {
	Query      string                  `json:"query"`
	Results    []external.SearchResult `json:"results"`
	TotalCount int                     `json:"total_count"`
	Source     string                  `json:"source"`
	CachedAt   *time.Time              `json:"cached_at,omitempty"` // set only when served from cache; nil on a fresh search
	SearchedAt time.Time               `json:"searched_at"`
}

// NewService wires the aggregator and cache together. Pass a nil cache to
// disable caching entirely; the service then queries sources on every call.
func NewService(aggregator *external.Aggregator, cache cache.Cache) *Service {
	return &Service{
		aggregator: aggregator,
		cache:      cache,
		// Catalogs change slowly, so a day-long TTL trades freshness for far
		// fewer external API calls (which are rate-limited and/or billed).
		cacheTTL: 24 * time.Hour,
	}
}

// Search returns results for req, serving from cache on a hit and otherwise
// querying the aggregator and caching the response. An empty req.Source fans
// out to all sources; a cache miss is not an error. The returned response has
// CachedAt set only when it came from cache.
func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()
	normalizedQuery := normalizeQuery(req.Query)

	if normalizedQuery == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if req.MaxResults <= 0 {
		req.MaxResults = 20
	}
	if req.MaxResults > 50 {
		req.MaxResults = 50
	}

	source := req.Source
	if source == "" {
		source = "all"
	}

	cacheKey := s.generateCacheKey(normalizedQuery, req.Source, req.MaxResults)

	if s.cache != nil {
		cachedResp, err := s.getFromCache(cacheKey)
		if err == nil && cachedResp != nil {
			metrics.RecordCacheHit("search")
			logger.Debug("Search cache hit",
				zap.String("query", normalizedQuery),
				zap.String("source", source),
			)
			metrics.RecordSearchRequest(source, "success", time.Since(startTime).Seconds())
			return cachedResp, nil
		}
		metrics.RecordCacheMiss("search")
	}

	var results []external.SearchResult
	var err error

	if req.Source != "" {
		results, err = s.aggregator.SearchBySource(ctx, normalizedQuery, req.Source, req.MaxResults)
	} else {
		results, err = s.aggregator.Search(ctx, normalizedQuery, req.MaxResults)
	}

	if err != nil {
		logger.Error("Search failed",
			zap.String("query", normalizedQuery),
			zap.String("source", source),
			zap.Error(err),
		)
		metrics.RecordSearchRequest(source, "error", time.Since(startTime).Seconds())
		return nil, fmt.Errorf("search failed: %w", err)
	}

	response := &SearchResponse{
		Query:      normalizedQuery,
		Results:    results,
		TotalCount: len(results),
		Source:     req.Source,
		SearchedAt: time.Now(),
	}

	if s.cache != nil {
		if err := s.cacheResults(cacheKey, response); err != nil {
			logger.Warn("Failed to cache search results",
				zap.String("query", normalizedQuery),
				zap.Error(err),
			)
		}
	}

	logger.Info("Search completed",
		zap.String("query", normalizedQuery),
		zap.String("source", source),
		zap.Int("result_count", len(results)),
		zap.Duration("duration", time.Since(startTime)),
	)

	metrics.RecordSearchRequest(source, "success", time.Since(startTime).Seconds())

	return response, nil
}

// getFromCache returns the cached response, stamping CachedAt so callers can
// distinguish a cache hit from a fresh search. A miss surfaces as an error.
func (s *Service) getFromCache(cacheKey string) (*SearchResponse, error) {
	resultJSON, err := s.cache.GetString(cacheKey)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal([]byte(resultJSON), &response); err != nil {
		return nil, err
	}

	now := time.Now()
	response.CachedAt = &now

	return &response, nil
}

func (s *Service) cacheResults(cacheKey string, response *SearchResponse) error {
	resultJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return s.cache.SetString(cacheKey, string(resultJSON), s.cacheTTL)
}

// generateCacheKey builds the key from source, query, and maxResults. Empty
// source maps to "all" so it matches the fan-out path, and maxResults is part
// of the key because a request for more results is not satisfiable from a
// smaller cached set.
func (s *Service) generateCacheKey(query, source string, maxResults int) string {
	if source == "" {
		source = "all"
	}
	return fmt.Sprintf("search:%s:%s:%d", source, query, maxResults)
}

// normalizeQuery lowercases and collapses whitespace so that queries differing
// only in case or spacing share one cache entry.
func normalizeQuery(query string) string {
	query = strings.ToLower(query)

	query = strings.TrimSpace(query)

	query = strings.Join(strings.Fields(query), " ")

	return query
}

// InvalidateCache drops the entry for exactly this query/source/maxResults
// triple; other variants of the same query remain cached. No-op without a cache.
func (s *Service) InvalidateCache(query, source string, maxResults int) error {
	if s.cache == nil {
		return nil
	}

	normalizedQuery := normalizeQuery(query)
	cacheKey := s.generateCacheKey(normalizedQuery, source, maxResults)

	return s.cache.Del(cacheKey)
}

// ClearAllCache clears all search cache (useful for maintenance).
//
// Not yet implemented: it returns an error rather than nil so callers don't
// mistake a no-op for a successful purge.
func (s *Service) ClearAllCache(ctx context.Context) error {
	return fmt.Errorf("search cache purging not implemented")
}

// GetCacheStats reports whether this query is cached and, if so, its remaining
// TTL. Without a cache configured it reports (false, 0, nil).
func (s *Service) GetCacheStats(ctx context.Context, query, source string, maxResults int) (bool, time.Duration, error) {
	if s.cache == nil {
		return false, 0, nil
	}

	normalizedQuery := normalizeQuery(query)
	cacheKey := s.generateCacheKey(normalizedQuery, source, maxResults)

	exists, err := s.cache.Exists(ctx, cacheKey)
	if err != nil {
		return false, 0, err
	}

	if !exists {
		return false, 0, nil
	}

	ttl, err := s.cache.TTL(ctx, cacheKey)
	if err != nil {
		return true, 0, err
	}

	return true, ttl, nil
}
