package module

import (
	"stressTest/gameproto/movement"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	movementSys                byte  = 1
	movementActorMove          byte  = 1
	movementEntityMove         byte  = 1
	movementActorSetDir        byte  = 2
	movementEntitySetDir       byte  = 2
	movementEntityFlash        byte  = 3
	movementSyncEntityPosition byte  = 4
	movementActorMoveFail      byte  = 5
	movementRequestName              = "CActorMove"
	movementSetDirRequestName        = "CActorSetDir"
	movementRequestTimeout     int64 = 10
)

func (r *Robot) CActorMove(curX int32, curY int32, destX int32, destY int32, tick int32) {
	req := &movement.CActorMove_1{
		CurX:  curX,
		CurY:  curY,
		DestX: destX,
		DestY: destY,
		Tick:  tick,
	}
	r.AsyncBegin(movementRequestName, movementRequestTimeout, true)
	if err := r.sendGatewayRequest(movementRequestName, movementSys, movementActorMove, req, true); err != nil {
		r.Mark.Remove(movementRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(movementRequestName)
}

func (r *Robot) CActorSetDir(dir int32) {
	req := &movement.CActorSetDir_2{Dir: dir}
	r.AsyncBegin(movementSetDirRequestName, movementRequestTimeout, false)
	if err := r.sendGatewayRequest(movementSetDirRequestName, movementSys, movementActorSetDir, req, false); err != nil {
		r.Mark.Remove(movementSetDirRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(movementSetDirRequestName)
}

func (r *Robot) handleMovementGatewayResponse(sys byte, msgID byte, body []byte) bool {
	if sys != movementSys {
		return false
	}
	switch msgID {
	case movementEntityMove:
		r.handleEntityMove(body)
	case movementEntitySetDir:
		r.handleEntitySetDir(body)
	case movementEntityFlash:
		r.handleEntityFlash(body)
	case movementSyncEntityPosition:
		r.handleSyncEntityPosition(body)
	case movementActorMoveFail:
		r.handleActorMoveFail(body)
	default:
		return false
	}
	return true
}

func (r *Robot) handleEntityMove(body []byte) {
	var resp movement.SEntityMove_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_entity_move_1 error: %v", err)
		r.AsyncFire(movementSetDirRequestName, 0, STEP_FAIL, 0)
		r.AsyncFire(movementRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.cacheEntityPosition(resp.GetHandle(), resp.GetDestX(), resp.GetDestY(), 0)
	if r.Client != nil && resp.GetHandle() == r.Client.Handle {
		r.Client.PX = float32(resp.GetDestX())
		r.Client.PY = float32(resp.GetDestY())
		r.Client.LastMoveTick = resp.GetTick()
		r.Client.MoveBlocked = false
		r.AsyncFire(movementSetDirRequestName, int64(len(body)), STEP_SUCC, 0)
		r.AsyncFire(movementRequestName, int64(len(body)), STEP_SUCC, 0)
	}
}

func (r *Robot) handleEntitySetDir(body []byte) {
	var resp movement.SEntitySetDir_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_entity_set_dir_2 error: %v", err)
		return
	}
	r.cacheEntityPosition(resp.GetHandle(), 0, 0, resp.GetDir())
}

func (r *Robot) handleEntityFlash(body []byte) {
	var resp movement.SEntityFlash_3
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_entity_flash_3 error: %v", err)
		return
	}
	r.cacheEntityPosition(resp.GetHandle(), resp.GetX(), resp.GetY(), resp.GetDir())
	if r.Client != nil && resp.GetHandle() == r.Client.Handle {
		r.Client.PX = float32(resp.GetX())
		r.Client.PY = float32(resp.GetY())
	}
}

func (r *Robot) handleSyncEntityPosition(body []byte) {
	var resp movement.SSyncEntityPostion_4
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_sync_entity_postion_4 error: %v", err)
		return
	}
	r.cacheEntityPosition(resp.GetHandle(), resp.GetCurX(), resp.GetCurY(), 0)
	if r.Client != nil && resp.GetHandle() == r.Client.Handle {
		r.Client.PX = float32(resp.GetCurX())
		r.Client.PY = float32(resp.GetCurY())
	}
}

func (r *Robot) handleActorMoveFail(body []byte) {
	var resp movement.SActorMoveFail_5
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_actor_move_fail_5 error: %v", err)
		r.AsyncFire(movementRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client == nil {
		return
	}
	r.Client.PX = float32(resp.GetSrvX())
	r.Client.PY = float32(resp.GetSrvY())
	r.Client.MoveBlocked = true
	r.AsyncFire(movementRequestName, int64(len(body)), STEP_FAIL, "SActorMoveFail_5")
}

func (r *Robot) cacheEntityPosition(handle int64, x int32, y int32, dir int32) {
	if r.Client == nil || handle == 0 {
		return
	}
	if r.Client.Entities == nil {
		r.Client.Entities = make(map[int64]*EntityState)
	}
	entity := r.Client.Entities[handle]
	if entity == nil {
		entity = &EntityState{Handle: handle}
		r.Client.Entities[handle] = entity
	}
	if x != 0 || y != 0 {
		entity.X = x
		entity.Y = y
	}
	if dir != 0 {
		entity.Dir = dir
	}
}
