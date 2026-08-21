package module

import (
	"stressTest/gameproto/arena"
	"stressTest/global"
	"stressTest/logger"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	arenaSys          byte = 101
	arenaInfo         byte = 1
	arenaSelectList   byte = 2
	arenaFight        byte = 4
	arenaBattleResult byte = 5
	arenaEnterFuben   byte = 6

	arenaInfoRequestName        = "CArenaInfo"
	arenaListRequestName        = "CArenaPkList"
	arenaFightRequestName       = "CArenaFight"
	arenaRequestTimeout   int64 = 10
)

func (r *Robot) CArenaInfo() {
	r.AsyncBegin(arenaInfoRequestName, arenaRequestTimeout, false)
	if err := r.sendGatewayRequest(arenaInfoRequestName, arenaSys, arenaInfo, &arena.CArenaInfo_1{}, false); err != nil {
		r.Mark.Remove(arenaInfoRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(arenaInfoRequestName)
}

func (r *Robot) CArenaPkList() {
	r.AsyncBegin(arenaListRequestName, arenaRequestTimeout, false)
	if err := r.sendGatewayRequest(arenaListRequestName, arenaSys, arenaSelectList, &arena.CArenaPkList_2{}, false); err != nil {
		r.Mark.Remove(arenaListRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(arenaListRequestName)
}

func (r *Robot) CArenaFight(index int32) {
	if index <= 0 {
		index = 1
	}
	req := &arena.CArenaFight_4{Idx: index}
	r.AsyncBegin(arenaFightRequestName, arenaRequestTimeout, false)
	if err := r.sendGatewayRequest(arenaFightRequestName, arenaSys, arenaFight, req, false); err != nil {
		r.Mark.Remove(arenaFightRequestName)
		logger.Error(err.Error())
		return
	}
	if r.Client != nil {
		r.Client.ArenaFightPending = false
		r.Client.ArenaInFight = true
	}
	r.AsyncWait(arenaFightRequestName)
}

func (r *Robot) handleArenaGatewayResponse(sys byte, msgID byte, body []byte) bool {
	switch sys {
	case arenaSys:
		switch msgID {
		case arenaInfo:
			r.handleArenaInit(body)
		case arenaSelectList:
			r.handleArenaSelectList(body)
		case arenaBattleResult:
			r.handleArenaBattleResult(body)
		case arenaEnterFuben:
			r.handleArenaEnterFuben(body)
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func (r *Robot) handleArenaInit(body []byte) {
	var resp arena.SArenaInit_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_arena_init_1 error: %v", err)
		r.AsyncFire(arenaInfoRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil {
		r.Client.ArenaRank = resp.GetRank()
		r.Client.ArenaScore = resp.GetScore()
	}
	r.AsyncFire(arenaInfoRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleArenaSelectList(body []byte) {
	var resp arena.SArenaSelectList_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_arena_select_list_2 error: %v", err)
		r.AsyncFire(arenaListRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil {
		r.Client.ArenaEnemies = append([]*arena.ArenaEnemyInfo(nil), resp.GetItem()...)
		r.Client.ArenaFightPending = len(resp.GetItem()) > 0
		if r.Client.ArenaFightPending {
			r.Client.ArenaFightReadyAt = time.Now().Add(global.ArenaFightDelay)
		}
	}
	r.AsyncFire(arenaListRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleArenaEnterFuben(body []byte) {
	var resp arena.SArenaEnterFuben_6
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_arena_enter_fuben_6 error: %v", err)
		r.AsyncFire(arenaFightRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil {
		r.Client.ArenaInFight = true
		r.Client.ArenaFightPending = false
	}
	r.AsyncFire(arenaFightRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleArenaBattleResult(body []byte) {
	var resp arena.SArenaBattleResult_5
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_arena_battle_result_5 error: %v", err)
		r.AsyncFire(arenaFightRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil {
		r.Client.ArenaLastWin = resp.GetWin()
		r.Client.ArenaScore = resp.GetScore()
		r.Client.ArenaRank = resp.GetRank()
		r.Client.ArenaFightPending = false
		r.Client.ArenaFightReadyAt = time.Time{}
		r.Client.ArenaInFight = false
		r.Client.ArenaEnemies = nil
		r.Client.ArenaCooldownUntil = time.Now().Add(global.ArenaCooldown)
	}
	r.AsyncFire(arenaFightRequestName, int64(len(body)), STEP_SUCC, 0)
}
