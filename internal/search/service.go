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

// Service handles search operations with caching
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
	CachedAt   *time.Time              `json:"cached_at,omitempty"`
	SearchedAt time.Time               `json:"searched_at"`
}

// NewService creates a new search service
func NewService(aggregator *external.Aggregator, cache cache.Cache) *Service {
	return &Service{
		aggregator: aggregator,
		cache:      cache,
		cacheTTL:   24 * time.Hour, // Cache search results for 1 day
	}
}

// Search performs a search with caching
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

// getFromCache retrieves search results from cache
func (s *Service) getFromCache(cacheKey string) (*SearchResponse, error) {
	resultJSON, err := s.cache.GetString(cacheKey)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal([]byte(resultJSON), &response); err != nil {
		return nil, err
	}

	// Set the cached timestamp
	now := time.Now()
	response.CachedAt = &now

	return &response, nil
}

// cacheResults stores search results in cache
func (s *Service) cacheResults(cacheKey string, response *SearchResponse) error {
	resultJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return s.cache.SetString(cacheKey, string(resultJSON), s.cacheTTL)
}

// generateCacheKey creates a cache key for the search
func (s *Service) generateCacheKey(query, source string, maxResults int) string {
	if source == "" {
		source = "all"
	}
	return fmt.Sprintf("search:%s:%s:%d", source, query, maxResults)
}

// normalizeQuery normalizes a search query for consistent caching
func normalizeQuery(query string) string {
	query = strings.ToLower(query)

	query = strings.TrimSpace(query)

	query = strings.Join(strings.Fields(query), " ")

	return query
}

// InvalidateCache removes cached results for a specific query
func (s *Service) InvalidateCache(query, source string, maxResults int) error {
	if s.cache == nil {
		return nil
	}

	normalizedQuery := normalizeQuery(query)
	cacheKey := s.generateCacheKey(normalizedQuery, source, maxResults)

	return s.cache.Del(cacheKey)
}

// ClearAllCache clears all search cache (useful for maintenance)
func (s *Service) ClearAllCache(ctx context.Context) error {
	//TODO: implement properly cache purging
	return nil
}

// GetCacheStats returns cache statistics for a query
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
