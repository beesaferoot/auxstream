package handlers

import (
	"auxstream/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateArtistRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateArtistHandler creates a new artist
func CreateArtistHandler(c *gin.Context, r db.ArtistRepo) {
	var req CreateArtistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	artist, err := r.CreateArtist(c, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": artist,
	})
}

// GetArtistByIdHandler fetches an artist by ID
func GetArtistByIdHandler(c *gin.Context, r db.ArtistRepo) {
	artistIdStr := c.Param("id")
	artistId, err := uuid.Parse(artistIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid artist ID format"))
		return
	}

	artist, err := r.GetArtistById(c, artistId)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("artist not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": artist,
	})
}

type GetArtistTracksQueryParams struct {
	PageSize int `form:"pagesize" binding:"gte=0,lte=100"`
	PageNum  int `form:"pagenumber" binding:"gte=1"`
}

// GetArtistTracksHandler fetches all tracks for an artist with pagination
func GetArtistTracksHandler(c *gin.Context, trackRepo db.TrackRepo, artistRepo db.ArtistRepo) {
	artistIdStr := c.Param("id")
	artistId, err := uuid.Parse(artistIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid artist ID format"))
		return
	}

	// Verify artist exists
	_, err = artistRepo.GetArtistById(c, artistId)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("artist not found"))
		return
	}

	var params GetArtistTracksQueryParams
	// Set defaults
	params.PageSize = 20
	params.PageNum = 1

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	limit := params.PageSize
	offset := (params.PageNum - 1) * params.PageSize

	tracks, err := trackRepo.GetTracksByArtistId(c, artistId, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
		"meta": gin.H{
			"page":      params.PageNum,
			"page_size": params.PageSize,
			"artist_id": artistId,
		},
	})
}
