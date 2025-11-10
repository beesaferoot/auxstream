package handlers

import (
	"auxstream/internal/search"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SearchHandler handles unified search requests
func SearchHandler(c *gin.Context, searchService *search.Service) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, errorResponse("query parameter 'q' is required"))
		return
	}

	source := c.Query("source") // "local", "youtube", or empty for all
	maxResults := 20

	if maxResultsStr := c.Query("max_results"); maxResultsStr != "" {
		var maxResultsParam int
		if _, err := c.GetQuery("max_results"); err {
			if parsedMaxResults, parseErr := parseIntParam(maxResultsStr); parseErr == nil {
				maxResultsParam = parsedMaxResults
				if maxResultsParam > 0 {
					maxResults = maxResultsParam
				}
			}
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

// parseIntParam safely parses an integer parameter
func parseIntParam(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
