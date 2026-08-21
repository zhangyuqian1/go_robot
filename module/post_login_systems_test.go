package module

import (
	"testing"
	"time"

	"stressTest/gameproto/arena"
	"stressTest/gameproto/bag"
	"stressTest/gameproto/chat"
	"stressTest/gameproto/common"
	"stressTest/gameproto/equip"
	"stressTest/gameproto/misc"
	"stressTest/gameproto/movement"
	"stressTest/gameproto/observer"
	"stressTest/gameproto/pet"
	"stressTest/gameproto/property"
	"stressTest/gameproto/skill"
	"stressTest/global"
)

func TestMovementRouteCachesPositionAndBlockedState(t *testing.T) {
	r := newTestRobot()
	r.Client.Handle = 42
	r.Client.PX = 10
	r.Client.PY = 20

	r.routeGatewayResponse(movementSys, movementEntityMove, mustProtoBytes(t, &movement.SEntityMove_1{
		Handle: 42,
		CurX:   10,
		CurY:   20,
		DestX:  30,
		DestY:  40,
		Tick:   99,
	}))
	if r.Client.PX != 30 || r.Client.PY != 40 || r.Client.LastMoveTick != 99 {
		t.Fatalf("self move cache = px:%f py:%f tick:%d", r.Client.PX, r.Client.PY, r.Client.LastMoveTick)
	}
	if got := r.Client.Entities[42]; got == nil || got.X != 30 || got.Y != 40 {
		t.Fatalf("entity cache missing self move: %#v", got)
	}

	r.routeGatewayResponse(movementSys, movementActorMoveFail, mustProtoBytes(t, &movement.SActorMoveFail_5{SrvX: 7, SrvY: 8}))
	if !r.Client.MoveBlocked || r.Client.PX != 7 || r.Client.PY != 8 {
		t.Fatalf("move fail cache = blocked:%v px:%f py:%f", r.Client.MoveBlocked, r.Client.PX, r.Client.PY)
	}
}

func TestBattleRouteCachesMonstersAndDeathState(t *testing.T) {
	r := newTestRobot()
	r.Client.Handle = 100

	r.routeGatewayResponse(observerSys, observerMonsterAppear, mustProtoBytes(t, &observer.SMonsterAppear_3{
		Common: &common.EntityCommon{Handle: 200, Id: 64, X: 11, Y: 22},
		CurHp:  99,
		MaxHp:  100,
		Name:   "骸骨巨龙",
	}))
	if got := r.Client.Monsters[200]; got == nil || got.Handle != 200 || got.ID != 64 || got.CurHP != 99 || got.X != 11 || got.Y != 22 || got.Name != "骸骨巨龙" {
		t.Fatalf("monster cache = %#v", got)
	}
	r.routeGatewayResponse(observerSys, observerEntityDeath, mustProtoBytes(t, &observer.SEntityDeath_8{Handle: 200}))
	if _, ok := r.Client.Monsters[200]; ok || r.Client.BattleTargetHandle != 0 {
		t.Fatalf("dead monster was not removed, target:%d monsters:%#v", r.Client.BattleTargetHandle, r.Client.Monsters)
	}

	r.routeGatewayResponse(observerSys, observerEntityDeath, mustProtoBytes(t, &observer.SEntityDeath_8{Handle: 100}))
	if !r.Client.SelfDead {
		t.Fatalf("self death was not cached")
	}
	r.routeGatewayResponse(propertySys, propertyCloseReliveDialog, mustProtoBytes(t, &property.SCloseReliveDialog_8{}))
	if r.Client.SelfDead {
		t.Fatalf("self death was not cleared after relive close")
	}
}
func TestBattleRouteCachesBossInfoAndStartSkillResult(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin(minimapBossInfoRequestName, battleRequestTimeout, false)

	r.routeGatewayResponse(miscSys, miscMinimapBossInfoResult, mustProtoBytes(t, &misc.SMinimapBossInfo_8{
		Sid: 3,
		Bosses: []*misc.SMinimapBossInfo_8Info{
			{Id: 64, X: 20, Y: 97, Relive: 0},
			{Id: 63, X: 17, Y: 42, Relive: 5},
		},
	}))

	if r.Mark.Has(minimapBossInfoRequestName) {
		t.Fatalf("%s should be fired by s_minimap_boss_info_8", minimapBossInfoRequestName)
	}
	if got := r.Client.BattleBosses[64]; got == nil || got.ID != 64 || got.X != 20 || got.Y != 97 || got.Relive != 0 {
		t.Fatalf("boss 64 cache = %#v", got)
	}
	if got := r.Client.BattleBosses[63]; got == nil || got.Relive != 5 {
		t.Fatalf("boss 63 cache = %#v", got)
	}

	r.AsyncBegin(startUseSkillRequestName, battleRequestTimeout, false)
	r.routeGatewayResponse(skillSys, skillStartUseSkillResult, mustProtoBytes(t, &skill.SStartUseSkill_2{SkillId: 4209, Result: 0, RandVal: 7038}))
	if r.Mark.Has(startUseSkillRequestName) {
		t.Fatalf("%s should be fired by s_start_use_skill_2", startUseSkillRequestName)
	}
	if r.Client.BattleSkillRandVal != 7038 {
		t.Fatalf("BattleSkillRandVal = %d, want 7038", r.Client.BattleSkillRandVal)
	}
}

