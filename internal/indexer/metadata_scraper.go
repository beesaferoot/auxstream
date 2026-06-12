package indexer

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScrapedMetadata is one track's metadata as extracted by a scraper; fields the
// source doesn't expose are left zero.
type ScrapedMetadata struct {
	ID          string
	Title       string
	Artist      string
	Album       string
	Duration    int // in seconds
	Thumbnail   string
	SourceURL   string
	Source      string
	Description string
	ReleaseDate *time.Time
	Genre       string
}

// MetadataScraper is the per-source contract: scrape one track's metadata from
// its page, search the source, and report the source name used as its registry key.
type MetadataScraper interface {
	ScrapeTrack(ctx context.Context, url string) (*ScrapedMetadata, error)
	SearchTracks(ctx context.Context, query string, limit int) ([]*ScrapedMetadata, error)
	GetSourceName() string
}

// BaseScraper provides the shared HTTP fetch used by source-specific scrapers,
// which embed it and supply their own parsing.
type BaseScraper struct {
	client     *http.Client
	sourceName string
}

// NewBaseScraper returns a BaseScraper with a 30s HTTP timeout for slow pages.
func NewBaseScraper(sourceName string) *BaseScraper {
	return &BaseScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		sourceName: sourceName,
	}
}

// FetchHTML GETs url and parses it into a goquery document. Any non-200 status
// is returned as an error rather than a partial document.
func (s *BaseScraper) FetchHTML(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Pose as a desktop browser; some sources serve different markup (or block)
	// requests with a non-browser User-Agent.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

// GenerateTrackID derives a stable ID by hashing source and URL together, so
// the same track always maps to the same ID and IDs can't collide across sources.
func GenerateTrackID(source, url string) string {
	hash := md5.Sum([]byte(source + ":" + url))
	return hex.EncodeToString(hash[:])
}

// ExtractDuration parses a duration to seconds, accepting both "M:SS" clock
// strings and ISO 8601 ("PT1H2M3S"). Unrecognized input yields 0.
func ExtractDuration(durationStr string) int {
	if matched, _ := regexp.MatchString(`^\d+:\d+$`, durationStr); matched {
		parts := strings.Split(durationStr, ":")
		if len(parts) == 2 {
			var minutes, seconds int
			fmt.Sscanf(parts[0], "%d", &minutes)
			fmt.Sscanf(parts[1], "%d", &seconds)
			return minutes*60 + seconds
		}
	}

	if strings.HasPrefix(durationStr, "PT") {
		re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
		matches := re.FindStringSubmatch(durationStr)
		if len(matches) > 0 {
			var hours, minutes, seconds int
			if matches[1] != "" {
				fmt.Sscanf(matches[1], "%d", &hours)
			}
			if matches[2] != "" {
				fmt.Sscanf(matches[2], "%d", &minutes)
			}
			if matches[3] != "" {
				fmt.Sscanf(matches[3], "%d", &seconds)
			}
			return hours*3600 + minutes*60 + seconds
		}
	}

	return 0
}

type AudiomackScraper struct {
	*BaseScraper
}

func NewAudiomackScraper() *AudiomackScraper {
	return &AudiomackScraper{
		BaseScraper: NewBaseScraper("audiomack"),
	}
}

// ScrapeTrack reads Audiomack track metadata from the page's OpenGraph/meta
// tags. Missing tags are left as zero values rather than treated as errors.
func (s *AudiomackScraper) ScrapeTrack(ctx context.Context, url string) (*ScrapedMetadata, error) {
	doc, err := s.FetchHTML(ctx, url)
	if err != nil {
		return nil, err
	}

	metadata := &ScrapedMetadata{
		ID:        GenerateTrackID("audiomack", url),
		SourceURL: url,
		Source:    "audiomack",
	}

	if title := doc.Find("meta[property='og:title']").AttrOr("content", ""); title != "" {
		metadata.Title = title
	}

	if artist := doc.Find("meta[name='music:musician']").AttrOr("content", ""); artist != "" {
		metadata.Artist = artist
	}

	if thumbnail := doc.Find("meta[property='og:image']").AttrOr("content", ""); thumbnail != "" {
		metadata.Thumbnail = thumbnail
	}

	if desc := doc.Find("meta[property='og:description']").AttrOr("content", ""); desc != "" {
		metadata.Description = desc
	}

	if duration := doc.Find("meta[property='music:duration']").AttrOr("content", ""); duration != "" {
		metadata.Duration = ExtractDuration(duration)
	}

	return metadata, nil
}

// SearchTracks is a stub: Audiomack has no search endpoint wired up yet, so it
// logs and returns no results (not an error) to keep callers working.
func (s *AudiomackScraper) SearchTracks(ctx context.Context, query string, limit int) ([]*ScrapedMetadata, error) {
	log.Printf("Audiomack search not implemented yet for query: %s", query)
	return []*ScrapedMetadata{}, nil
}

func (s *AudiomackScraper) GetSourceName() string {
	return "audiomack"
}

type BoomplayScraper struct {
	*BaseScraper
}

func NewBoomplayScraper() *BoomplayScraper {
	return &BoomplayScraper{
		BaseScraper: NewBaseScraper("boomplay"),
	}
}

// ScrapeTrack reads Boomplay metadata from OpenGraph tags. Boomplay has no
// dedicated artist tag, so the artist is split out of an "Artist - Title" og:title.
func (s *BoomplayScraper) ScrapeTrack(ctx context.Context, url string) (*ScrapedMetadata, error) {
	doc, err := s.FetchHTML(ctx, url)
	if err != nil {
		return nil, err
	}

	metadata := &ScrapedMetadata{
		ID:        GenerateTrackID("boomplay", url),
		SourceURL: url,
		Source:    "boomplay",
	}

	if title := doc.Find("meta[property='og:title']").AttrOr("content", ""); title != "" {
		metadata.Title = title
	}

	if metadata.Title != "" && strings.Contains(metadata.Title, " - ") {
		parts := strings.SplitN(metadata.Title, " - ", 2)
		if len(parts) == 2 {
			metadata.Artist = parts[0]
			metadata.Title = parts[1]
		}
	}

	if thumbnail := doc.Find("meta[property='og:image']").AttrOr("content", ""); thumbnail != "" {
		metadata.Thumbnail = thumbnail
	}

	if desc := doc.Find("meta[property='og:description']").AttrOr("content", ""); desc != "" {
		metadata.Description = desc
	}

	return metadata, nil
}

// SearchTracks is a stub: no Boomplay search endpoint is wired up yet, so it
// logs and returns no results (not an error) to keep callers working.
func (s *BoomplayScraper) SearchTracks(ctx context.Context, query string, limit int) ([]*ScrapedMetadata, error) {
	log.Printf("Boomplay search not implemented yet for query: %s", query)
	return []*ScrapedMetadata{}, nil
}

func (s *BoomplayScraper) GetSourceName() string {
	return "boomplay"
}

// ScraperRegistry maps a source name to its scraper, letting ScrapeURL dispatch
// by the source detected from a URL.
type ScraperRegistry struct {
	scrapers map[string]MetadataScraper
}

// NewScraperRegistry returns a registry pre-populated with the built-in
// Audiomack and Boomplay scrapers.
func NewScraperRegistry() *ScraperRegistry {
	registry := &ScraperRegistry{
		scrapers: make(map[string]MetadataScraper),
	}

	registry.Register(NewAudiomackScraper())
	registry.Register(NewBoomplayScraper())

	return registry
}

func (r *ScraperRegistry) Register(scraper MetadataScraper) {
	r.scrapers[scraper.GetSourceName()] = scraper
}

func (r *ScraperRegistry) GetScraper(source string) (MetadataScraper, bool) {
	scraper, exists := r.scrapers[source]
	return scraper, exists
}

// ScrapeURL detects the source from url and delegates to its scraper. It errors
// if the URL is from an unrecognized host or no scraper is registered for it.
func (r *ScraperRegistry) ScrapeURL(ctx context.Context, url string) (*ScrapedMetadata, error) {
	source := DetectSourceFromURL(url)
	if source == "" {
		return nil, fmt.Errorf("unsupported URL: %s", url)
	}

	scraper, exists := r.GetScraper(source)
	if !exists {
		return nil, fmt.Errorf("no scraper for source: %s", source)
	}

	return scraper.ScrapeTrack(ctx, url)
}

// DetectSourceFromURL returns the source name for a URL by host substring, or
// "" if unrecognized. Note it recognizes soundcloud and youtube even though no
// scraper is registered for them, so detection and scraper support differ.
func DetectSourceFromURL(url string) string {
	url = strings.ToLower(url)

	if strings.Contains(url, "audiomack.com") {
		return "audiomack"
	}
	if strings.Contains(url, "boomplay.com") {
		return "boomplay"
	}
	if strings.Contains(url, "soundcloud.com") {
		return "soundcloud"
	}
	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		return "youtube"
	}

	return ""
}
