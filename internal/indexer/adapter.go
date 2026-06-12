package indexer

// IndexerAdapter exposes the indexing service to callers that must not depend on
// the indexer's internal ScrapedMetadata type, mapping results to IndexedTrackResult.
type IndexerAdapter struct {
	service *IndexingService
}

// NewIndexerAdapter wraps an IndexingService.
func NewIndexerAdapter(service *IndexingService) *IndexerAdapter {
	return &IndexerAdapter{
		service: service,
	}
}

// IndexedTrackResult is the adapter's source-agnostic view of an indexed track,
// decoupling callers from the indexer's ScrapedMetadata.
type IndexedTrackResult struct {
	ID          string
	Title       string
	Artist      string
	Duration    int
	Thumbnail   string
	Source      string
	SourceURL   string
	Description string
}

func (a *IndexerAdapter) SearchIndexedTracks(source, query string, limit int) []IndexedTrackResult {
	tracks := a.service.SearchIndexedTracks(source, query, limit)

	result := make([]IndexedTrackResult, len(tracks))
	for i, track := range tracks {
		result[i] = IndexedTrackResult{
			ID:          track.ID,
			Title:       track.Title,
			Artist:      track.Artist,
			Duration:    track.Duration,
			Thumbnail:   track.Thumbnail,
			Source:      track.Source,
			SourceURL:   track.SourceURL,
			Description: track.Description,
		}
	}

	return result
}
