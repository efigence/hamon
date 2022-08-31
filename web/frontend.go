package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
)

func (b *WebBackend) Frontend(c *gin.Context) {
	frontend := c.Param("name")
	backendList := []string{}
	if _, ok := b.stats.FrontendToBackend[frontend]; !ok {
		c.String(http.StatusNotFound, "no such frontend")
		return
	}
	for k, _ := range b.stats.FrontendToBackend[frontend].RequestRate {
		backendList = append(backendList, k)
	}
	sort.Strings(backendList)
	c.HTML(http.StatusOK, "frontend.tmpl", gin.H{
		"title":    fmt.Sprintf("frontend [%s] stats", frontend),
		"frontend": frontend,
		"backends": backendList,
	})

}
