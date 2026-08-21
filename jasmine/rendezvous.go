package jasmine

import (
	"sync/atomic"
	"time"
)

type Rendezvous interface {
	Done()
	Wait()
}

var RendezvousPoints = make(map[string]*RendezvousPoint)

type RendezvousPoint struct {
	threshold        int64
	currentThreshold int64
}

func NewRendezvousPoint(name string, threshold int64) *RendezvousPoint {
	_, ok := RendezvousPoints[name]
	if ok {
		if RendezvousPoints[name].currentThreshold > 0 {
			return RendezvousPoints[name]
		}
	}
	rendezvousPoint := &RendezvousPoint{
		threshold:        threshold,
		currentThreshold: threshold,
	}
	return rendezvousPoint
}

func (rendezvousPoint *RendezvousPoint) Done(delta int64) {
	atomic.AddInt64(&rendezvousPoint.currentThreshold, delta)
}

func (rendezvousPoint *RendezvousPoint) Wait() {
	for {
		if rendezvousPoint.currentThreshold <= 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
