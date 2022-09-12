package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strconv"
)

func (b *WebBackend) V1TopRate(c *gin.Context) {
	minRate, err := strconv.ParseFloat(c.Param("rate"), 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("need valid filteredRate in path[%s]", err))
	}
	rate := b.stats.TopRate()
	filteredRate := map[string]float64{}
	for ip, ipRate := range rate {
		if ipRate > minRate {
			parsedIp := net.ParseIP(ip)
			if parsedIp != nil {
				v, _ := filteredRate[ip]
				filteredRate[ip] = v + ipRate
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"ip_rate": filteredRate,
	})
}
