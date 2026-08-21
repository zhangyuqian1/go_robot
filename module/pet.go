package module

import (
	"stressTest/gameproto/pet"
	"stressTest/global"
	"stressTest/logger"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	petSys    byte = 19
	petUpdate byte = 1
	petRemove byte = 2
	petBattle byte = 7
	petSummon byte = 8

	petBattleRequestName       = "CPetBattle"
	petSummonRequestName       = "CPetSummon"
	petRequestTimeout    int64 = 10
)

func (r *Robot) CPetBattle(seriesID int32, index int32) {
	if index <= 0 {
		index = 1
	}
	req := &pet.CPetBattle_7{SeriseId: seriesID, Index: index}
	r.AsyncBegin(petBattleRequestName, petRequestTimeout, false)
	if err := r.sendGatewayRequest(petBattleRequestName, petSys, petBattle, req, false); err != nil {
		r.Mark.Remove(petBattleRequestName)
		logger.Error(err.Error())
		return
	}
	if r.Client != nil {
		r.Client.SelectedPetSeriesID = seriesID
		r.Client.PetBattlePending = false
	}
	r.AsyncWait(petBattleRequestName)
}

func (r *Robot) CPetSummon(index int32) {
	if index <= 0 {
		index = 1
	}
	req := &pet.CPetSummon_8{Index: index}
	r.AsyncBegin(petSummonRequestName, petRequestTimeout, false)
	if err := r.sendGatewayRequest(petSummonRequestName, petSys, petSummon, req, false); err != nil {
		r.Mark.Remove(petSummonRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(petSummonRequestName)
}

func (r *Robot) handlePetGatewayResponse(sys byte, msgID byte, body []byte) bool {
	if sys != petSys {
		return false
	}
	switch msgID {
	case petUpdate:
		r.handlePetUpdate(body)
	case petRemove:
		r.handlePetRemove(body)
	default:
		return false
	}
	return true
}

func (r *Robot) handlePetUpdate(body []byte) {
	var resp pet.SPetUpdate_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_pet_update_1 error: %v", err)
		r.AsyncFire(petBattleRequestName, 0, STEP_FAIL, 0)
		r.AsyncFire(petSummonRequestName, 0, STEP_FAIL, 0)
		return
	}
	if r.Client == nil {
		return
	}
	if r.Client.Pets == nil {
		r.Client.Pets = make(map[int32]*PetState)
	}
	for _, item := range resp.GetPetInfo() {
		if item.GetSeriseId() == 0 {
			continue
		}
		r.Client.Pets[item.GetSeriseId()] = &PetState{
			SeriesID: item.GetSeriseId(),
			ID:       item.GetId(),
			Level:    item.GetLevel(),
			Star:     item.GetStar(),
		}
		if r.Client.SelectedPetSeriesID == 0 {
			r.Client.SelectedPetSeriesID = item.GetSeriseId()
		}
	}
	if len(resp.GetPetInfo()) > 0 {
		r.Client.PetBattlePending = true
		r.Client.PetCooldownUntil = time.Time{}
	}
	r.AsyncFire(petBattleRequestName, int64(len(body)), STEP_SUCC, 0)
	r.AsyncFire(petSummonRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handlePetRemove(body []byte) {
	var resp pet.SPetRemove_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_pet_remove_2 error: %v", err)
		return
	}
	if r.Client == nil {
		return
	}
	for _, seriesID := range resp.GetRemoveSerise() {
		delete(r.Client.Pets, seriesID)
		if r.Client.SelectedPetSeriesID == seriesID {
			r.Client.SelectedPetSeriesID = 0
		}
	}
	r.Client.PetCooldownUntil = time.Now().Add(global.PetCooldown)
	r.Client.PetBattlePending = true
}
