package module

import (
	"testing"

	"stressTest/gameproto/common"
)

func TestProtocolRequestsRemoveMarkOnSendFailure(t *testing.T) {
	series := &common.Series{Series: 1, Time: 2, Server: 3}
	tests := []struct {
		name        string
		requestName string
		send        func(*Robot)
	}{
		{name: "actor set dir", requestName: movementSetDirRequestName, send: func(r *Robot) { r.CActorSetDir(90) }},
		{name: "actor move", requestName: movementRequestName, send: func(r *Robot) { r.CActorMove(1, 2, 3, 4, 5) }},
		{name: "normal attack", requestName: battleRequestName, send: func(r *Robot) { r.CNormalAttack(1001, 1) }},
		{name: "start use skill", requestName: startUseSkillRequestName, send: func(r *Robot) { r.CStartUseSkill(1001, 4209, 90, 0) }},
		{name: "skill frame proxy", requestName: skillFrameProxyRequestName, send: func(r *Robot) {
			r.CSkillFrameProxy(SkillFrameProxyRequest{
				SkillID:           4209,
				SkillLevel:        1,
				CasterHandler:     "1001",
				TargetHandler:     "2001",
				DirectionX:        1,
				TargetX:           11,
				TargetY:           22,
				WeaponSlotIndex:   0,
				EffectTriggerRate: 1,
			})
		}},
		{name: "damage entity", requestName: damageEntityRequestName, send: func(r *Robot) { r.CDamageEntity(2001, 4209, 42090, 0) }},
		{name: "minimap boss info", requestName: minimapBossInfoRequestName, send: func(r *Robot) { r.CMinimapBossInfo(3) }},
		{name: "apply relive", requestName: reliveRequestName, send: func(r *Robot) { r.CApplyRelive(1) }},
		{name: "pet battle", requestName: petBattleRequestName, send: func(r *Robot) { r.CPetBattle(2001, 1) }},
		{name: "pet summon", requestName: petSummonRequestName, send: func(r *Robot) { r.CPetSummon(1) }},
		{name: "get bag items", requestName: equipBagRequestName, send: func(r *Robot) { r.CGetBagItems(1) }},
		{name: "take on equip", requestName: equipTakeOnRequestName, send: func(r *Robot) { r.CTakeOnEquip(1, series) }},
		{name: "recover bag item", requestName: equipSellRequestName, send: func(r *Robot) { r.CRecBagItem([]*common.UserItem{{Series: series}}) }},
		{name: "arena info", requestName: arenaInfoRequestName, send: func(r *Robot) { r.CArenaInfo() }},
		{name: "arena list", requestName: arenaListRequestName, send: func(r *Robot) { r.CArenaPkList() }},
		{name: "arena fight", requestName: arenaFightRequestName, send: func(r *Robot) { r.CArenaFight(1) }},
		{name: "gm chat", requestName: gmSendChatRequestName, send: func(r *Robot) { r.CSendChatMsg(1, "gm") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRobot()
			tt.send(r)
			if r.Mark.Has(tt.requestName) {
				t.Fatalf("%s mark should be removed after send failure", tt.requestName)
			}
		})
	}
}
