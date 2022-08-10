package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"
)

var frameGraphProbes = 256

type runtimeStats struct {
	RuntimeGC    []float32     `json:"runtime_gc"`
	HeapHistory  []float32     `json:"heap_history"`
	Frames       int           `json:"frames"`
	GCStats      debug.GCStats `json:"gc_stats"`
	PauseHistory []float32     `json:"pause_history"`
	PauseMax     float32       `json:"pause_max"`
	CPUSys       []float32     `json:"cpu_sys"`
	CPUUser      []float32     `json:"cpu_user"`
	PrevGCCount  uint32        `json:"prev_gc_count"`
	PrevCPUSys   int64         `json:"-"`
	PrevCPUUser  int64         `json:"-"`
	PrevTS       time.Time     `json:"-"`
}

var MemStats runtime.MemStats
var stats = runtimeStats{
	RuntimeGC:    make([]float32, frameGraphProbes),
	HeapHistory:  make([]float32, frameGraphProbes),
	CPUSys:       make([]float32, frameGraphProbes),
	CPUUser:      make([]float32, frameGraphProbes),
	PauseHistory: make([]float32, len(MemStats.PauseNs)),
}

func updateGC() {
	runtime.ReadMemStats(&MemStats)
	stats.RuntimeGC[stats.Frames%frameGraphProbes] = float32(MemStats.NumGC - stats.PrevGCCount)
	stats.HeapHistory[stats.Frames%frameGraphProbes] = float32(MemStats.HeapInuse)
	stats.PauseMax = 0
	for idx, p := range MemStats.PauseNs {
		us := float32(p / 1000)
		stats.PauseHistory[idx] = us
		if stats.PauseMax < us {
			stats.PauseMax = us
		}
	}
	var rusage syscall.Rusage
	now := time.Now()
	timediff := now.Sub(stats.PrevTS)
	stats.PrevTS = now
	syscall.Getrusage(syscall.RUSAGE_SELF, &rusage)
	SysCPUUsage := rusage.Stime.Nano() - stats.PrevCPUSys
	UserCPUUsage := rusage.Utime.Nano() - stats.PrevCPUUser
	stats.PrevCPUSys = rusage.Stime.Nano()
	stats.PrevCPUUser = rusage.Utime.Nano()
	// integer division first in case we get called often, not wanna lose accuracy in floats
	stats.CPUSys[stats.Frames%frameGraphProbes] = float32(SysCPUUsage/timediff.Microseconds()) / 1000
	stats.CPUUser[stats.Frames%frameGraphProbes] = float32(UserCPUUsage/timediff.Microseconds()) / 1000
	stats.Frames++
	stats.PrevGCCount = MemStats.NumGC
}
func (b *WebBackend) GCStats(c *gin.Context) {
	// should probably be in lock but whatever
	c.JSON(http.StatusOK, stats)
}
