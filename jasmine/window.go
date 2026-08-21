package jasmine

import "time"

type LocalWindow struct {
	start int64
	count int64
}

func NewLocalWindow() (*LocalWindow, StopFunc) {
	return &LocalWindow{}, func() {}
}

func (w *LocalWindow) Start() time.Time {
	return time.Unix(0, w.start)
}

func (w *LocalWindow) Count() int64 {
	return w.count
}

func (w *LocalWindow) AddCount(n int64) {
	w.count += n
}

func (w *LocalWindow) Reset(s time.Time, c int64) {
	w.start = s.UnixNano()
	w.count = c
}

func (w *LocalWindow) Sync(now time.Time) {}

type (
	SyncRequest struct {
		Key     string
		Start   int64
		Count   int64
		Changes int64
	}

	SyncResponse struct {
		// Whether the synchronization succeeds.
		OK    bool
		Start int64
		// The changes accumulated by the local limiter.
		Changes int64
		// The total changes accumulated by all the other limiters.
		OtherChanges int64
	}

	MakeFunc   func() SyncRequest
	HandleFunc func(SyncResponse)
)
