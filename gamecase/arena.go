package gamecase

import (
	"context"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

func Arena(ctx context.Context) {
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
			if global.ArenaRate == nil || global.ArenaRate.Allow() {
				arenaOnce(r)
			}
			time.Sleep(global.ArenaListInterval)
		}
	}
}

func initArena(r *module.Robot) {
	if r == nil {
		return
	}
	r.CArenaInfo()
	if global.ArenaGMCommand != "" {
		r.CSendChatMsg(1, global.ArenaGMCommand)
	}
}

func arenaOnce(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	now := time.Now()
	if !timeReady(now, r.Client.ArenaCooldownUntil) || r.Client.ArenaInFight {
		return
	}
	if r.Client.ArenaFightPending && len(r.Client.ArenaEnemies) > 0 {
		if !timeReady(now, r.Client.ArenaFightReadyAt) {
			return
		}
		r.CArenaFight(global.ArenaChallengeIndex)
		return
	}
	r.CArenaPkList()
}
