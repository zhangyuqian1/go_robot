package gamecase

import (
	"context"
	"fmt"
	"math/rand"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

func TeamDungeon(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	r.Client.LoginSceneReady = make(chan struct{}, 1)
	RobotInit(r)
	if !r.Client.IsWebConnected {
		return
	}
	rendezvousPending := true
	defer func() {
		if rendezvousPending {
			global.LoginRendezvous.Done(-1)
		}
	}()
	if !waitLoginScene(ctx, r.Client.LoginSceneReady) {
		logger.Errorf("robot %s local login did not enter scene", r.PTName)
		return
	}
	accountID, actorID := r.Client.AccountID, r.Client.ActorID
	r.Client.TeamDungeonMatchReady = make(chan int32, 1)
	r.Client.TransmitTarget = make(chan module.TransmitTarget, 1)
	r.CSendChatMsg(0, "@set 10187 35")
	matchID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(3) + 4
	r.CTeamDungeonMatch(int32(matchID))
	matchTimer := time.NewTimer(30 * time.Second)
	defer matchTimer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-matchTimer.C:
		logger.Errorf("robot %s timed out waiting for s_team_dungeon_match_success_14", r.PTName)
		return
	case matchID := <-r.Client.TeamDungeonMatchReady:
		if ctx.Err() != nil {
			return
		}
		r.CTeamDungeonReplay(matchID, 1)
	}
	var target module.TransmitTarget
	transmitTimer := time.NewTimer(30 * time.Second)
	defer transmitTimer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-transmitTimer.C:
		logger.Errorf("robot %s timed out waiting for s_start_transmit_5", r.PTName)
		return
	case target = <-r.Client.TransmitTarget:
	}
	address, err := crossServerWebSocketURL(r.Client.WebSocketSvr, target)
	if err != nil {
		logger.Errorf("robot %s invalid cross-server route: %v", r.PTName, err)
		return
	}
	r.CReqClientKickSelf()
	r.Close()
	if ctx.Err() != nil {
		return
	}
	r.Client.WebSocketSvr = address
	r.Client.LoginDone = false
	r.Client.LoginStarted = false
	r.Client.Handle = 0
	r.Client.AccountID = 0
	r.Client.ActorID = 0
	r.Client.LastLoginCode = 0
	r.Client.LastActorLoginResult = 0
	r.Client.TransmitTarget = nil
	r.Client.TeamDungeonMatchReady = nil
	r.Client.LoginSceneReady = make(chan struct{}, 1)
	r.TickStopSignal = make(chan struct{})
	r.ConnStopSignal = make(chan struct{})
	r.AsyncMonitSignal = make(chan struct{})
	for global.LoginRate != nil && !global.LoginRate.Allow() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Millisecond):
		}
	}
	logger.Infof("robot %s connecting to team dungeon gateway %s", r.PTName, address)
	if err := r.CreateWebSocketConn(); err != nil {
		logger.Errorf("robot %s cross-server connection failed: %v", r.PTName, err)
		return
	}

	//r.CReqCheckPassword(120, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 120, 120, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	// 21001
	r.CReqCheckPassword(121, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 121, 121, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	//21003
	//r.CReqCheckPassword(125, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 125, 125, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	//21004
	//r.CReqCheckPassword(126, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 126, 126, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	if !r.Client.IsWebConnected || r.Client.LastLoginCode != 0 || r.Client.AccountID == 0 || r.Client.ActorID == 0 {
		logger.Errorf("robot %s cross-server authentication failed, code=%d account=%d actor=%d", r.PTName, r.Client.LastLoginCode, r.Client.AccountID, r.Client.ActorID)
		return
	}
	if r.Client.AccountID != accountID {
		logger.Errorf("robot %s cross-server account changed: %d -> %d", r.PTName, accountID, r.Client.AccountID)
		return
	}
	r.CReqClientReadyLogin(actorID)
	r.CReqActorLogin(accountID, actorID)
	r.CSendChatMsg(0, "@superman")
	if !waitLoginScene(ctx, r.Client.LoginSceneReady) {
		logger.Errorf("robot %s cross-server login did not enter team dungeon scene", r.PTName)
		return
	}
	logger.Infof("robot %s entered cross-server team dungeon: %s", r.PTName, address)
	global.LoginRendezvous.Done(-1)
	rendezvousPending = false
	global.LoginRendezvous.Wait()
	<-ctx.Done()
	logger.Debug("登出")
}
