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

func (b *WebBackend) FrontendTop(c *gin.Context) {
	name := c.Param("name")
	c.Writer.Header().Set("Cache-Control", "public, max-age=2, immutable")
	if f, ok := b.stats.FrontendTopRequest[name]; ok {
		topList, top := f.List()
		c.JSON(http.StatusOK, gin.H{
			"top_list": topList,
			"top":      top,
		})
	} else {
		c.String(http.StatusNotFound, "frontend not found")
	}
}
