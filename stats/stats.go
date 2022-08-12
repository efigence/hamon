package stats

import (
	"github.com/efigence/go-haproxy"
	"github.com/efigence/go-libs/ewma"
	"time"
)

const probes = 1024
const interval = time.Second

type Stats struct {
	f_total_ewma       map[string]*ewma.Ewma
	FrontendDurationMs map[string][]float64 `json:"frontend_duration_ms"`
}

func New(ch chan haproxy.HTTPRequest) *Stats {
	s := &Stats{
		f_total_ewma:       make(map[string]*ewma.Ewma, 0),
		FrontendDurationMs: map[string][]float64{},
	}
	go func() {
		for ev := range ch {
			if _, ok := s.f_total_ewma[ev.FrontendName]; !ok {
				s.f_total_ewma[ev.FrontendName] = ewma.NewEwma(interval * 1)
				s.f_total_ewma[ev.FrontendName].Set(float64(ev.TotalDurationMs), time.Now())
			}
			s.f_total_ewma[ev.FrontendName].UpdateNow(float64(ev.TotalDurationMs))
		}
	}()
	go func() {
		i := 0
		for {
			time.Sleep(interval)
			i++
			for k, v := range s.f_total_ewma {
				if _, ok := s.FrontendDurationMs[k]; !ok {
					s.FrontendDurationMs[k] = make([]float64, probes)
				}
				s.FrontendDurationMs[k][i%probes] = v.Current
			}
		}
	}()
	return s
}
