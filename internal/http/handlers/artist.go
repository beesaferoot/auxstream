package handlers

import (
	"auxstream/internal/db"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateArtistRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateArtistHandler reads a JSON body with a required name and responds 201
// with the created artist.
func CreateArtistHandler(c *gin.Context, r db.ArtistRepo) {
	var req CreateArtistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	artist, err := r.CreateArtist(c, req.Name)
	if err != nil {
		log.Printf("CreateArtist error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to create artist"))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": artist,
	})
}

// GetArtistByIdHandler reads the id from the URL path; 400 on a malformed UUID
// and 404 when the artist is absent.
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

// GetArtistTracksHandler pages an artist's tracks (pagesize/pagenumber query
// params, defaulting to 20/1). The artist id comes from the path; a missing
// artist yields 404 before any track lookup.
func GetArtistTracksHandler(c *gin.Context, trackRepo db.TrackRepo, artistRepo db.ArtistRepo) {
	artistIdStr := c.Param("id")
	artistId, err := uuid.Parse(artistIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid artist ID format"))
		return
	}

	_, err = artistRepo.GetArtistById(c, artistId)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("artist not found"))
		return
	}

	// Seed defaults before binding; absent query params leave these in place.
	var params GetArtistTracksQueryParams
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
		log.Printf("GetTracksByArtistId error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to fetch tracks"))
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
