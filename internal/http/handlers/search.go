package handlers

import (
	"auxstream/internal/search"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SearchHandler handles unified search requests across all configured sources.
func SearchHandler(c *gin.Context, searchService *search.Service) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, errorResponse("query parameter 'q' is required"))
		return
	}

	source := c.Query("source")
	maxResults := 20

	if maxResultsStr := c.Query("max_results"); maxResultsStr != "" {
		if parsed, err := strconv.Atoi(maxResultsStr); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	searchReq := search.SearchRequest{
		Query:      query,
		MaxResults: maxResults,
		Source:     source,
	}

	results, err := searchService.Search(c.Request.Context(), searchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}
