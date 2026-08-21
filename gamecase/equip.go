package gamecase

import (
	"context"
	"math/rand"
	"stressTest/gameproto/common"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

func Equip(ctx context.Context) {
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	RobotInit(r)

	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.EquipRate == nil || global.EquipRate.Allow() {
				equipOnce(r)
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}

func equipOnce(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	action := nextEquipAction(r)
	if action == nil || action.Kind == module.EquipActionNone {
		r.CGetBagItems(1)
		return
	}
	switch action.Kind {
	case module.EquipActionTakeOn:
		r.CTakeOnEquip(action.Pos, action.Series)
	case module.EquipActionSell:
		r.CRecBagItem(action.Items)
	}
}

func nextEquipAction(r *module.Robot) *module.EquipAction {
	if r == nil || r.Client == nil || len(r.Client.PendingEquipItems) == 0 {
		return nil
	}
	items := r.Client.PendingEquipItems
	if len(items) > 1 {
		return &module.EquipAction{Kind: module.EquipActionSell, Items: append([]*common.UserItem(nil), items...)}
	}
	item := items[0]
	if item == nil || item.GetSeries() == nil {
		return nil
	}
	if rand.Intn(2) == 0 {
		return &module.EquipAction{Kind: module.EquipActionTakeOn, Pos: 1, Series: item.GetSeries(), Items: items}
	}
	return &module.EquipAction{Kind: module.EquipActionSell, Series: item.GetSeries(), Items: items}
}
