package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SoundCloudClient handles SoundCloud API interactions
type SoundCloudClient struct {
	clientID   string
	httpClient *http.Client
	baseURL    string
}

// SoundCloudSearchResult represents a normalized search result from SoundCloud
type SoundCloudSearchResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Duration    int    `json:"duration"` // in seconds
	Thumbnail   string `json:"thumbnail"`
	Source      string `json:"source"`
	ExternalID  string `json:"external_id"`
	StreamURL   string `json:"stream_url"`
	Description string `json:"description"`
}

// SoundCloudAPIResponse represents the raw API response
type SoundCloudAPIResponse struct {
	Collection []struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Duration    int    `json:"duration"` // Duration in milliseconds
		User        struct {
			Username string `json:"username"`
		} `json:"user"`
		ArtworkURL   string `json:"artwork_url"`
		PermalinkURL string `json:"permalink_url"`
		Streamable   bool   `json:"streamable"`
	} `json:"collection"`
	NextHref string `json:"next_href"`
}

// NewSoundCloudClient creates a new SoundCloud API client
func NewSoundCloudClient(clientID string) *SoundCloudClient {
	return &SoundCloudClient{
		clientID: clientID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.soundcloud.com",
	}
}

// Search returns up to maxResults streamable tracks matching query, normalized
// with durations in seconds and thumbnails upgraded to 500x500. Non-streamable
// tracks are filtered out, so fewer than maxResults may come back.
func (s *SoundCloudClient) Search(ctx context.Context, query string, maxResults int) ([]SoundCloudSearchResult, error) {
	if s.clientID == "" {
		return nil, fmt.Errorf("soundcloud client ID not configured")
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("client_id", s.clientID)
	params.Add("limit", fmt.Sprintf("%d", maxResults))
	params.Add("linked_partitioning", "1")

	searchURL := fmt.Sprintf("%s/tracks?%s", s.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("soundcloud API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp SoundCloudAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []SoundCloudSearchResult
	for _, track := range apiResp.Collection {
		// Skip non-streamable tracks: we can't play them, so they're noise.
		if !track.Streamable {
			continue
		}

		// API reports duration in milliseconds; our model uses seconds.
		durationSeconds := track.Duration / 1000

		thumbnail := track.ArtworkURL
		if thumbnail == "" {
			thumbnail = getDefaultSoundCloudThumbnail()
		} else {
			thumbnail = upgradeSoundCloudThumbnail(thumbnail)
		}

		results = append(results, SoundCloudSearchResult{
			ID:          fmt.Sprintf("%d", track.ID),
			Title:       track.Title,
			Artist:      track.User.Username,
			Duration:    durationSeconds,
			Thumbnail:   thumbnail,
			Source:      "soundcloud",
			ExternalID:  fmt.Sprintf("%d", track.ID),
			StreamURL:   track.PermalinkURL,
			Description: track.Description,
		})

		// Defensive cap: the API limit already bounds the collection, but enforce
		// maxResults here too in case it returns more.
		if len(results) >= maxResults {
			break
		}
	}

	return results, nil
}

// GetTrack fetches one track by SoundCloud ID, normalized like Search results.
// Returns an error if the track is not streamable.
func (s *SoundCloudClient) GetTrack(ctx context.Context, trackID string) (*SoundCloudSearchResult, error) {
	if s.clientID == "" {
		return nil, fmt.Errorf("soundcloud client ID not configured")
	}

	params := url.Values{}
	params.Add("client_id", s.clientID)

	trackURL := fmt.Sprintf("%s/tracks/%s?%s", s.baseURL, trackID, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", trackURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("soundcloud API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse single track response (same structure as collection item)
	var track struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Duration    int    `json:"duration"`
		User        struct {
			Username string `json:"username"`
		} `json:"user"`
		ArtworkURL   string `json:"artwork_url"`
		PermalinkURL string `json:"permalink_url"`
		Streamable   bool   `json:"streamable"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !track.Streamable {
		return nil, fmt.Errorf("track is not streamable")
	}

	durationSeconds := track.Duration / 1000

	thumbnail := track.ArtworkURL
	if thumbnail == "" {
		thumbnail = getDefaultSoundCloudThumbnail()
	} else {
		thumbnail = upgradeSoundCloudThumbnail(thumbnail)
	}

	return &SoundCloudSearchResult{
		ID:          fmt.Sprintf("%d", track.ID),
		Title:       track.Title,
		Artist:      track.User.Username,
		Duration:    durationSeconds,
		Thumbnail:   thumbnail,
		Source:      "soundcloud",
		ExternalID:  fmt.Sprintf("%d", track.ID),
		StreamURL:   track.PermalinkURL,
		Description: track.Description,
	}, nil
}

// ResolveURL turns a public SoundCloud page URL into track metadata via the
// /resolve endpoint. Errors if the URL resolves to something other than a track
// (e.g. a user or playlist) or to a non-streamable track.
func (s *SoundCloudClient) ResolveURL(ctx context.Context, soundcloudURL string) (*SoundCloudSearchResult, error) {
	if s.clientID == "" {
		return nil, fmt.Errorf("soundcloud client ID not configured")
	}

	params := url.Values{}
	params.Add("url", soundcloudURL)
	params.Add("client_id", s.clientID)

	resolveURL := fmt.Sprintf("%s/resolve?%s", s.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", resolveURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("soundcloud API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var track struct {
		ID          int64  `json:"id"`
		Kind        string `json:"kind"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Duration    int    `json:"duration"`
		User        struct {
			Username string `json:"username"`
		} `json:"user"`
		ArtworkURL   string `json:"artwork_url"`
		PermalinkURL string `json:"permalink_url"`
		Streamable   bool   `json:"streamable"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if track.Kind != "track" {
		return nil, fmt.Errorf("URL does not resolve to a track")
	}

	if !track.Streamable {
		return nil, fmt.Errorf("track is not streamable")
	}

	durationSeconds := track.Duration / 1000

	thumbnail := track.ArtworkURL
	if thumbnail == "" {
		thumbnail = getDefaultSoundCloudThumbnail()
	} else {
		thumbnail = upgradeSoundCloudThumbnail(thumbnail)
	}

	return &SoundCloudSearchResult{
		ID:          fmt.Sprintf("%d", track.ID),
		Title:       track.Title,
		Artist:      track.User.Username,
		Duration:    durationSeconds,
		Thumbnail:   thumbnail,
		Source:      "soundcloud",
		ExternalID:  fmt.Sprintf("%d", track.ID),
		StreamURL:   track.PermalinkURL,
		Description: track.Description,
	}, nil
}

// upgradeSoundCloudThumbnail rewrites a thumbnail URL to the 500x500 variant.
// SoundCloud encodes size in the filename (large.jpg is only 100x100, crop.jpg
// 400x400, t300x300.jpg 300x300), so we swap any of those for t500x500.jpg.
func upgradeSoundCloudThumbnail(thumbnailURL string) string {
	if thumbnailURL == "" {
		return ""
	}

	result := thumbnailURL
	result = replaceSize(result, "large.jpg", "t500x500.jpg")
	result = replaceSize(result, "t300x300.jpg", "t500x500.jpg")
	result = replaceSize(result, "crop.jpg", "t500x500.jpg")

	return result
}

// replaceSize swaps oldSize for newSize only when url ends with oldSize, so it
// rewrites the size suffix without touching a coincidental match earlier in the URL.
func replaceSize(url, oldSize, newSize string) string {
	if len(url) > len(oldSize) {
		idx := len(url) - len(oldSize)
		if url[idx:] == oldSize {
			return url[:idx] + newSize
		}
	}
	return url
}

// getDefaultSoundCloudThumbnail returns SoundCloud's generic avatar, used when a
// track has no artwork.
func getDefaultSoundCloudThumbnail() string {
	return "https://a-v2.sndcdn.com/assets/images/sc-icons/ios-a62dfc8f.png"
}

// SearchTrending lists tracks, optionally filtered by genre, normalized like
// Search results. An empty genre returns tracks across all genres.
func (s *SoundCloudClient) SearchTrending(ctx context.Context, genre string, maxResults int) ([]SoundCloudSearchResult, error) {
	if s.clientID == "" {
		return nil, fmt.Errorf("soundcloud client ID not configured")
	}

	params := url.Values{}
	params.Add("client_id", s.clientID)
	params.Add("limit", fmt.Sprintf("%d", maxResults))
	params.Add("linked_partitioning", "1")

	if genre != "" {
		params.Add("genres", genre)
	}

	trendingURL := fmt.Sprintf("%s/tracks?%s", s.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", trendingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("soundcloud API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp SoundCloudAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []SoundCloudSearchResult
	for _, track := range apiResp.Collection {
		if !track.Streamable {
			continue
		}

		durationSeconds := track.Duration / 1000
		thumbnail := track.ArtworkURL
		if thumbnail == "" {
			thumbnail = getDefaultSoundCloudThumbnail()
		} else {
			thumbnail = upgradeSoundCloudThumbnail(thumbnail)
		}

		results = append(results, SoundCloudSearchResult{
			ID:          fmt.Sprintf("%d", track.ID),
			Title:       track.Title,
			Artist:      track.User.Username,
			Duration:    durationSeconds,
			Thumbnail:   thumbnail,
			Source:      "soundcloud",
			ExternalID:  fmt.Sprintf("%d", track.ID),
			StreamURL:   track.PermalinkURL,
			Description: track.Description,
		})

		if len(results) >= maxResults {
			break
		}
	}

	return results, nil
}
