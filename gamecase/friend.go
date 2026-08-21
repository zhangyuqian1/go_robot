package gamecase

import (
	"context"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

func Friend(ctx context.Context) {
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
			if global.FriendRate == nil || global.FriendRate.Allow() {
				friendOnce(r)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func friendOnce(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	if global.FriendAid <= 0 {
		return
	}
	// 1 = SocialRelation_SOCIAL_FRIEND，加好友
	r.CAddSocial(16783890, 1)
}
