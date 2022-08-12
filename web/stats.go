package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (b *WebBackend) FrontendStats(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "public, max-age=2, immutable")
	// should probably be in lock but whatever
	c.JSON(http.StatusOK, b.stats)
}
