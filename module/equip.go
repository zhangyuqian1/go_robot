package module

import (
	"fmt"
	"stressTest/gameproto/bag"
	"stressTest/gameproto/common"
	"stressTest/gameproto/equip"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	equipSys           byte = 9
	equipTakeOn        byte = 1
	equipList          byte = 1
	equipTakeOnNotify  byte = 2
	equipTakeOffNotify byte = 3
	equipAttrUpdate    byte = 4

	bagSys           byte = 10
	bagGetItems      byte = 1
	bagPushItem      byte = 1
	bagAddNewItem    byte = 2
	bagItemNumChange byte = 3
	bagDeleteItem    byte = 4
	bagRecItem       byte = 3
	bagRecItemNotify byte = 7
	bagUseGift       byte = 7

	equipTakeOnRequestName       = "CTakeOnEquip"
	equipSellRequestName         = "CRecBagItem"
	equipBagRequestName          = "CGetBagItems"
	bagUseGiftRequestName        = "CUseBagGift"
	equipRequestTimeout    int64 = 10
)

func (r *Robot) CGetBagItems(bagType int32) {
	req := &bag.CGetBagItems_1{Type: bagType}
	r.AsyncBegin(equipBagRequestName, equipRequestTimeout, false)
	if err := r.sendGatewayRequest(equipBagRequestName, bagSys, bagGetItems, req, false); err != nil {
		r.Mark.Remove(equipBagRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(equipBagRequestName)
}

func (r *Robot) CTakeOnEquip(pos int32, series *common.Series) {
	if series == nil {
		logger.Error(fmt.Sprintf("序列号3: %v", series))
		return
	}
	// if pos <= 0 {
	// 	pos = 1 // ← 把文档规定的 pos=0 强制改成 1！
	// }
	logger.Error(fmt.Sprintf("序列号2: %v", series))
	req := &equip.CTakeOnEquip_1{Pos: pos, Series: series}
	r.AsyncBegin(equipTakeOnRequestName, equipRequestTimeout, false)
	if err := r.sendGatewayRequest(equipTakeOnRequestName, equipSys, equipTakeOn, req, false); err != nil {
		r.Mark.Remove(equipTakeOnRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(equipTakeOnRequestName)
}

func (r *Robot) CRecBagItem(items []*common.UserItem) {
	if len(items) == 0 {
		return
	}
	req := &bag.CRecBagItem_3{IsAuto: false}
	for _, item := range items {
		if item == nil || item.GetSeries() == nil {
			continue
		}
		req.Items = append(req.Items, &bag.CRecBagItem_3Single{BagType: 1, Series: item.GetSeries()})
	}
	if len(req.GetItems()) == 0 {
		return
	}
	r.AsyncBegin(equipSellRequestName, equipRequestTimeout, false)
	if err := r.sendGatewayRequest(equipSellRequestName, bagSys, bagRecItem, req, false); err != nil {
		r.Mark.Remove(equipSellRequestName)
		logger.Error(err.Error())
		return
	}
	for _, item := range items {
		r.clearPendingEquipItem(item.GetSeries())
	}
	r.AsyncWait(equipSellRequestName)
}

// CUseBagGift 使用角色礼包道具
func (r *Robot) CUseBagGift(bagType int32, series *common.Series, useCnt int32) {
	if series == nil {
		return
	}
	req := &bag.CUseBagGift_7{
		BagType: bagType,
		Serise:  series,
		UseCnt:  useCnt,
	}
	r.AsyncBegin(bagUseGiftRequestName, equipRequestTimeout, false)
	if err := r.sendGatewayRequest(bagUseGiftRequestName, bagSys, bagUseGift, req, false); err != nil {
		r.Mark.Remove(bagUseGiftRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(bagUseGiftRequestName)
}

func (r *Robot) handleEquipGatewayResponse(sys byte, msgID byte, body []byte) bool {
	switch sys {
	case equipSys:
		switch msgID {
		case equipList:
			r.handleEquipList(body)
		case equipTakeOnNotify:
			r.handleTakeOnEquipNotify(body)
		case equipTakeOffNotify, equipAttrUpdate:
			return true
		default:
			return false
		}
		return true
	case bagSys:
		switch msgID {
		case bagPushItem:
			r.handlePushBagItem(body)
		case bagAddNewItem:
			r.handleBagAddNewItem(body)
		case bagItemNumChange:
			return true
		case bagDeleteItem:
			r.handleBagDeleteItem(body)
		case bagRecItemNotify:
			r.handleRecBagItem(body)
		default:
			return false
		}
		return true
	default:
		return false
	}
}

func (r *Robot) handleEquipList(body []byte) {
	var resp equip.SEquipList_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_equip_list_1 error: %v", err)
		return
	}
	r.setPendingEquipItems(resp.GetItems())
}

func (r *Robot) handleTakeOnEquipNotify(body []byte) {
	var resp equip.STakeOnEquip_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_take_on_equip_2 error: %v", err)
		r.AsyncFire(equipTakeOnRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.clearPendingEquipItem(resp.GetSeries())
	r.AsyncFire(equipTakeOnRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handlePushBagItem(body []byte) {
	var resp bag.SPushBagItem_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_push_bag_item_1 error: %v", err)
		r.AsyncFire(equipBagRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.setPendingEquipItems(resp.GetList())
	r.AsyncFire(equipBagRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleBagAddNewItem(body []byte) {
	var resp bag.SBagAddNewItem_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_bag_add_new_item_2 error: %v", err)
		return
	}
	if resp.GetItem() == nil || r.Client == nil {
		return
	}
	r.Client.PendingEquipItems = append(r.Client.PendingEquipItems, resp.GetItem())
	if series := resp.GetItem().GetSeries(); series != nil {
		// 礼包 950 缓存到 GiftSeries，武器 1001~1120 缓存到 WeaponSeries，避免单值竞争
		if resp.GetItem().GetId() == 950 {
			r.Client.GiftSeries = series
		} else if resp.GetItem().GetId() >= 1001 && resp.GetItem().GetId() <= 1120 {
			r.Client.WeaponSeries = series
		}
	}
}

func (r *Robot) handleBagDeleteItem(body []byte) {
	var resp bag.SBagDeleteItem_4
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_bag_delete_item_4 error: %v", err)
		return
	}
	r.clearPendingEquipItem(resp.GetSeries())
}

func (r *Robot) handleRecBagItem(body []byte) {
	var resp bag.SRecBagItem_7
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_rec_bag_item_7 error: %v", err)
		r.AsyncFire(equipSellRequestName, 0, STEP_FAIL, 0)
		return
	}
	for _, item := range resp.GetItems() {
		r.clearPendingEquipItem(item.GetSeries())
	}
	r.AsyncFire(equipSellRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) setPendingEquipItems(items []*common.UserItem) {
	if r.Client == nil {
		return
	}
	r.Client.PendingEquipItems = append([]*common.UserItem(nil), items...)
}

func (r *Robot) clearPendingEquipItem(series *common.Series) {
	if r.Client == nil || series == nil || len(r.Client.PendingEquipItems) == 0 {
		return
	}
	kept := r.Client.PendingEquipItems[:0]
	for _, item := range r.Client.PendingEquipItems {
		if item == nil || !sameSeries(item.GetSeries(), series) {
			kept = append(kept, item)
		}
	}
	r.Client.PendingEquipItems = kept
}

func sameSeries(left *common.Series, right *common.Series) bool {
	if left == nil || right == nil {
		return false
	}
	return left.GetTime() == right.GetTime() && left.GetSeries() == right.GetSeries() && left.GetServer() == right.GetServer()
}
