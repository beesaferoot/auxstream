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

type ScrapedMetadata struct {
	ID          string
	Title       string
	Artist      string
	Album       string
	Duration    int
	Thumbnail   string
	SourceURL   string
	Source      string
	Description string
	ReleaseDate *time.Time
	Genre       string
}

type MetadataScraper interface {
	ScrapeTrack(ctx context.Context, url string) (*ScrapedMetadata, error)
	SearchTracks(ctx context.Context, query string, limit int) ([]*ScrapedMetadata, error)
	GetSourceName() string
}

type BaseScraper struct {
	client     *http.Client
	sourceName string
}

func NewBaseScraper(sourceName string) *BaseScraper {
	return &BaseScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		sourceName: sourceName,
	}
}

func (s *BaseScraper) FetchHTML(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

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

func GenerateTrackID(source, url string) string {
	hash := md5.Sum([]byte(source + ":" + url))
	return hex.EncodeToString(hash[:])
}

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

func (s *BoomplayScraper) SearchTracks(ctx context.Context, query string, limit int) ([]*ScrapedMetadata, error) {
	log.Printf("Boomplay search not implemented yet for query: %s", query)
	return []*ScrapedMetadata{}, nil
}

func (s *BoomplayScraper) GetSourceName() string {
	return "boomplay"
}

type ScraperRegistry struct {
	scrapers map[string]MetadataScraper
}

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
