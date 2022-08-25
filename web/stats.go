package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strconv"
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

func (b *WebBackend) TopRate(c *gin.Context) {
	minRate, err := strconv.ParseFloat(c.Param("rate"), 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("need valid rate in path[%s]", err))
	}
	rate := map[string]float64{}
	for _, v := range b.stats.FrontendTopRequest {
		_, v := v.List()
		for ip, ipRate := range v {
			if ipRate > minRate {
				parsedIp := net.ParseIP(ip)
				if parsedIp != nil {
					v, _ := rate[ip]
					rate[ip] = v + ipRate
				}
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"ip_rate": rate,
	})
}
