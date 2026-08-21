package module

import (
	"stressTest/gameproto/misc"
	"stressTest/gameproto/observer"
	"stressTest/gameproto/property"
	"stressTest/gameproto/skill"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	skillSys                  byte = 5
	skillNormalAttack         byte = 1
	skillDamageOtherValResult byte = 1
	skillStartUseSkillRequest byte = 2
	skillStartUseSkillResult  byte = 2
	skillDamageEntityRequest  byte = 3
	skillFrameProxyResult     byte = 4
	skillFrameProxyRequest    byte = 5
	skillEntityBeAttack       byte = 5

	miscSys                   byte = 15
	miscMinimapBossInfo       byte = 10
	miscMinimapBossInfoResult byte = 8

	observerSys              byte = 16
	observerActorAppear      byte = 2
	observerMonsterAppear    byte = 3
	observerEntityPropChange byte = 5
	observerEntityDisappear  byte = 6
	observerEntityDeath      byte = 8
	observerPetAppear        byte = 10

	propertyApplyRelive        byte = 1
	propertyActorMapDataChange byte = 2
	propertyOpenReliveDialog   byte = 7
	propertyCloseReliveDialog  byte = 8

	battleRequestName                = "CNormalAttack"
	startUseSkillRequestName         = "CStartUseSkill"
	skillFrameProxyRequestName       = "CSkillFrameProxy"
	damageEntityRequestName          = "CDamageEntity"
	minimapBossInfoRequestName       = "CMinimapBossInfo"
	reliveRequestName                = "CApplyRelive"
	battleRequestTimeout       int64 = 10
)

type SkillFrameProxyRequest struct {
	SkillID                int32
	SkillLevel             int32
	ChargeTime             int32
	IsChargeSkill          bool
	DirectionX             float32
	DirectionY             float32
	DirectionZ             float32
	CasterHandler          string
	TargetHandler          string
	TargetX                float32
	TargetY                float32
	TargetZ                float32
	LandingPoints          []*skill.SkillFrameVec3
	AttackWeaponAngle      float32
	AttackWeaponIsFaceLeft bool
	WeaponSlotIndex        int32
	EffectTriggerRate      int32
	RandomTargets          []*skill.SkillFrameRandomTarget
}

