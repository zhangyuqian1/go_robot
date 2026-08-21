package gamecase

import (
	"testing"

	"stressTest/module"
)

func TestSelectBattleTargetCachesAliveMonster(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.Monsters = map[int64]*module.MonsterState{
		100: {Handle: 100, CurHP: 0},
		200: {Handle: 200, CurHP: 99},
	}

	target := selectBattleTarget(r)
	if target == nil || target.Handle != 200 || r.Client.BattleTargetHandle != 200 {
		t.Fatalf("selectBattleTarget = %#v target:%d, want monster 200", target, r.Client.BattleTargetHandle)
	}
}

func TestSelectBattleTargetPrefersMonsterAppearBoss(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.Monsters = map[int64]*module.MonsterState{
		100: {Handle: 100, ID: 63, Name: "其它Boss", CurHP: 99},
		200: {Handle: 200, ID: 64, Name: "骸骨巨龙", CurHP: 99},
	}

	target := selectBattleTarget(r)
	if target == nil || target.Handle != 200 || r.Client.BattleTargetHandle != 200 {
		t.Fatalf("selectBattleTarget = %#v target:%d, want monster appear boss 64", target, r.Client.BattleTargetHandle)
	}
}

func TestSelectBattleTargetSkipsOtherBossWhenMonsterAppearBossMissing(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.BattleTargetHandle = 100
	r.Client.Monsters = map[int64]*module.MonsterState{
		100: {Handle: 100, ID: 63, Name: "其它Boss", CurHP: 99},
	}

	target := selectBattleTarget(r)
	if target != nil || r.Client.BattleTargetHandle != 0 {
		t.Fatalf("selectBattleTarget = %#v target:%d, want no target before monster appear boss 64", target, r.Client.BattleTargetHandle)
	}
}

func TestBuildBattleSkillFrameRequestUsesDocumentDefaults(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.Handle = 908066816
	r.Client.BattleSkillRandVal = 2626
	target := &module.MonsterState{Handle: 5477761381, X: 2024, Y: 10780}

	req := buildBattleSkillFrameRequest(r, target, 4209, 62)

	if req.SkillID != 4209 || req.SkillLevel != 1 || req.ChargeTime != 0 {
		t.Fatalf("skill frame skill fields = id:%d level:%d charge:%d", req.SkillID, req.SkillLevel, req.ChargeTime)
	}
	if req.CasterHandler != "908066816" || req.TargetHandler != "5477761381" {
		t.Fatalf("skill frame handlers = caster:%q target:%q", req.CasterHandler, req.TargetHandler)
	}
	if req.TargetX != 2024 || req.TargetY != 10780 || req.WeaponSlotIndex != 0 || req.EffectTriggerRate != 2626 {
		t.Fatalf("skill frame target/defaults = x:%f y:%f weapon:%d effect:%d", req.TargetX, req.TargetY, req.WeaponSlotIndex, req.EffectTriggerRate)
	}
}

func TestBattleRangeIndexFollowsSkillID(t *testing.T) {
	if got := battleRangeIndex(4209); got != 42090 {
		t.Fatalf("battleRangeIndex(4209) = %d, want 42090", got)
	}
}
