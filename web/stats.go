package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strconv"
)

type FrontendBackendStats struct {
	TotalDurationMs    map[string][]float64 `json:"total_duration_ms"`
	RequestDurationMs  map[string][]float64 `json:"request_duration_ms"`
	ResponseDurationMs map[string][]float64 `json:"response_duration_ms"`
	RequestRate        map[string][]float64 `json:"request_rate"`
}

func (b *WebBackend) FrontendBackendStats(c *gin.Context) {
	frontend := c.Param("frontend")
	if _, ok := b.stats.FrontendToBackend[frontend]; !ok {
		c.String(http.StatusNotFound, "no such frontend")
		return
	}

	c.Writer.Header().Set("Cache-Control", "public, max-age=2, immutable")
	stats := FrontendBackendStats{
		TotalDurationMs:    b.stats.FrontendToBackend[frontend].TotalDurationMs,
		RequestDurationMs:  b.stats.FrontendToBackend[frontend].RequestDurationMs,
		ResponseDurationMs: b.stats.FrontendToBackend[frontend].ResponseDurationMs,
		RequestRate:        b.stats.FrontendToBackend[frontend].RequestRate,
	}
	_ = stats
	c.JSON(http.StatusOK, stats)
}

type FrontendStats struct {
	DurationMs  map[string][]float64 `json:"duration_ms"`
	RequestRate map[string][]float64 `json:"request_rate"`
}

func (b *WebBackend) FrontendStats(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "public, max-age=2, immutable")
	stats := FrontendStats{
		DurationMs:  b.stats.Frontends.ResponseDurationMs,
		RequestRate: b.stats.Frontends.RequestRate,
	}
	_ = stats
	c.JSON(http.StatusOK, stats)
}

func (b *WebBackend) FrontendTop(c *gin.Context) {
	name := c.Param("name")
	c.Writer.Header().Set("Cache-Control", "public, max-age=2, immutable")
	if f, ok := b.stats.Frontends.TopRequest[name]; ok {
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
	for _, v := range b.stats.Frontends.TopRequest {
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
