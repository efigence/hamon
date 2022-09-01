package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"time"
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
func (b *WebBackend) V1FrontendTop(c *gin.Context) {
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

type FrontendStats struct {
	DurationMs  map[string][]float64 `json:"duration_ms"`
	RequestRate map[string][]float64 `json:"request_rate"`
}

func (b *WebBackend) V1FrontendStats(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "public, max-age=2, immutable")
	stats := FrontendStats{
		DurationMs:  b.stats.Frontends.ResponseDurationMs,
		RequestRate: b.stats.Frontends.RequestRate,
	}
	_ = stats
	c.JSON(http.StatusOK, stats)
}

type FrontendBackendStats struct {
	TotalDurationMs    map[string][]float64 `json:"total_duration_ms"`
	RequestDurationMs  map[string][]float64 `json:"request_duration_ms"`
	ResponseDurationMs map[string][]float64 `json:"response_duration_ms"`
	RequestRate        map[string][]float64 `json:"request_rate"`
}

func (b *WebBackend) V1FrontendBackendStats(c *gin.Context) {
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

type SlowList struct {
	TS                 time.Time `json:"ts"`
	Path               string    `json:"path"`
	Client             string    `json:"client"`
	Backend            string    `json:"backend"`
	Server             string    `json:"server"`
	ServerConnCount    int       `json:"server_conn_count"`
	TotalDurationMs    int       `json:"total_duration_ms"`
	RequestDurationMs  int       `json:"request_duration_ms"`
	ResponseDurationMs int       `json:"response_duration_ms"`
}

func (b *WebBackend) V1FrontendSlow(c *gin.Context) {
	frontend := c.Param("name")
	if _, ok := b.stats.Frontends.Slowlog[frontend]; !ok {
		c.String(http.StatusNotFound, "no such frontend")
		return
	}
	slowlog := b.stats.Frontends.GetSlowlog(frontend)
	sort.Slice(slowlog, func(i, j int) bool {
		if slowlog[i].TS > slowlog[j].TS {
			return true
		}
		return false
	})

	c.JSON(http.StatusOK, slowlog)
}
func (b *WebBackend) Slow(c *gin.Context) {
	frontend := c.Param("name")
	if _, ok := b.stats.Frontends.Slowlog[frontend]; !ok {
		c.String(http.StatusNotFound, "no such frontend")
		return
	}
	slowlog := b.stats.Frontends.GetSlowlog(frontend)
	sort.Slice(slowlog, func(i, j int) bool {
		if slowlog[i].TS > slowlog[j].TS {
			return true
		}
		return false
	})
	slow := []SlowList{}
	for _, e := range slowlog {
		slow = append(slow, SlowList{
			TS:                 e.Timestamp(),
			Path:               e.RequestPath,
			Client:             e.ClientIP,
			Backend:            e.BackendName,
			Server:             e.ServerName,
			ServerConnCount:    int(e.ServerConn),
			TotalDurationMs:    e.TotalDurationMs,
			RequestDurationMs:  e.RequestHeaderDurationMs,
			ResponseDurationMs: e.ResponseHeaderDurationMs,
		})
	}
	c.HTML(http.StatusOK, "slowlog.tmpl", gin.H{
		"title": fmt.Sprintf("frontend [%s] slowlog", frontend),
		"slow":  slow,
	})
}
