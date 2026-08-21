package gamecase

import (
	"context"
	"stressTest/global"
	"stressTest/logger"
	"time"
)

func PostLoginMixed(ctx context.Context) {
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	RobotInit(r)

	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()
	initArena(r)

	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.MoveRate == nil || global.MoveRate.Allow() {
				actorMove(r)
			}
			if global.BattleRate == nil || global.BattleRate.Allow() {
				battleOnce(r)
			}
			if global.PetRate == nil || global.PetRate.Allow() {
				petOnce(r)
			}
			if global.EquipRate == nil || global.EquipRate.Allow() {
				equipOnce(r)
			}
			if global.ArenaRate == nil || global.ArenaRate.Allow() {
				arenaOnce(r)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func timeReady(now time.Time, target time.Time) bool {
	return target.IsZero() || now.After(target) || now.Equal(target)
}