func TestPetRouteCachesAndSchedulesBattle(t *testing.T) {
	r := newTestRobot()
	global.PetCooldown = 50 * time.Millisecond

	r.routeGatewayResponse(petSys, petUpdate, mustProtoBytes(t, &pet.SPetUpdate_1{
		Type: 1,
		PetInfo: []*pet.PetInfo{{
			SeriseId: 700,
			Id:       701,
			Level:    2,
		}},
	}))
	if got := r.Client.Pets[700]; got == nil || got.ID != 701 || got.Level != 2 {
		t.Fatalf("pet cache = %#v", got)
	}
	if !r.Client.PetBattlePending || r.Client.SelectedPetSeriesID != 700 {
		t.Fatalf("pet battle state = pending:%v selected:%d", r.Client.PetBattlePending, r.Client.SelectedPetSeriesID)
	}

	r.routeGatewayResponse(petSys, petRemove, mustProtoBytes(t, &pet.SPetRemove_2{RemoveSerise: []int32{700}}))
	if _, ok := r.Client.Pets[700]; ok || r.Client.PetCooldownUntil.IsZero() || !r.Client.PetBattlePending {
		t.Fatalf("pet remove state = pets:%#v cooldown:%s pending:%v", r.Client.Pets, r.Client.PetCooldownUntil, r.Client.PetBattlePending)
	}
}

func TestEquipAndBagRoutesCacheItemsAndDecideAction(t *testing.T) {
	r := newTestRobot()
	item1 := &common.UserItem{Series: &common.Series{Time: 1, Series: 2, Server: 3}, Id: 1001, Count: 1}
	item2 := &common.UserItem{Series: &common.Series{Time: 4, Series: 5, Server: 6}, Id: 1002, Count: 1}

	r.routeGatewayResponse(equipSys, equipList, mustProtoBytes(t, &equip.SEquipList_1{Items: []*common.UserItem{item1}}))
	if got := r.Client.PendingEquipItems; len(got) != 1 || got[0].GetId() != 1001 {
		t.Fatalf("single equip cache = %#v", got)
	}
	r.routeGatewayResponse(bagSys, bagPushItem, mustProtoBytes(t, &bag.SPushBagItem_1{Type: 1, List: []*common.UserItem{item1, item2}}))
	if got := r.Client.PendingEquipItems; len(got) != 2 {
		t.Fatalf("multi equip cache len = %d", len(got))
	}
}

func TestArenaRouteCachesListAndResult(t *testing.T) {
	r := newTestRobot()

	r.routeGatewayResponse(arenaSys, arenaSelectList, mustProtoBytes(t, &arena.SArenaSelectList_2{
		Item: []*arena.ArenaEnemyInfo{{Id: 9, Name: "enemy", Score: 100}},
	}))
	if len(r.Client.ArenaEnemies) != 1 || r.Client.ArenaEnemies[0].GetId() != 9 || !r.Client.ArenaFightPending || r.Client.ArenaFightReadyAt.IsZero() {
		t.Fatalf("arena list cache = enemies:%#v pending:%v ready:%s", r.Client.ArenaEnemies, r.Client.ArenaFightPending, r.Client.ArenaFightReadyAt)
	}

	r.routeGatewayResponse(arenaSys, arenaBattleResult, mustProtoBytes(t, &arena.SArenaBattleResult_5{Win: 1, Score: 200, Rank: 3}))
	if r.Client.ArenaFightPending || r.Client.ArenaInFight || r.Client.ArenaLastWin != 1 || r.Client.ArenaScore != 200 || r.Client.ArenaRank != 3 || r.Client.ArenaCooldownUntil.IsZero() {
		t.Fatalf("arena result cache = pending:%v inFight:%v win:%d score:%d rank:%d cooldown:%s", r.Client.ArenaFightPending, r.Client.ArenaInFight, r.Client.ArenaLastWin, r.Client.ArenaScore, r.Client.ArenaRank, r.Client.ArenaCooldownUntil)
	}
}

func TestGMRouteFiresChatAndTipResponses(t *testing.T) {
	tests := []struct {
		name  string
		msgID byte
		body  []byte
	}{
		{name: "chat send result", msgID: gmSendChatResult},
		{name: "tip message", msgID: gmTipMsgWithStr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRobot()
			r.AsyncBegin(gmSendChatRequestName, gmRequestTimeout, false)
			body := tt.body
			if body == nil {
				switch tt.msgID {
				case gmTipMsgWithStr:
					body = mustProtoBytes(t, &chat.STipmsgWithStr_2{Type: 1, Str: "GM Succeed!"})
				default:
					body = mustProtoBytes(t, &chat.SSendChatMsg_5{})
				}
			}

			r.routeGatewayResponse(gmSys, tt.msgID, body)

			if r.Mark.Has(gmSendChatRequestName) {
				t.Fatalf("%s should be fired by gm response %d", gmSendChatRequestName, tt.msgID)
			}
		})
	}
}
