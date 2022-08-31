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
var procstats = runtimeStats{
	RuntimeGC:    make([]float32, frameGraphProbes),
	HeapHistory:  make([]float32, frameGraphProbes),
	CPUSys:       make([]float32, frameGraphProbes),
	CPUUser:      make([]float32, frameGraphProbes),
	PauseHistory: make([]float32, len(MemStats.PauseNs)),
}

func updateGC() {
	runtime.ReadMemStats(&MemStats)
	procstats.RuntimeGC[procstats.Frames%frameGraphProbes] = float32(MemStats.NumGC - procstats.PrevGCCount)
	procstats.HeapHistory[procstats.Frames%frameGraphProbes] = float32(MemStats.HeapInuse)
	procstats.PauseMax = 0
	for idx, p := range MemStats.PauseNs {
		us := float32(p / 1000)
		procstats.PauseHistory[idx] = us
		if procstats.PauseMax < us {
			procstats.PauseMax = us
		}
	}
	var rusage syscall.Rusage
	now := time.Now()
	timediff := now.Sub(procstats.PrevTS)
	procstats.PrevTS = now
	syscall.Getrusage(syscall.RUSAGE_SELF, &rusage)
	SysCPUUsage := rusage.Stime.Nano() - procstats.PrevCPUSys
	UserCPUUsage := rusage.Utime.Nano() - procstats.PrevCPUUser
	procstats.PrevCPUSys = rusage.Stime.Nano()
	procstats.PrevCPUUser = rusage.Utime.Nano()
	// integer division first in case we get called often, not wanna lose accuracy in floats
	procstats.CPUSys[procstats.Frames%frameGraphProbes] = float32(SysCPUUsage/timediff.Microseconds()) / 1000
	procstats.CPUUser[procstats.Frames%frameGraphProbes] = float32(UserCPUUsage/timediff.Microseconds()) / 1000
	procstats.Frames++
	procstats.PrevGCCount = MemStats.NumGC
}
func (b *WebBackend) V1GCStats(c *gin.Context) {
	// should probably be in lock but whatever
	c.JSON(http.StatusOK, procstats)
}
