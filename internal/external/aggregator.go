package external

import (
	"auxstream/internal/db"
	"auxstream/internal/logger"
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// SearchResult represents a unified search result from any source
type SearchResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Duration    int    `json:"duration"` // in seconds
	Thumbnail   string `json:"thumbnail"`
	Source      string `json:"source"` // "local", "youtube", "soundcloud"
	ExternalID  string `json:"external_id,omitempty"`
	StreamURL   string `json:"stream_url"`
	Description string `json:"description,omitempty"`
}

// Aggregator combines search results from multiple sources
type Aggregator struct {
	youtubeClient    *YouTubeClient
	soundcloudClient *SoundCloudClient
	trackRepo        db.TrackRepo
}

// NewAggregator creates a new search aggregator
func NewAggregator(youtubeClient *YouTubeClient, soundcloudClient *SoundCloudClient, trackRepo db.TrackRepo) *Aggregator {
	return &Aggregator{
		youtubeClient:    youtubeClient,
		soundcloudClient: soundcloudClient,
		trackRepo:        trackRepo,
	}
}

// Search performs a unified search across all available sources
func (a *Aggregator) Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []SearchResult
	)

	availableSources := 1 // the local database is always searchable
	if a.youtubeClient != nil && a.youtubeClient.apiKey != "" {
		availableSources++
	}
	if a.soundcloudClient != nil && a.soundcloudClient.clientID != "" {
		availableSources++
	}

	resultsPerSource := maxResults / availableSources
	if resultsPerSource < 5 {
		resultsPerSource = 5
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		localResults, err := a.searchLocal(ctx, query, resultsPerSource)
		if err != nil {
			logger.Error("Error searching local database", zap.Error(err))
			return
		}
		mu.Lock()
		results = append(results, localResults...)
		mu.Unlock()
	}()

	if a.youtubeClient != nil && a.youtubeClient.apiKey != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ytResults, err := a.searchYouTube(ctx, query, resultsPerSource)
			if err != nil {
				logger.Error("Error searching YouTube", zap.Error(err))
				return
			}
			mu.Lock()
			results = append(results, ytResults...)
			mu.Unlock()
		}()
	}

	if a.soundcloudClient != nil && a.soundcloudClient.clientID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scResults, err := a.searchSoundCloud(ctx, query, resultsPerSource)
			if err != nil {
				logger.Error("Error searching SoundCloud", zap.Error(err))
				return
			}
			mu.Lock()
			results = append(results, scResults...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

func (a *Aggregator) searchLocal(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	tracks, err := a.trackRepo.GetTrackByTitle(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search local tracks: %w", err)
	}

	artistTracks, err := a.trackRepo.GetTrackByArtist(ctx, query)
	if err != nil {
		logger.Error("Error searching by artist", zap.Error(err))
	} else {
		tracks = append(tracks, artistTracks...)
	}

	seen := make(map[string]bool)
	var uniqueTracks []*db.Track
	for _, track := range tracks {
		if !seen[track.ID.String()] {
			seen[track.ID.String()] = true
			uniqueTracks = append(uniqueTracks, track)
		}
	}

	var results []SearchResult
	for i, track := range uniqueTracks {
		if i >= maxResults {
			break
		}

		results = append(results, SearchResult{
			ID:        track.ID.String(),
			Title:     track.Title,
			Artist:    track.Artist.Name,
			Duration:  track.Duration,
			Thumbnail: track.Thumbnail,
			Source:    "local",
			StreamURL: fmt.Sprintf("/api/v1/serve/%s", track.File),
		})
	}

	return results, nil
}

// searchYouTube searches YouTube for tracks
func (a *Aggregator) searchYouTube(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	ytResults, err := a.youtubeClient.Search(ctx, query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to search YouTube: %w", err)
	}

	var results []SearchResult
	for _, ytResult := range ytResults {
		results = append(results, SearchResult{
			ID:          ytResult.ID,
			Title:       ytResult.Title,
			Artist:      ytResult.Artist,
			Duration:    ytResult.Duration,
			Thumbnail:   ytResult.Thumbnail,
			Source:      "youtube",
			ExternalID:  ytResult.ExternalID,
			StreamURL:   ytResult.StreamURL,
			Description: ytResult.Description,
		})
	}

	return results, nil
}

// searchSoundCloud searches SoundCloud for tracks
func (a *Aggregator) searchSoundCloud(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	scResults, err := a.soundcloudClient.Search(ctx, query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to search SoundCloud: %w", err)
	}

	var results []SearchResult
	for _, scResult := range scResults {
		results = append(results, SearchResult{
			ID:          scResult.ID,
			Title:       scResult.Title,
			Artist:      scResult.Artist,
			Duration:    scResult.Duration,
			Thumbnail:   scResult.Thumbnail,
			Source:      "soundcloud",
			ExternalID:  scResult.ExternalID,
			StreamURL:   scResult.StreamURL,
			Description: scResult.Description,
		})
	}

	return results, nil
}

// SearchBySource searches a specific source only
func (a *Aggregator) SearchBySource(ctx context.Context, query string, source string, maxResults int) ([]SearchResult, error) {
	switch source {
	case "local":
		return a.searchLocal(ctx, query, maxResults)
	case "youtube":
		if a.youtubeClient == nil || a.youtubeClient.apiKey == "" {
			return nil, fmt.Errorf("youtube client not configured")
		}
		return a.searchYouTube(ctx, query, maxResults)
	case "soundcloud":
		if a.soundcloudClient == nil || a.soundcloudClient.clientID == "" {
			return nil, fmt.Errorf("soundcloud client not configured")
		}
		return a.searchSoundCloud(ctx, query, maxResults)
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}
}
