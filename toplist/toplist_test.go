package toplist

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

var dataShort = []string{
	"1.2.3.4",
	"1.2.3.4",
	"1.2.3.4",
	"1.2.3.4",
	"2.2.3.4",
	"2.2.3.4",
	"2.2.3.4",
	"3.2.3.4",
	"3.2.3.4",
	"4.2.3.4",
}

func TestToplist(t *testing.T) {

	list := New(4, time.Second)

	for _, d := range dataShort {
		list.Add(d)
	}
	list.recalculate()
	order, out := list.List()
	assert.Equal(t, []string{"1.2.3.4", "2.2.3.4", "3.2.3.4", "4.2.3.4"}, order)
	assert.InDeltaMapValues(t, map[string]float64{
		"1.2.3.4": 4.0,
		"2.2.3.4": 3.0,
		"3.2.3.4": 2.0,
		"4.2.3.4": 1.0,
	}, out, 0.01)

	list.lastRecalc = time.Now().Add(time.Second * -1)
	list.Add("5.2.3.4")
	list.Add("5.2.3.4")
	list.Add("5.2.3.4")
	list.recalculate()
	order, out = list.List()
	assert.InDeltaMapValues(t, map[string]float64{
		"1.2.3.4": 4.0,
		"2.2.3.4": 3.0,
		"3.2.3.4": 2.0,
		"5.2.3.4": 3.0,
	}, out, 0.01)
	//	assert.Equal(t, []string{"1.2.3.4", "2.2.3.4", "3.2.3.4", "4.2.3.4"}, order)
	list.lastRecalc = time.Now().Add(time.Second * -1)
	list.Add("2.2.3.4")
	list.Add("4.2.3.4")
	list.Add("4.2.3.4")
	list.Add("4.2.3.4")
	list.recalculate()
	order, out = list.List()
	assert.InDelta(t, list.topList["4.2.3.4"].CurrentNow(), 3.0, 0.1, "4.2.3.4")
	assert.InDelta(t, out["1.2.3.4"], 4.0, 0.1, "1.2.3.4")
	assert.InDelta(t, out["2.2.3.4"], 3.6, 0.1, "2.2.3.4")
	assert.InDelta(t, out["4.2.3.4"], 3.0, 0.1, "5.2.3.4")
	assert.InDelta(t, out["5.2.3.4"], 3.0, 0.1, "3.2.3.4")

	assert.Equal(t, []string{"1.2.3.4", "2.2.3.4", "4.2.3.4", "5.2.3.4"}, order)

}

func BenchmarkToplist_Add(b *testing.B) {
	list := New(128, time.Second)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Add(strconv.Itoa(i))
	}
	//b.StopTimer()
	//list.recalculate()
	//_, out := list.List()
	//b.Logf("top %+v", out)
}
