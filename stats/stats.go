package stats

import (
	"fmt"
	"github.com/efigence/go-haproxy"
	"github.com/efigence/go-mon"
	"time"
)

const probes = 1024
const interval = time.Second
const slowReqLogSize = 64

type Stats struct {
	Frontends         *StatBlock
	Backends          *StatBlock
	FrontendToBackend map[string]*StatBlock
}

func New(ch chan haproxy.HTTPRequest) *Stats {
	s := &Stats{
		Frontends:         newStatBlock(),
		Backends:          newStatBlock(),
		FrontendToBackend: make(map[string]*StatBlock),
	}
	go func() {
		for ev := range ch {
			s.Frontends.Update(ev, ev.FrontendName)
			s.Backends.Update(ev, ev.BackendName)
			if _, ok := s.FrontendToBackend[ev.FrontendName]; !ok {
				s.FrontendToBackend[ev.FrontendName] = newStatBlock()
			}
			s.FrontendToBackend[ev.FrontendName].Update(ev, ev.BackendName)
		}
	}()
	go s.runStats()
	return s
}

func (s *Stats) runStats() {
	st := map[int]mon.Metric{}
	quantiles := 10
	step := 10
	for i := 0; i < quantiles; i++ {
		rateQ := i * step
		if i == 0 {
			// no need to return "above zero" metric
			st[0] = mon.NewGauge("nothing")
		} else {
			st[i] = mon.GlobalRegistry.MustRegister("hamon_ip_above_rate", mon.NewGauge("requests"), map[string]string{"rate": fmt.Sprintf("%d", rateQ)})
		}
	}
	for {
		time.Sleep(time.Second)
		top := s.TopRate()
		sum := map[int]int{}
		for i := 0; i < step; i++ {
			sum[i] = 0
		}
		for _, rate := range top {
			for i := 0; i < quantiles; i++ {
				if rate > float64(i*step) {
					sum[i]++
				}
			}
		}
		for i := 0; i < quantiles; i++ {
			st[i].Update(float64(sum[i]))
		}
	}
}
