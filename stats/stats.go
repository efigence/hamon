package stats

import (
	"github.com/efigence/go-haproxy"
	"github.com/efigence/go-libs/ewma"
	"github.com/efigence/hamon/toplist"
	"time"
)

const probes = 1024
const interval = time.Second

type Stats struct {
	f_total_ewma        map[string]*ewma.Ewma
	f_rate              map[string]*ewma.EwmaRate
	FrontendDurationMs  map[string][]float64 `json:"frontend_duration_ms"`
	FrontendRequestRate map[string][]float64 `json:"frontend_request_rate"`
	FrontendTopRequest  map[string]*toplist.Toplist
}

func New(ch chan haproxy.HTTPRequest) *Stats {
	s := &Stats{
		f_total_ewma:        make(map[string]*ewma.Ewma, 0),
		f_rate:              make(map[string]*ewma.EwmaRate, 0),
		FrontendDurationMs:  map[string][]float64{},
		FrontendRequestRate: map[string][]float64{},
		FrontendTopRequest:  map[string]*toplist.Toplist{},
	}
	go func() {
		for ev := range ch {
			// ignore duration on bad requests
			// we're interested whether backend responds
			// this is mostly done to ignore browser preflight connections
			ignoreDuration := false
			if ev.BadReq && ev.ServerName == "<NOSRV>" {
				if ev.TerminationReason == haproxy.TerminationClientAbort ||
					ev.TerminationReason == haproxy.TerminationClientWait {
					ignoreDuration = true
				}
			}
			if _, ok := s.f_total_ewma[ev.FrontendName]; !ok {
				s.f_total_ewma[ev.FrontendName] = ewma.NewEwma(interval * 1)
				s.f_rate[ev.FrontendName] = ewma.NewEwmaRate(interval * 1)
				if !ignoreDuration {
					s.f_total_ewma[ev.FrontendName].Set(float64(ev.ResponseHeaderDurationMs), time.Now())
				}
				s.FrontendTopRequest[ev.FrontendName] = toplist.New(20, time.Minute, 2048)

			}
			if !ignoreDuration {
				s.f_total_ewma[ev.FrontendName].UpdateNow(float64(ev.ResponseHeaderDurationMs))
			}
			s.f_rate[ev.FrontendName].UpdateNow()
			s.FrontendTopRequest[ev.FrontendName].Add(ev.ClientIP)
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
					s.FrontendRequestRate[k] = make([]float64, probes)
				}
				// this is for rolling pointer
				//s.FrontendDurationMs[k][i%probes] = v.Current
				//s.FrontendRequestRate[k][i%probes] = s.f_rate[k].CurrentNow()

				// this is for new data coming from the left
				s.FrontendDurationMs[k] = append([]float64{v.Current}, s.FrontendDurationMs[k][:len(s.FrontendDurationMs[k])-1]...)
				s.FrontendRequestRate[k] = append([]float64{s.f_rate[k].CurrentNow()}, s.FrontendRequestRate[k][:len(s.FrontendRequestRate[k])-1]...)
			}
		}
	}()
	return s
}