func (r *Robot) CNormalAttack(handle int64, skillID int32) {
	req := &skill.CNormalAttack_1{Handle: handle, SkillId: skillID}
	r.AsyncBegin(battleRequestName, battleRequestTimeout, true)
	if err := r.sendGatewayRequest(battleRequestName, skillSys, skillNormalAttack, req, true); err != nil {
		r.Mark.Remove(battleRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(battleRequestName)
}

func (r *Robot) CStartUseSkill(handle int64, skillID int32, dir int32, chargeTime int32) {
	req := &skill.CStartUseSkill_2{Handle: handle, SkillId: skillID, Dir: dir, ChargeTime: chargeTime}
	r.AsyncBegin(startUseSkillRequestName, battleRequestTimeout, true)
	if err := r.sendGatewayRequest(startUseSkillRequestName, skillSys, skillStartUseSkillRequest, req, true); err != nil {
		r.Mark.Remove(startUseSkillRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(startUseSkillRequestName)
}

func (r *Robot) CSkillFrameProxy(req SkillFrameProxyRequest) {
	if req.SkillLevel <= 0 {
		req.SkillLevel = 1
	}
	pb := &skill.CSkillFrameProxy_5{
		SkillId:                req.SkillID,
		SkillLevel:             req.SkillLevel,
		ChargeTime:             req.ChargeTime,
		IsChargeSkill:          req.IsChargeSkill,
		Direction:              &skill.SkillFrameVec3{X: req.DirectionX, Y: req.DirectionY, Z: req.DirectionZ},
		CasterHandler:          req.CasterHandler,
		TargetHandler:          req.TargetHandler,
		TargetPos:              &skill.SkillFrameVec3{X: req.TargetX, Y: req.TargetY, Z: req.TargetZ},
		LandingPoints:          req.LandingPoints,
		AttackWeaponAngle:      req.AttackWeaponAngle,
		AttackWeaponIsFaceLeft: req.AttackWeaponIsFaceLeft,
		WeaponSlotIndex:        req.WeaponSlotIndex,
		EffectTriggerRate:      req.EffectTriggerRate,
		RandomTargets:          req.RandomTargets,
	}
	r.AsyncBegin(skillFrameProxyRequestName, battleRequestTimeout, true)
	if err := r.sendGatewayRequest(skillFrameProxyRequestName, skillSys, skillFrameProxyRequest, pb, true); err != nil {
		r.Mark.Remove(skillFrameProxyRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(skillFrameProxyRequestName)
}

func (r *Robot) CDamageEntity(handle int64, skillID int32, rangeIdx int32, grid int32) {
	req := &skill.CDamageEntity_3{Handle: handle, SkillId: skillID, RangeIdx: rangeIdx, Grid: grid}
	r.AsyncBegin(damageEntityRequestName, battleRequestTimeout, true)
	if err := r.sendGatewayRequest(damageEntityRequestName, skillSys, skillDamageEntityRequest, req, true); err != nil {
		r.Mark.Remove(damageEntityRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(damageEntityRequestName)
}

func (r *Robot) CMinimapBossInfo(sid int32) {
	req := &misc.CMinimapBossInfo_10{Sid: sid}
	r.AsyncBegin(minimapBossInfoRequestName, battleRequestTimeout, false)
	if err := r.sendGatewayRequest(minimapBossInfoRequestName, miscSys, miscMinimapBossInfo, req, false); err != nil {
		r.Mark.Remove(minimapBossInfoRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(minimapBossInfoRequestName)
}

func (r *Robot) CApplyRelive(reliveType int32) {
	req := &property.CApplyRelive_1{Type: reliveType}
	r.AsyncBegin(reliveRequestName, battleRequestTimeout, false)
	if err := r.sendGatewayRequest(reliveRequestName, propertySys, propertyApplyRelive, req, false); err != nil {
		r.Mark.Remove(reliveRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(reliveRequestName)
}

func (r *Robot) handleBattleGatewayResponse(sys byte, msgID byte, body []byte) bool {
	switch sys {
	case observerSys:
		switch msgID {
		case observerActorAppear:
			r.handleActorAppear(body)
		case observerMonsterAppear:
			r.handleMonsterAppear(body)
		case observerEntityPropChange:
			r.handleEntityPropChange(body)
		case observerEntityDisappear:
			r.handleEntityDisappear(body)
		case observerEntityDeath:
			r.handleEntityDeath(body)
		case observerPetAppear:
			return true
		default:
			return false
		}
		return true
	case skillSys:
		switch msgID {
		case skillDamageOtherValResult:
			r.handleDamageOtherVal(body)
		case skillStartUseSkillResult:
			r.handleStartUseSkill(body)
		case skillFrameProxyResult:
			r.handleSkillFrameProxy(body)
		case skillEntityBeAttack:
			r.handleEntityBeAttack(body)
		default:
			return false
		}
		return true
	case miscSys:
		switch msgID {
		case miscMinimapBossInfoResult:
			r.handleMinimapBossInfo(body)
		default:
			return false
		}
		return true
	case propertySys:
		switch msgID {
		case propertyActorMapDataChange:
			r.handleActorMapDataChange(body)
		case propertyOpenReliveDialog:
			r.handleOpenReliveDialog(body)
		case propertyCloseReliveDialog:
			r.handleCloseReliveDialog(body)
		default:
			return false
		}
		return true
	default:
		return false
	}
}

func (r *Robot) handleActorAppear(body []byte) {
	var resp observer.SActorAppear_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_actor_appear_2 error: %v", err)
		return
	}
	common := resp.GetCommon()
	if common == nil {
		return
	}
	r.cacheEntityPosition(common.GetHandle(), common.GetX(), common.GetY(), common.GetDir())
}

func (r *Robot) handleMonsterAppear(body []byte) {
	var resp observer.SMonsterAppear_3
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_monster_appear_3 error: %v", err)
		return
	}
	common := resp.GetCommon()
	if r.Client == nil || common == nil || common.GetHandle() == 0 {
		return
	}
	if r.Client.Monsters == nil {
		r.Client.Monsters = make(map[int64]*MonsterState)
	}
	r.Client.Monsters[common.GetHandle()] = &MonsterState{
		Handle: common.GetHandle(),
		ID:     common.GetId(),
		Midx:   resp.GetMidx(),
		X:      common.GetX(),
		Y:      common.GetY(),
		CurHP:  resp.GetCurHp(),
		MaxHP:  resp.GetMaxHp(),
		Name:   resp.GetName(),
		Skills: append([]int32(nil), resp.GetSkills()...),
	}
	r.cacheEntityPosition(common.GetHandle(), common.GetX(), common.GetY(), common.GetDir())
}

func (r *Robot) handleEntityPropChange(body []byte) {
	var resp observer.SEntityPropChange_5
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_entity_prop_change_5 error: %v", err)
		r.AsyncFire(skillFrameProxyRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil && resp.GetHandle() != 0 {
		if r.Client.Entities == nil {
			r.Client.Entities = make(map[int64]*EntityState)
		}
		entity := r.Client.Entities[resp.GetHandle()]
		if entity == nil {
			entity = &EntityState{Handle: resp.GetHandle()}
			r.Client.Entities[resp.GetHandle()] = entity
		}
		entity.Props = copyInt64Map(resp.GetVals())
	}
	r.AsyncFire(skillFrameProxyRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleEntityDisappear(body []byte) {
	var resp observer.SEntityDisappear_6
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_entity_disappear_6 error: %v", err)
		return
	}
	r.removeEntity(resp.GetHandle())
}

func (r *Robot) handleEntityDeath(body []byte) {
	var resp observer.SEntityDeath_8
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_entity_death_8 error: %v", err)
		return
	}
	if r.Client != nil && resp.GetHandle() == r.Client.Handle {
		r.Client.SelfDead = true
		r.Client.BattleTargetHandle = 0
		return
	}
	r.removeEntity(resp.GetHandle())
}

func (r *Robot) handleDamageOtherVal(body []byte) {
	var resp skill.SDamageOtherVal_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_damage_other_val_1 error: %v", err)
		r.AsyncFire(damageEntityRequestName, 0, STEP_FAIL, 0)
		r.AsyncFire(battleRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.AsyncFire(damageEntityRequestName, int64(len(body)), STEP_SUCC, 0)
	r.AsyncFire(battleRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleStartUseSkill(body []byte) {
	var resp skill.SStartUseSkill_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_start_use_skill_2 error: %v", err)
		r.AsyncFire(startUseSkillRequestName, 0, STEP_FAIL, 0)
		return
	}
	if resp.GetResult() != 0 {
		r.AsyncFire(startUseSkillRequestName, int64(len(body)), STEP_SUCC, 0)
		return
	}
	if r.Client != nil {
		r.Client.BattleSkillRandVal = resp.GetRandVal()
	}
	r.AsyncFire(startUseSkillRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleSkillFrameProxy(body []byte) {
	var resp skill.SSkillFrameProxy_4
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_skill_frame_proxy_4 error: %v", err)
		r.AsyncFire(skillFrameProxyRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.AsyncFire(skillFrameProxyRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleEntityBeAttack(body []byte) {
	var resp skill.SEntityBeattack_5
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_entity_beattack_5 error: %v", err)
		r.AsyncFire(skillFrameProxyRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.AsyncFire(skillFrameProxyRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleMinimapBossInfo(body []byte) {
	var resp misc.SMinimapBossInfo_8
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_minimap_boss_info_8 error: %v", err)
		r.AsyncFire(minimapBossInfoRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil {
		bosses := make(map[int32]*BossState, len(resp.GetBosses()))
		for _, boss := range resp.GetBosses() {
			if boss.GetId() == 0 {
				continue
			}
			bosses[boss.GetId()] = &BossState{
				ID:     boss.GetId(),
				X:      boss.GetX(),
				Y:      boss.GetY(),
				Relive: boss.GetRelive(),
			}
		}
		r.Client.BattleBosses = bosses
	}
	r.AsyncFire(minimapBossInfoRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleActorMapDataChange(body []byte) {
	var resp property.SActorMapDataChange_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_actor_map_data_change_2 error: %v", err)
		return
	}
	if r.Client == nil {
		return
	}
	if r.Client.MapData == nil {
		r.Client.MapData = make(map[int32]int64)
	}
	for key, value := range resp.GetVals() {
		r.Client.MapData[key] = value
	}
}

func (r *Robot) handleOpenReliveDialog(body []byte) {
	var resp property.SOpenReliveDialog_7
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_open_relive_dialog_7 error: %v", err)
		return
	}
	if r.Client != nil {
		r.Client.SelfDead = true
	}
}

func (r *Robot) handleCloseReliveDialog(body []byte) {
	var resp property.SCloseReliveDialog_8
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_close_relive_dialog_8 error: %v", err)
		r.AsyncFire(reliveRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil {
		r.Client.SelfDead = false
	}
	r.AsyncFire(reliveRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) removeEntity(handle int64) {
	if r.Client == nil || handle == 0 {
		return
	}
	delete(r.Client.Entities, handle)
	delete(r.Client.Monsters, handle)
	if r.Client.BattleTargetHandle == handle {
		r.Client.BattleTargetHandle = 0
	}
}

func copyInt64Map(source map[int32]int64) map[int32]int64 {
	if len(source) == 0 {
		return nil
	}
	copied := make(map[int32]int64, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}
