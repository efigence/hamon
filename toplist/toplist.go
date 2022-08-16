package toplist

import (
	"github.com/efigence/go-libs/ewma"
	"sort"
	"sync"
	"time"
)

type Toplist struct {
	sync.Mutex
	toplistLock sync.Mutex
	size        int
	decay       time.Duration
	bufferSize  int
	topList     map[string]*ewma.EwmaRate
	events      []string
	eventsIdx   int
	lastRecalc  time.Time
}

// Initializes toplist with size (in entries), decay and buffer size(128*size if not defined, min 4*size)
func New(size int, decay time.Duration, bufferSize ...int) *Toplist {
	t := Toplist{
		size:       size,
		decay:      decay,
		lastRecalc: time.Now().Add(decay * -1),
		topList:    make(map[string]*ewma.EwmaRate, size),
	}
	if len(bufferSize) < 1 && t.bufferSize < size*4 {
		t.bufferSize = size * 128
	} else {
		t.bufferSize = bufferSize[0]
	}
	return &t
}

func (t *Toplist) Add(s string) {
	t.Lock()
	if len(t.events) > t.bufferSize {
		t.Unlock()
		t.recalculate()
		t.Lock()
	}
	t.events = append(t.events, s)
	defer t.Unlock()

}
func (t *Toplist) recalculate() {
	t.Lock()
	m := make(map[string]int)
	for _, s := range t.events {
		m[s]++
	}
	t.events = []string{}
	t.Unlock()
	t.toplistLock.Lock()
	defer t.toplistLock.Unlock()
	now := time.Now()
	diffSec := now.Sub(t.lastRecalc).Seconds()
	t.lastRecalc = now
	for k, _ := range m {
		if _, ok := t.topList[k]; ok {
			t.topList[k].UpdateValueNow(float64(m[k]))
		} else {
			perSecond := float64(m[k]) / diffSec
			t.topList[k] = ewma.NewEwmaRate(t.decay)
			t.topList[k].Set(perSecond, time.Now())
		}
	}
	if len(t.topList) > t.bufferSize {
		top := make(map[string]float64, len(t.topList))
		topKeys := make([]string, len(t.topList))
		i := 0
		for k, v := range t.topList {
			top[k] = v.CurrentNow()
			topKeys[i] = k
			i++
		}
		sort.Slice(topKeys, func(i, j int) bool {
			return top[topKeys[i]] > top[topKeys[j]]
		})
		toDelete := topKeys[t.bufferSize/2+t.size:]
		for _, k := range toDelete {
			delete(t.topList, k)
		}

	}

}

func (t *Toplist) List() (order []string, values map[string]float64) {
	top := make(map[string]float64, len(t.topList))
	topKeys := make([]string, len(t.topList))
	i := 0
	for k, v := range t.topList {
		top[k] = v.CurrentNow()
		topKeys[i] = k
		i++
	}
	sort.Slice(topKeys, func(i, j int) bool {
		return top[topKeys[i]] > top[topKeys[j]]
	})
	if len(top) > t.size {

		toDelete := topKeys[t.size:]
		for _, k := range toDelete {
			delete(top, k)
		}
		topKeys = topKeys[:t.size]
	}
	return topKeys, top
}
