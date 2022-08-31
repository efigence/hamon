package stats

import (
	"github.com/efigence/go-haproxy"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSortSlowlog(t *testing.T) {
	l := []haproxy.HTTPRequest{
		{TS: 100,
			TotalDurationMs: 4000,
			RequestPath:     "/overtime",
		},
		{TS: 200,
			TotalDurationMs: 3000,
			RequestPath:     "/overtime",
		},
		{TS: 500,
			TotalDurationMs: 5000,
			RequestPath:     "/1",
		},
		{TS: 300,
			TotalDurationMs: 1000,
			RequestPath:     "/3",
		},
		{TS: 400,
			TotalDurationMs: 2000,
			RequestPath:     "/2",
		},
		{TS: 120,
			TotalDurationMs: 7000,
			RequestPath:     "/overtime",
		},
	}
	sortSlowlog(l, time.Unix(0, 250000))
	assert.EqualValues(t, "/1", l[0].RequestPath)
	assert.EqualValues(t, "/2", l[1].RequestPath)
	assert.EqualValues(t, "/3", l[2].RequestPath)
	assert.EqualValues(t, "/overtime", l[3].RequestPath)
	assert.EqualValues(t, "/overtime", l[4].RequestPath)
	assert.EqualValues(t, "/overtime", l[5].RequestPath)
}
