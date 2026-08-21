package gamecase

import (
	"context"
	"stressTest/global"
	"stressTest/logger"
	"time"
)

// Pray 循环活动祈福抽卡压测案例
func Pray(ctx context.Context) {
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
			if global.PrayRate == nil || global.PrayRate.Allow() {
				r.CPraySearch(10)
				r.CPraySearch(10)
				r.CPraySearch(10)
				r.CPraySearch(10)
				r.CPraySearch(10)
				r.CPraySearch(10)
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}
