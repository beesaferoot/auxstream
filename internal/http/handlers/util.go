package handlers

import "github.com/gin-gonic/gin"

// MaxUploadBytes is the maximum allowed size of a single uploaded audio file.
// It defaults to 5 MiB and is overridden at startup from configuration
// (MAX_UPLOAD_BYTES). Enforced per file on both the single and bulk paths to
// bound memory use and keep the upload surface from being abused.
var MaxUploadBytes int64 = 5 << 20

// looksLikeMP3 reports whether head begins with a plausible MP3 signature:
// either an ID3v2 tag or an MPEG audio frame sync. This is a content check, so
// it rejects non-audio payloads that were merely renamed with a .mp3 extension.
func looksLikeMP3(head []byte) bool {
	if len(head) < 3 {
		return false
	}
	// ID3v2-tagged files start with the ASCII bytes "ID3".
	if head[0] == 'I' && head[1] == 'D' && head[2] == '3' {
		return true
	}
	// Headerless MP3 frames start with an 11-bit frame sync: 0xFF followed by
	// a byte whose top three bits are set.
	if head[0] == 0xFF && head[1]&0xE0 == 0xE0 {
		return true
	}
	return false
}

// contextKey is a private type for request-context keys, avoiding collisions
// with keys defined in other packages (a bare string key risks silent clashes).
type contextKey string

// CacheContextKey is where the request-scoped cache client is stored on the
// request context by the cache-injection middleware.
const CacheContextKey contextKey = "cacheClient"

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}
