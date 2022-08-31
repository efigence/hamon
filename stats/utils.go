package stats

import (
	"github.com/efigence/go-haproxy"
	"sort"
	"time"
)

// put requests older than TS at the end, then ones older
func sortSlowlog(sb []haproxy.HTTPRequest, oldest time.Time) {
	timeUs := oldest.UnixMicro()
	sort.Slice(sb, func(i, j int) bool {
		if timeUs > sb[i].TS {
			return false
		}
		if timeUs > sb[j].TS {
			return true
		}
		if sb[i].TotalDurationMs > sb[j].TotalDurationMs {
			return true
		}
		return false
	})
}
