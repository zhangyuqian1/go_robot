package module

import (
	"stressTest/gameproto/defaultmsg"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

type TransmitTarget struct {
	Addr string
	Port int32
}

func (r *Robot) handleStartTransmit(body []byte) {
	var resp defaultmsg.SStartTransmit_5
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_start_transmit_5: %v", err)
		r.AsyncFire(chaosBattleFieldEnterRequestName, 0, STEP_FAIL, 0)
		r.AsyncFire(teamDungeonReplayRequestName, 0, STEP_FAIL, 0)
		r.AsyncFire(dragonCityEnterRequestName, 0, STEP_FAIL, 0)
		return
	}
	if resp.GetAddr() == "" || resp.GetPort() <= 0 || resp.GetPort() > 65535 {
		logger.Errorf("invalid transmit target: %s:%d", resp.GetAddr(), resp.GetPort())
		r.AsyncFire(chaosBattleFieldEnterRequestName, int64(len(body)), STEP_FAIL, 0)
		r.AsyncFire(teamDungeonReplayRequestName, int64(len(body)), STEP_FAIL, 0)
		r.AsyncFire(dragonCityEnterRequestName, int64(len(body)), STEP_FAIL, 0)
		return
	}
	if r.Client.TransmitTarget != nil {
		select {
		case r.Client.TransmitTarget <- TransmitTarget{Addr: resp.GetAddr(), Port: resp.GetPort()}:
		default:
		}
	}
	r.AsyncFire(chaosBattleFieldEnterRequestName, int64(len(body)), STEP_SUCC, 0)
	r.AsyncFire(teamDungeonReplayRequestName, int64(len(body)), STEP_SUCC, 0)
	r.AsyncFire(dragonCityEnterRequestName, int64(len(body)), STEP_SUCC, 0)
}
