package api

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
)

type sequenceDoer struct {
	codes  []int
	bodies []string
}

func (d *sequenceDoer) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	d.bodies = append(d.bodies, string(body))
	code := d.codes[len(d.bodies)-1]
	header := http.Header{}
	header.Set("x-ratelimit-reset", strconv.FormatInt(time.Unix(1000, 0).Add(10*time.Second).Unix(), 10))
	return &http.Response{StatusCode: code, Header: header, Body: io.NopCloser(bytes.NewReader(nil))}, nil
}

func TestRetryDoer(t *testing.T) {
	var waits []time.Duration
	newDoer := func(inner HTTPDoer) *retryDoer {
		return &retryDoer{
			inner: inner,
			sleep: func(d time.Duration) { waits = append(waits, d) },
			now:   func() time.Time { return time.Unix(1000, 0) },
		}
	}
	request := func() *http.Request {
		req, _ := http.NewRequest("PATCH", "http://example.test", bytes.NewBufferString("body"))
		return req
	}

	t.Run("解除まで待って同じ本文で再送する", func(t *testing.T) {
		waits = nil
		inner := &sequenceDoer{codes: []int{429, 200}}
		resp, err := newDoer(inner).Do(request())
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("resp=%v err=%v", resp, err)
		}
		if inner.bodies[1] != "body" {
			t.Errorf("再送の本文 = %q", inner.bodies[1])
		}
		if len(waits) != 1 || waits[0] != 11*time.Second {
			t.Errorf("waits = %v", waits)
		}
	})

	t.Run("上限回数で諦める", func(t *testing.T) {
		inner := &sequenceDoer{codes: []int{429, 429, 429, 429}}
		resp, err := newDoer(inner).Do(request())
		if err != nil || resp.StatusCode != 429 || len(inner.bodies) != 4 {
			t.Errorf("resp=%v err=%v calls=%d", resp, err, len(inner.bodies))
		}
	})
}
