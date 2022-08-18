package toplist

import (
	"fmt"
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
	// the initial addition to the toplist is damped to avoid spiky behaviour
	// that happens wheyn say a bunch of HTTP/2.0 connections get recorded
	// in same sample packet but actual rate over time is lower
	assert.Equal(t, []string{"1.2.3.4", "2.2.3.4", "3.2.3.4", "4.2.3.4"}, order)
	assert.InDeltaMapValues(t, map[string]float64{
		"1.2.3.4": 2.0,
		"2.2.3.4": 1.5,
		"3.2.3.4": 1.0,
		"4.2.3.4": 0.5,
	}, out, 0.01)

	list.lastRecalc = time.Now().Add(time.Second * -1)
	list.Add("5.2.3.4")
	list.Add("5.2.3.4")
	list.Add("5.2.3.4")
	list.recalculate()
	order, out = list.List()
	assert.InDeltaMapValues(t, map[string]float64{
		"1.2.3.4": 2.0,
		"2.2.3.4": 1.5,
		"3.2.3.4": 1.0,
		"5.2.3.4": 1.5,
	}, out, 0.01)
	//	assert.Equal(t, []string{"1.2.3.4", "2.2.3.4", "3.2.3.4", "4.2.3.4"}, order)
	list.lastRecalc = time.Now().Add(time.Second * -1)
	list.Add("2.2.3.4")
	list.Add("4.2.3.4")
	list.Add("4.2.3.4")
	list.Add("4.2.3.4")
	list.recalculate()
	order, out = list.List()
	assert.InDeltaMapValues(t, map[string]float64{
		"1.2.3.4": 2.0,
		"2.2.3.4": 2.0,
		"4.2.3.4": 2.0,
		"5.2.3.4": 1.5,
	}, out, 0.01)
	assert.Equal(t, []string{"4.2.3.4", "2.2.3.4", "1.2.3.4", "5.2.3.4"}, order)

}

func BenchmarkToplist_Add(b *testing.B) {
	sizes := []int{
		32,
		128,
		512,
		2048,
	}
	multiplier := []int{
		2,
		4,
		8,
		16,
	}
	for _, size := range sizes {
		b.Run(
			fmt.Sprintf("size: %d, buffer: default", size),
			func(b *testing.B) {
				list := New(size, time.Second)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					list.Add(strconv.Itoa(i))
				}
			})
		for _, mult := range multiplier {
			bufferSize := size * mult
			b.Run(
				fmt.Sprintf("size: %d, buffer: %d", size, bufferSize),
				func(b *testing.B) {
					list := New(size, time.Second, bufferSize)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						list.Add(strconv.Itoa(i))
					}
				})
		}
	}
}
