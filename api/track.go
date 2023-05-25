package api

import (
	"auxstream/db"
	fs "auxstream/file_system"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FetchTracksHandler fetch tracks by artist (limit results < 100)
func FetchTracksHandler(c *gin.Context) {

	artist := c.Query("artist")
	tracks, err := db.DAO.SearchTrackByArtist(c, artist)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
	})

}

// AddTrackHandler add track to the system
func AddTrackHandler(c *gin.Context) {
	trackTittle := c.PostForm("title")
	trackArtist := c.PostForm("artist")
	file, err := c.FormFile("audio")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "audio for track not found",
		})
		return
	}
	raw_file, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "unable to access track audio",
		})
		return
	}
	raw_bytes := make([]byte, file.Size)
	_, err = raw_file.Read(raw_bytes)
	fileName, err := fs.Store.Save(raw_bytes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("audio upload failed: %s", err.Error()),
		})
		return
	}
	track, err := db.DAO.CreateTrack(c, trackTittle, trackArtist, fileName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("audio upload failed: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": track,
	})
}
