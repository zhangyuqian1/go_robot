package gamecase

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

func PVP(ctx context.Context) {
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
	r.Client.TransmitTarget = make(chan module.TransmitTarget, 1)
	r.CChaosBattleFieldEnter()
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
	select {
	case <-ctx.Done():
		return
	default:
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
	logger.Infof("robot %s connecting to cross-server gateway %s", r.PTName, address)
	if err := r.CreateWebSocketConn(); err != nil {
		logger.Errorf("robot %s cross-server connection failed: %v", r.PTName, err)
		return
	}

	// 21000
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
	if !waitLoginScene(ctx, r.Client.LoginSceneReady) {
		logger.Errorf("robot %s cross-server login did not enter scene", r.PTName)
		return
	}
	logger.Infof("robot %s cross-server login completed: %s", r.PTName, address)
	global.LoginRendezvous.Done(-1)
	rendezvousPending = false
	global.LoginRendezvous.Wait()
	<-ctx.Done()
	logger.Debug("登出")
}

func crossServerWebSocketURL(current string, target module.TransmitTarget) (string, error) {
	address, err := url.Parse(current)
	if err != nil {
		return "", err
	}
	if (address.Scheme != "ws" && address.Scheme != "wss") || address.Host == "" {
		return "", fmt.Errorf("invalid websocket address %q", current)
	}
	if target.Addr == "" || target.Port <= 0 || target.Port > 65535 {
		return "", fmt.Errorf("invalid transmit target %s:%d", target.Addr, target.Port)
	}
	address.Host = net.JoinHostPort(target.Addr, strconv.Itoa(int(target.Port)))
	return address.String(), nil
}

func waitLoginScene(ctx context.Context, ready <-chan struct{}) bool {
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return false
	case <-ready:
		return true
	}
}
