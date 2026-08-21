package gamecase

import (
	"context"
	"stressTest/global"
	"stressTest/logger"
	"time"
)

// Draw 技能抽卡压测案例
func Draw(ctx context.Context) {
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	RobotInit(r)

	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()
	r.CSendChatMsg(0, "@additem 9 9999")
	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.DrawRate == nil || global.DrawRate.Allow() {
				r.CDraw(1, false)
				r.CDraw(1, false)
				r.CDraw(1, false)
				r.CDraw(1, false)
				r.CDraw(1, false)
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}
