package module

import (
	"fmt"
	"stressTest/gameproto/defaultmsg"
	"stressTest/global"
	"stressTest/logger"
	"sync"
	"time"
)

func (r *Robot) StartHeartBeat() {
	if r == nil || r.Client == nil || !r.Client.LoginDone || !r.Client.IsWebConnected || r.Ticking {
		return
	}
	if r.TickStopSignal == nil {
		r.TickStopSignal = make(chan struct{})
	} else {
		select {
		case <-r.TickStopSignal:
			r.TickStopSignal = make(chan struct{})
		default:
		}
	}
	r.TickStopOnce = sync.Once{}
	r.Ticking = true
	r.Client.PingSending = true
	r.TickSyncGroup.Add(1)
	go func() {
		defer r.TickSyncGroup.Done()
		defer func() {
			r.Ticking = false
			if r.Client != nil {
				r.Client.PingSending = false
			}
			logger.Info(fmt.Sprintf("Client %s HeartBeat goroutine exited", r.PTName))
		}()
		r.HeartBeat()
	}()
}

func (r *Robot) HeartBeat() {
	ticker := time.NewTicker(heartbeatInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if r.Client == nil || !r.Client.PingSending || !r.Client.IsWebConnected {
				logger.Info(fmt.Sprintf("Client %s HeartBeat Break", r.PTName))
				return
			}
			if err := r.HeartBeatReq(); err != nil {
				logger.Error(fmt.Sprintf("Client %s HeartBeat request failed: %s", r.PTName, err))
				r.CloseSocket()
				return
			}
		case <-r.TickStopSignal:
			logger.Info(fmt.Sprintf("Client %s Recv Stop Heart Beat", r.PTName))
			return
		}
	}
}

func (r *Robot) HeartBeatReq() error {
	return r.sendGatewayRequest("CHeartbeat", defaultSys, defaultHeartbeatRequest, buildHeartbeatRequest(time.Now()), false)
}

func (r *Robot) StopHeartBeat() {
	if r.TickStopSignal != nil {
		if r.Ticking {
			r.signalHeartbeatStop()
		}
		r.TickSyncGroup.Wait()
		logger.Info(fmt.Sprintf("Client %s Stop HeartBeat", r.PTName))
	} else {
		logger.Error(fmt.Sprintf("Client %s TickStopSignal is nil, can't stop HeartBeat", r.PTName))
	}
	r.TickStopSignal = nil
}

func (r *Robot) signalHeartbeatStop() {
	if r.TickStopSignal == nil {
		return
	}
	r.TickStopOnce.Do(func() {
		close(r.TickStopSignal)
	})
}

func buildHeartbeatRequest(now time.Time) *defaultmsg.CHeartbeat_2 {
	return &defaultmsg.CHeartbeat_2{Tick: int32(now.Unix())}
}

func heartbeatInterval() time.Duration {
	if global.HeartbeatInterval > 0 {
		return global.HeartbeatInterval
	}
	return 10 * time.Second
}
