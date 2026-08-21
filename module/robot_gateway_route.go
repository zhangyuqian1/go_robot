package module

import (
	"fmt"
	"stressTest/gameproto/defaultmsg"
	"stressTest/gameproto/property"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	defaultSys              byte = 0
	defaultActorLoginResult byte = 1
	defaultHeartbeatRequest byte = 2
	defaultClientReadyLogin byte = 7
	defaultHeartbeat        byte = 10
	defaultActorEnterScene  byte = 13

	propertySys       byte = 2
	propertyActorInit byte = 1
)

func (r *Robot) routeGatewayResponse(sys byte, msgID byte, body []byte) {
	switch {
	case r.handleLoginGatewayResponse(sys, msgID, body):
		return
	case r.handleMovementGatewayResponse(sys, msgID, body):
		return
	case r.handleBattleGatewayResponse(sys, msgID, body):
		return
	case r.handlePetGatewayResponse(sys, msgID, body):
		return
	case r.handleEquipGatewayResponse(sys, msgID, body):
		return
	case r.handleGMGatewayResponse(sys, msgID, body):
		return
	case r.handleArenaGatewayResponse(sys, msgID, body):
		return
	case r.handleSkillGatewayResponse(sys, msgID, body):
		return
	case r.handleFriendGatewayResponse(sys, msgID, body):
		return
	case sys == defaultSys && msgID == defaultActorLoginResult:
		r.handleActorLoginResult(body)
	case sys == propertySys && msgID == propertyActorInit:
		r.handleActorInit(body)
	case sys == defaultSys && msgID == defaultHeartbeat:
		r.handleHeartbeat(body)
	case sys == defaultSys && msgID == defaultActorEnterScene:
		r.handleActorEnterScene(body)
	default:
		logger.Debug(fmt.Sprintf("unknown gateway response sys:%d msgId:%d account:%s", sys, msgID, r.PTName))
	}
}

func (r *Robot) handleActorLoginResult(body []byte) {
	var resp defaultmsg.SActorLoginResult_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		r.AsyncFire(loginRequestActorLogin, 0, STEP_FAIL, 0)
		return
	}
	r.Client.LastActorLoginResult = resp.GetResult()
	if resp.GetResult() != 0 {
		r.AsyncFire(loginRequestActorLogin, int64(len(body)), STEP_FAIL, resp.GetResult())
		return
	}
}

func (r *Robot) handleActorInit(body []byte) {
	var resp property.SActorInit_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_actor_init_1 error: %v", err)
		return
	}
	r.Client.Handle = resp.GetHandle()
	r.Client.ActorID = resp.GetId()
	r.Client.RawServerID = resp.GetRawSrv()
	r.Client.ActorName = resp.GetName()
	r.Client.CurrentServerID = resp.GetCurSrv()
	r.Client.MapData = copyMapData(resp.GetMapData())
}

func (r *Robot) handleHeartbeat(body []byte) {
	var resp defaultmsg.SHeartbeat_10
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_heartbeat_10 error: %v", err)
		return
	}
	r.Client.LastHeartbeatTick = resp.GetTick()
}

func (r *Robot) handleActorEnterScene(body []byte) {
	var resp defaultmsg.SActorEnterScene_13
	if err := proto.Unmarshal(body, &resp); err != nil {
		r.AsyncFire(loginRequestEnterScene, 0, STEP_FAIL, 0)
		return
	}
	r.Client.SceneID = resp.GetSid()
	r.Client.FubenID = resp.GetFid()
	r.Client.SceneType = resp.GetType()
	r.Client.PX = resp.GetPx()
	r.Client.PY = resp.GetPy()
	r.cacheEntityPosition(r.Client.Handle, int32(resp.GetPx()), int32(resp.GetPy()), 0)
	r.Client.LoginDone = true
	r.Client.LoginStarted = false
	r.StartHeartBeat()
	r.AsyncFire(loginRequestReadyLogin, int64(len(body)), STEP_SUCC, 0)
	r.AsyncFire(loginRequestActorLogin, int64(len(body)), STEP_SUCC, 0)
	r.AsyncFire(loginRequestEnterScene, int64(len(body)), STEP_SUCC, 0)
}

func copyMapData(source map[int32]int64) map[int32]int64 {
	if len(source) == 0 {
		return nil
	}
	copied := make(map[int32]int64, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}
