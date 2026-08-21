package gamecase

import (
	"context"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

func Pet(ctx context.Context) {
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	RobotInit(r)

	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.PetRate == nil || global.PetRate.Allow() {
				petOnce(r)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func petOnce(r *module.Robot) {
	if r == nil || r.Client == nil || !timeReady(time.Now(), r.Client.PetCooldownUntil) {
		return
	}
	seriesID := r.Client.SelectedPetSeriesID
	if seriesID == 0 {
		seriesID = firstPetSeriesID(r)
	}
	if seriesID == 0 {
		return
	}
	r.CPetBattle(seriesID, global.PetBattleIndex)
}

func firstPetSeriesID(r *module.Robot) int32 {
	if r == nil || r.Client == nil {
		return 0
	}
	for seriesID := range r.Client.Pets {
		return seriesID
	}
	return 0
}
