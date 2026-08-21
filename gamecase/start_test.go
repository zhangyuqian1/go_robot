package gamecase

import (
	"testing"
	"time"

	"stressTest/gameproto/common"
)

func TestWaitGiftSeriesReturnsWhenCached(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.GiftSeries = &common.Series{Time: 1, Series: 2, Server: 3}

	start := time.Now()
	waitGiftSeries(r)

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("waitGiftSeries took too long: %v", elapsed)
	}
	if r.Client.GiftSeries.GetSeries() != 2 {
		t.Fatalf("GiftSeries = %#v, want series=2", r.Client.GiftSeries)
	}
}

func TestWaitGiftSeriesTimesOut(t *testing.T) {
	r := newPostLoginTestRobot()

	start := time.Now()
	waitGiftSeries(r)

	if elapsed := time.Since(start); elapsed < 4*time.Second {
		t.Fatalf("waitGiftSeries should wait until timeout, took %v", elapsed)
	}
	if r.Client.GiftSeries != nil {
		t.Fatalf("GiftSeries should be nil, got %#v", r.Client.GiftSeries)
	}
}
