package gamecase

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

const (
	battleSceneID        int32 = 3
	battleMoveCenterX    int32 = 20
	battleMoveCenterY    int32 = 100
	battleMoveScatter    int32 = 10
	battleDefaultSkillID int32 = 4209
	battleTargetBossID   int32 = 64
	battleTargetBossName       = "骸骨巨龙"
)

func Battle(ctx context.Context) {
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	RobotInit(r)
	time.Sleep(1 * time.Second)
	initBattle(r)

	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.BattleRate == nil || global.BattleRate.Allow() {
				battleOnce(r)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func initBattle(r *module.Robot) {
	if r == nil {
		return
	}
	x := battleMoveCenterX + int32(rand.Intn(int(battleMoveScatter*2+1))) - battleMoveScatter
	y := battleMoveCenterY + int32(rand.Intn(int(battleMoveScatter*2+1))) - battleMoveScatter
	r.CSendChatMsg(0, "@superman")
	r.CSendChatMsg(0, fmt.Sprintf("@MoveTo %d %d %d", battleSceneID, x, y))
	r.CMinimapBossInfo(battleSceneID)
}

func battleOnce(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	if r.Client.SelfDead {
		r.CApplyRelive(1)
		initBattle(r)
		return
	}
	target := selectBattleTarget(r)
	if target == nil {
		return
	}
	skillID := battleSkillID()
	dir := battleDirection(r, target)
	r.CStartUseSkill(r.Client.Handle, skillID, dir, 0)
	r.CSkillFrameProxy(buildBattleSkillFrameRequest(r, target, skillID, dir))
	r.CDamageEntity(target.Handle, skillID, battleRangeIndex(skillID), 0)
	time.Sleep(100 * time.Millisecond)
}

func selectBattleTarget(r *module.Robot) *module.MonsterState {
	if r == nil || r.Client == nil || len(r.Client.Monsters) == 0 {
		return nil
	}
	if target := selectBattleBossTarget(r); target != nil {
		return target
	}
	if hasKnownBattleBoss(r.Client.Monsters) {
		r.Client.BattleTargetHandle = 0
		return nil
	}
	if r.Client.BattleTargetHandle != 0 {
		if target := r.Client.Monsters[r.Client.BattleTargetHandle]; target != nil && target.CurHP != 0 {
			return target
		}
	}
	for handle, monster := range r.Client.Monsters {
		if monster == nil || monster.CurHP == 0 {
			continue
		}
		r.Client.BattleTargetHandle = handle
		return monster
	}
	return nil
}

func selectBattleBossTarget(r *module.Robot) *module.MonsterState {
	if r.Client.BattleTargetHandle != 0 {
		if target := r.Client.Monsters[r.Client.BattleTargetHandle]; isBattleBossTarget(target) {
			return target
		}
	}
	for handle, monster := range r.Client.Monsters {
		if !isBattleBossTarget(monster) {
			continue
		}
		r.Client.BattleTargetHandle = handle
		return monster
	}
	return nil
}

func isBattleBossTarget(monster *module.MonsterState) bool {
	if monster == nil || monster.CurHP == 0 {
		return false
	}
	return monster.ID == battleTargetBossID || monster.Name == battleTargetBossName
}

func hasKnownBattleBoss(monsters map[int64]*module.MonsterState) bool {
	for _, monster := range monsters {
		if monster != nil && (monster.ID != 0 || monster.Name != "") {
			return true
		}
	}
	return false
}

func buildBattleSkillFrameRequest(r *module.Robot, target *module.MonsterState, skillID int32, dir int32) module.SkillFrameProxyRequest {
	directionX, directionY := battleDirectionVector(dir)
	effectTriggerRate := int32(0)
	casterHandler := ""
	if r != nil && r.Client != nil {
		effectTriggerRate = r.Client.BattleSkillRandVal
		casterHandler = fmt.Sprintf("%d", r.Client.Handle)
	}
	if effectTriggerRate <= 0 {
		effectTriggerRate = int32(rand.Intn(10000))
	}
	targetHandler := ""
	var targetX float32
	var targetY float32
	if target != nil {
		targetHandler = fmt.Sprintf("%d", target.Handle)
		targetX = float32(target.X)
		targetY = float32(target.Y)
	}
	return module.SkillFrameProxyRequest{
		SkillID:                skillID,
		SkillLevel:             1,
		ChargeTime:             0,
		IsChargeSkill:          false,
		DirectionX:             directionX,
		DirectionY:             directionY,
		DirectionZ:             0,
		CasterHandler:          casterHandler,
		TargetHandler:          targetHandler,
		TargetX:                targetX,
		TargetY:                targetY,
		TargetZ:                0,
		AttackWeaponAngle:      normalizeAngle(float32(90 - dir)),
		AttackWeaponIsFaceLeft: directionX < 0,
		WeaponSlotIndex:        0,
		EffectTriggerRate:      effectTriggerRate,
	}
}

func battleSkillID() int32 {
	if global.BattleNormalSkillID > 0 {
		return global.BattleNormalSkillID
	}
	return battleDefaultSkillID
}

func battleRangeIndex(skillID int32) int32 {
	if skillID <= 0 {
		return 0
	}
	return skillID * 10
}

func battleDirection(r *module.Robot, target *module.MonsterState) int32 {
	if r == nil || r.Client == nil || target == nil {
		return 0
	}
	dx := float64(target.X) - float64(r.Client.PX)
	dy := float64(target.Y) - float64(r.Client.PY)
	degrees := math.Atan2(dx, dy) * 180 / math.Pi
	if degrees < 0 {
		degrees += 360
	}
	return int32(math.Round(degrees)) % 360
}

func battleDirectionVector(dir int32) (float32, float32) {
	radians := float64(dir) * math.Pi / 180
	return float32(math.Sin(radians)), float32(math.Cos(radians))
}

func normalizeAngle(angle float32) float32 {
	for angle < 0 {
		angle += 360
	}
	for angle >= 360 {
		angle -= 360
	}
	return angle
}
