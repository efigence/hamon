package stats

import (
	"github.com/efigence/go-haproxy"
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
	return s
}
