package indexer

type IndexerAdapter struct {
	service *IndexingService
}

func NewIndexerAdapter(service *IndexingService) *IndexerAdapter {
	return &IndexerAdapter{
		service: service,
	}
}

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
