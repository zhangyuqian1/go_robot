package gamecase

import (
	"context"
	"stressTest/global"
	"stressTest/logger"
	"time"
)

func LoginOnline(ctx context.Context) {
	r := StartFunc()
	RobotInit(r)

	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()
	//
	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.LoginRate == nil || global.LoginRate.Allow() {
				// r.CReqCheckPassword(1001, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 1001, 1001, "10.200.255.60", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
				// if r.Client.ActorID == 0 {
				// 	r.CReqCreateActor(1, 1, fmt.Sprintf("L%d", r.Index), 1, "e10adc3949ba59abbe56e057f20f883e")
				// 	r.CReqGetActorList()
				// }
				// r.CReqClientReadyLogin(r.Client.ActorID)
				// r.CReqActorLogin(r.Client.AccountID, r.Client.ActorID)
				time.Sleep(10 * time.Hour)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
