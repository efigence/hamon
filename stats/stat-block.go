package stats

import (
	"github.com/efigence/go-haproxy"
	"github.com/efigence/go-libs/ewma"
	"github.com/efigence/hamon/toplist"
	"time"
)

type StatBlock struct {
	total_ewma         map[string]*ewma.Ewma
	request_ewma       map[string]*ewma.Ewma
	response_ewma      map[string]*ewma.Ewma
	rate               map[string]*ewma.EwmaRate
	TotalDurationMs    map[string][]float64        `json:"total_duration_ms"`
	RequestDurationMs  map[string][]float64        `json:"request_duration_ms"`
	ResponseDurationMs map[string][]float64        `json:"response_duration_ms"`
	RequestRate        map[string][]float64        `json:"request_rate"`
	TopRequest         map[string]*toplist.Toplist `json:"top_request"`
}

func newStatBlock() *StatBlock {
	sb := StatBlock{
		total_ewma:         make(map[string]*ewma.Ewma, 0),
		request_ewma:       make(map[string]*ewma.Ewma, 0),
		response_ewma:      make(map[string]*ewma.Ewma, 0),
		rate:               make(map[string]*ewma.EwmaRate, 0),
		TotalDurationMs:    map[string][]float64{},
		RequestDurationMs:  map[string][]float64{},
		ResponseDurationMs: map[string][]float64{},
		RequestRate:        map[string][]float64{},
		TopRequest:         map[string]*toplist.Toplist{},
	}
	go func() {
		i := 0
		for {
			time.Sleep(interval)
			i++
			for k, v := range sb.total_ewma {
				if _, ok := sb.TotalDurationMs[k]; !ok {
					sb.TotalDurationMs[k] = make([]float64, probes)
					sb.RequestDurationMs[k] = make([]float64, probes)
					sb.ResponseDurationMs[k] = make([]float64, probes)
					sb.RequestRate[k] = make([]float64, probes)
				}
				// this is for rolling pointer
				//sb.TotalDurationMs[k][i%probes] = v.Current
				//sb.RequestRate[k][i%probes] = sb.rate[k].CurrentNow()

				// this is for new data coming from the left
				sb.TotalDurationMs[k] = append([]float64{v.Current}, sb.TotalDurationMs[k][:len(sb.TotalDurationMs[k])-1]...)
				sb.RequestDurationMs[k] = append([]float64{sb.request_ewma[k].Current}, sb.RequestDurationMs[k][:len(sb.RequestDurationMs[k])-1]...)
				sb.ResponseDurationMs[k] = append([]float64{sb.response_ewma[k].Current}, sb.ResponseDurationMs[k][:len(sb.ResponseDurationMs[k])-1]...)
				sb.RequestRate[k] = append([]float64{sb.rate[k].CurrentNow()}, sb.RequestRate[k][:len(sb.RequestRate[k])-1]...)

			}
		}

	}()
	return &sb
}

func (sb *StatBlock) Update(ev haproxy.HTTPRequest, name string) {
	ignoreDuration := false
	if ev.BadReq && ev.ServerName == "<NOSRV>" {
		if ev.TerminationReason == haproxy.TerminationClientAbort ||
			ev.TerminationReason == haproxy.TerminationClientWait {
			ignoreDuration = true
		}
	}
	if _, ok := sb.total_ewma[name]; !ok {
		sb.total_ewma[name] = ewma.NewEwma(interval * 1)
		sb.request_ewma[name] = ewma.NewEwma(interval * 1)
		sb.response_ewma[name] = ewma.NewEwma(interval * 1)
		sb.rate[name] = ewma.NewEwmaRate(interval * 1)
		if !ignoreDuration {
			sb.total_ewma[name].Set(float64(ev.TotalDurationMs), time.Now())
			if ev.RequestHeaderDurationMs > 0 {
				sb.request_ewma[name].Set(float64(ev.RequestHeaderDurationMs), time.Now())
			}
			if ev.ResponseHeaderDurationMs > 0 {
				sb.response_ewma[name].Set(float64(ev.ResponseHeaderDurationMs), time.Now())
			}
		}
		sb.TopRequest[name] = toplist.New(20, time.Minute, 2048)

	}
	if !ignoreDuration {
		sb.total_ewma[name].UpdateNow(float64(ev.TotalDurationMs))
		if ev.RequestHeaderDurationMs > 0 {
			sb.request_ewma[name].UpdateNow(float64(ev.RequestHeaderDurationMs))
		}
		if ev.ResponseHeaderDurationMs > 0 {
			sb.response_ewma[name].UpdateNow(float64(ev.ResponseHeaderDurationMs))
		}
	}
	sb.rate[name].UpdateNow()
	sb.TopRequest[name].Add(ev.ClientIP)
}
