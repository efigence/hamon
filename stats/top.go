package stats

func (s *Stats) TopRate() map[string]float64 {
	hosts := make(map[string]float64)
	s.Frontends.RLock()
	defer s.Frontends.RUnlock()
	for _, v := range s.Frontends.TopRequest {
		_, v := v.List()
		for ip, ipRate := range v {
			hosts[ip] += ipRate
		}
	}
	return hosts
}
