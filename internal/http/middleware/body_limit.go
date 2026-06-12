package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySize caps the total size of a request body. It wraps the body in an
// http.MaxBytesReader, so an over-limit request fails when the handler reads it
// (returning a 413-style error) instead of being buffered in full. This is the
// hard ceiling that bounds whole-request size; the per-file limit enforced in
// the upload handlers bounds each individual file within that request.
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
