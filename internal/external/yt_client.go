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

// YouTubeClient handles YouTube Data API v3 interactions
type YouTubeClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// YouTubeSearchResult represents a normalized search result from YouTube
type YouTubeSearchResult struct {
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

// YouTubeAPIResponse represents the raw API response
type YouTubeAPIResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title        string `json:"title"`
			Description  string `json:"description"`
			ChannelTitle string `json:"channelTitle"`
			Thumbnails   struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
				Medium struct {
					URL string `json:"url"`
				} `json:"medium"`
				High struct {
					URL string `json:"url"`
				} `json:"high"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
	PageInfo struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
}

// YouTubeVideoDetailsResponse represents video details from the API
type YouTubeVideoDetailsResponse struct {
	Items []struct {
		ID             string `json:"id"`
		ContentDetails struct {
			Duration string `json:"duration"`
		} `json:"contentDetails"`
	} `json:"items"`
}

// NewYouTubeClient creates a new YouTube API client
func NewYouTubeClient(apiKey string) *YouTubeClient {
	return &YouTubeClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://www.googleapis.com/youtube/v3",
	}
}

// Search performs a search query on YouTube
func (y *YouTubeClient) Search(ctx context.Context, query string, maxResults int) ([]YouTubeSearchResult, error) {
	if y.apiKey == "" {
		return nil, fmt.Errorf("youtube API key not configured")
	}

	// Build the search URL
	params := url.Values{}
	params.Add("part", "snippet")
	params.Add("q", query)
	params.Add("type", "video")
	params.Add("videoCategoryId", "10") // Music category
	params.Add("maxResults", fmt.Sprintf("%d", maxResults))
	params.Add("key", y.apiKey)

	searchURL := fmt.Sprintf("%s/search?%s", y.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("youtube API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp YouTubeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var videoIDs []string
	for _, item := range apiResp.Items {
		if item.ID.VideoID != "" {
			videoIDs = append(videoIDs, item.ID.VideoID)
		}
	}

	// Get video durations
	durations, err := y.getVideoDurations(ctx, videoIDs)
	if err != nil {
		// Log error but don't fail the search
		durations = make(map[string]int)
	}

	var results []YouTubeSearchResult
	for _, item := range apiResp.Items {
		videoID := item.ID.VideoID
		if videoID == "" {
			continue
		}

		thumbnail := item.Snippet.Thumbnails.High.URL
		if thumbnail == "" {
			thumbnail = item.Snippet.Thumbnails.Medium.URL
		}
		if thumbnail == "" {
			thumbnail = item.Snippet.Thumbnails.Default.URL
		}

		duration := durations[videoID]

		results = append(results, YouTubeSearchResult{
			ID:          videoID,
			Title:       item.Snippet.Title,
			Artist:      item.Snippet.ChannelTitle,
			Duration:    duration,
			Thumbnail:   thumbnail,
			Source:      "youtube",
			ExternalID:  videoID,
			StreamURL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
			Description: item.Snippet.Description,
		})
	}

	return results, nil
}

// getVideoDurations fetches duration information for multiple videos
func (y *YouTubeClient) getVideoDurations(ctx context.Context, videoIDs []string) (map[string]int, error) {
	if len(videoIDs) == 0 {
		return make(map[string]int), nil
	}

	durations := make(map[string]int)

	const batchSize = 50
	for i := 0; i < len(videoIDs); i += batchSize {
		end := i + batchSize
		if end > len(videoIDs) {
			end = len(videoIDs)
		}

		batch := videoIDs[i:end]
		params := url.Values{}
		params.Add("part", "contentDetails")
		params.Add("id", joinStrings(batch, ","))
		params.Add("key", y.apiKey)

		videoURL := fmt.Sprintf("%s/videos?%s", y.baseURL, params.Encode())

		req, err := http.NewRequestWithContext(ctx, "GET", videoURL, nil)
		if err != nil {
			continue
		}

		resp, err := y.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var detailsResp YouTubeVideoDetailsResponse
			if err := json.NewDecoder(resp.Body).Decode(&detailsResp); err == nil {
				for _, item := range detailsResp.Items {
					duration := parseISO8601Duration(item.ContentDetails.Duration)
					durations[item.ID] = duration
				}
			}
		}
		resp.Body.Close()
	}

	return durations, nil
}

// parseISO8601Duration converts ISO 8601 duration (PT1H2M3S) to seconds
func parseISO8601Duration(duration string) int {
	// Simple parser for YouTube duration format: PT#H#M#S
	var hours, minutes, seconds int
	fmt.Sscanf(duration, "PT%dH%dM%dS", &hours, &minutes, &seconds)

	// Try without hours
	if hours == 0 && minutes == 0 && seconds == 0 {
		fmt.Sscanf(duration, "PT%dM%dS", &minutes, &seconds)
	}

	// Try just seconds
	if hours == 0 && minutes == 0 && seconds == 0 {
		fmt.Sscanf(duration, "PT%dS", &seconds)
	}

	return hours*3600 + minutes*60 + seconds
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
