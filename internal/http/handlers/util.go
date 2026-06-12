package handlers

import "github.com/gin-gonic/gin"

// MaxUploadBytes is the maximum allowed size of a single uploaded audio file.
// It defaults to 5 MiB and is overridden at startup from configuration
// (MAX_UPLOAD_BYTES). Enforced per file on both the single and bulk paths to
// bound memory use and keep the upload surface from being abused.
var MaxUploadBytes int64 = 5 << 20

// detectAudioFormat reports the audio format of a payload from its leading magic
// bytes, returning the canonical file extension and whether it is a supported type.
// This is a content check, so renamed/non-audio payloads are rejected.
func detectAudioFormat(head []byte) (ext string, ok bool) {
	if len(head) >= 3 && head[0] == 'I' && head[1] == 'D' && head[2] == '3' {
		return "mp3", true
	}
	if len(head) >= 2 && head[0] == 0xFF && head[1]&0xE0 == 0xE0 {
		return "mp3", true
	}
	if len(head) >= 4 && string(head[0:4]) == "fLaC" {
		return "flac", true
	}
	if len(head) >= 4 && string(head[0:4]) == "OggS" {
		return "ogg", true
	}
	if len(head) >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WAVE" {
		return "wav", true
	}
	if len(head) >= 8 && string(head[4:8]) == "ftyp" {
		return "m4a", true
	}
	return "", false
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
