package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	maxRetries       = 3
	defaultWait      = 60 * time.Second
	maxRateLimitWait = 16 * time.Minute // esaのレート制限は15分単位
)

// retryDoer 429を受けたらレート制限の解除まで待って再送する
type retryDoer struct {
	inner HTTPDoer
	sleep func(time.Duration)
	now   func() time.Time
}

func (d *retryDoer) Do(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		resp, err := d.inner.Do(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests || attempt >= maxRetries {
			return resp, err
		}

		wait := d.waitDuration(resp)
		resp.Body.Close()
		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}

		fmt.Fprintf(os.Stderr, "⏳ レート制限のため%d秒待機します（%d/%d回目）\n", int(wait.Seconds()), attempt+1, maxRetries)
		d.sleep(wait)
	}
}

func (d *retryDoer) waitDuration(resp *http.Response) time.Duration {
	wait := defaultWait
	if reset, err := strconv.ParseInt(resp.Header.Get("x-ratelimit-reset"), 10, 64); err == nil {
		wait = time.Unix(reset, 0).Sub(d.now()) + time.Second
	}
	if wait <= 0 {
		return time.Second
	}
	if wait > maxRateLimitWait {
		return maxRateLimitWait
	}
	return wait
}
