package gamecase

import (
	"testing"

	"stressTest/gameproto/common"
	"stressTest/module"
)

func TestNextEquipActionUsesCachedItems(t *testing.T) {
	r := newPostLoginTestRobot()
	item1 := &common.UserItem{Series: &common.Series{Time: 1, Series: 2, Server: 3}, Id: 1001, Count: 1}
	item2 := &common.UserItem{Series: &common.Series{Time: 4, Series: 5, Server: 6}, Id: 1002, Count: 1}

	r.Client.PendingEquipItems = []*common.UserItem{item1}
	if action := nextEquipAction(r); action == nil || action.Kind == module.EquipActionNone || action.Series.GetSeries() != 2 {
		t.Fatalf("single equip action = %#v", action)
	}

	r.Client.PendingEquipItems = []*common.UserItem{item1, item2}
	if action := nextEquipAction(r); action == nil || action.Kind != module.EquipActionSell || len(action.Items) != 2 {
		t.Fatalf("multi equip action = %#v", action)
	}
}
