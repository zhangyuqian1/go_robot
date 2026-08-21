package module

import (
	"testing"

	"stressTest/gameproto/bag"
	"stressTest/gameproto/common"
)

func TestHandleBagAddNewItemCachesGiftSeries(t *testing.T) {
	r := newTestRobot()
	series := &common.Series{Time: 209386384, Series: 1, Server: 100}
	item := &common.UserItem{Series: series, Id: 950, Count: 1}
	body := mustProtoBytes(t, &bag.SBagAddNewItem_2{Type: 3, Item: item})

	r.handleBagAddNewItem(body)

	if r.Client.GiftSeries == nil {
		t.Fatalf("GiftSeries not cached")
	}
	if r.Client.GiftSeries.GetSeries() != 1 || r.Client.GiftSeries.GetServer() != 100 || r.Client.GiftSeries.GetTime() != 209386384 {
		t.Fatalf("GiftSeries = %#v, want series=1 server=100 time=209386384", r.Client.GiftSeries)
	}
	if len(r.Client.PendingEquipItems) != 1 || r.Client.PendingEquipItems[0].GetId() != 950 {
		t.Fatalf("PendingEquipItems = %#v, want one item id=950", r.Client.PendingEquipItems)
	}
}

func TestHandleBagAddNewItemNilItem(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &bag.SBagAddNewItem_2{Type: 3})

	r.handleBagAddNewItem(body)

	if r.Client.GiftSeries != nil {
		t.Fatalf("GiftSeries should be nil, got %#v", r.Client.GiftSeries)
	}
}

func TestHandleBagAddNewItemCachesWeaponSeries(t *testing.T) {
	r := newTestRobot()
	series := &common.Series{Time: 209386384, Series: 1, Server: 100}
	item := &common.UserItem{Series: series, Id: 1001, Count: 1}
	body := mustProtoBytes(t, &bag.SBagAddNewItem_2{Type: 3, Item: item})

	r.handleBagAddNewItem(body)

	if r.Client.WeaponSeries == nil {
		t.Fatalf("WeaponSeries not cached")
	}
	if r.Client.WeaponSeries.GetSeries() != 1 || r.Client.WeaponSeries.GetServer() != 100 || r.Client.WeaponSeries.GetTime() != 209386384 {
		t.Fatalf("WeaponSeries = %#v, want series=1 server=100 time=209386384", r.Client.WeaponSeries)
	}
	if r.Client.GiftSeries != nil {
		t.Fatalf("GiftSeries should be nil for weapon item, got %#v", r.Client.GiftSeries)
	}
}
