package module

import (
	"stressTest/gameproto/skill"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	SkillSys   byte = 5   // 技能系统号
	skill_draw  byte = 105 // c_skill_draw_105 技能抽卡
	skill_data  byte = 100 // s_skill_data_100 下发技能信息
	skill_fight byte = 101 // s_skill_on_fight_101 通知技能上阵
	skill_del   byte = 104 // s_skill_delete_104 删除技能

	drawRequestName       = "CSkillDraw"
	drawRequestTimeout    int64 = 10
)

// SkillState 缓存的技能运行态数据
type SkillState struct {
	ID         int32
	Level      int32
	Star       int32
	Grid       int32
	Quality    int32
	MaxQuality int32
	Flag       int32
}

// CDraw 技能抽卡
func (r *Robot) CDraw(idx int32, isAd bool) {
	req := &skill.CSkillDraw_105{
		Idx:  idx,
		IsAd: isAd,
	}
	r.AsyncBegin(drawRequestName, drawRequestTimeout, false)
	if err := r.sendGatewayRequest(drawRequestName, SkillSys, skill_draw, req, false); err != nil {
		r.Mark.Remove(drawRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(drawRequestName)
}

// handleSkillGatewayResponse 技能系统 Gateway 响应路由
func (r *Robot) handleSkillGatewayResponse(sys byte, msgID byte, body []byte) bool {
	if sys != SkillSys {
		return false
	}
	switch msgID {
	case skill_data:
		r.handleSkillData(body)
	case skill_fight:
		r.handleSkillOnFight(body)
	case skill_del:
		r.handleSkillDelete(body)
	default:
		return false
	}
	return true
}

// handleSkillData 处理 s_skill_data_100 下发技能信息
func (r *Robot) handleSkillData(body []byte) {
	var resp skill.SSkillData_100
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_skill_data_100 error: %v", err)
		r.AsyncFire(drawRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client != nil {
		if r.Client.Skills == nil {
			r.Client.Skills = make(map[int32]*SkillState)
		}
		for _, info := range resp.GetSkills() {
			if info.GetId() == 0 {
				continue
			}
			r.Client.Skills[info.GetId()] = &SkillState{
				ID:         info.GetId(),
				Level:      info.GetLv(),
				Star:       info.GetStar(),
				Grid:       info.GetGrid(),
				Quality:    info.GetQuality(),
				MaxQuality: info.GetMaxQuality(),
				Flag:       info.GetFlag(),
			}
		}
	}
	r.AsyncFire(drawRequestName, int64(len(body)), STEP_SUCC, 0)
}

// handleSkillOnFight 处理 s_skill_on_fight_101 通知技能上阵
func (r *Robot) handleSkillOnFight(body []byte) {
	var resp skill.SSkillOnFight_101
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_skill_on_fight_101 error: %v", err)
		return
	}
	if r.Client != nil && r.Client.Skills != nil {
		if s, ok := r.Client.Skills[resp.GetSkillId()]; ok {
			s.Grid = resp.GetGrid()
		}
	}
}

// handleSkillDelete 处理 s_skill_delete_104 删除技能
func (r *Robot) handleSkillDelete(body []byte) {
	var resp skill.SSkillDelete_104
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_skill_delete_104 error: %v", err)
		return
	}
	if r.Client != nil && r.Client.Skills != nil {
		delete(r.Client.Skills, resp.GetSkillId())
	}
}
