package gamecase

import (
	"fmt"
	"net/http"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

func StartFunc() *module.Robot {
	r := &module.Robot{}

	if global.GAccountList != nil && global.PodRobotNum > len(global.GAccountList) {
		panic("文本用户量不足")
	}
	accountSufix := global.AccountDispatcher()
	if global.GAccountList != nil {
		r.PTName = global.GAccountList[accountSufix]

	} else {
		r.PTName = fmt.Sprintf("%s%d", global.AccountPrefix, accountSufix)
	}

	r.Index = accountSufix

	r.Mark = cmap.New[*module.Events]()

	logger.Debug(fmt.Sprintf("PTName: %s, PassId: %s", r.PTName, r.PassId))

	transConf := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 2,
		DisableKeepAlives:   false,
	}

	r.Client = &module.GameClient{
		WebSocketSvr: global.WebSocketSvr,
		OpenTLS:      false,
		HttpClient: &http.Client{
			Timeout:   time.Duration(30 * time.Second),
			Transport: transConf,
		},
	}
	return r
}

// waitGiftSeries 等待 s_bag_add_new_item_2 回调填充 GiftSeries，超时 5 秒，只缓存一个道具
func waitGiftSeries(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	deadline := time.Now().Add(5 * time.Second)
	for r.Client.GiftSeries == nil {
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// waitWeaponSeries 等待 s_bag_add_new_item_2 回调填充 WeaponSeries，超时 5 秒
func waitWeaponSeries(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	deadline := time.Now().Add(20 * time.Second)
	for r.Client.WeaponSeries == nil {
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// waitMinSkillID 等待穿戴武器后 s_skill_data_100 回调填充 MinSkillID，超时 5 秒
func waitMinSkillID(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	deadline := time.Now().Add(5 * time.Second)
	for r.Client.MinSkillID == 0 {
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// addItemsInRange 循环使用 GM 指令 @additem 获取道具，id 从 startID 到 endID（含），每个间隔 500ms
func addItemsInRange(r *module.Robot, startID, endID int32) {
	if r == nil {
		return
	}
	if startID > endID {
		startID, endID = endID, startID
	}
	for id := startID; id <= endID; id++ {
		r.CSendChatMsg(0, fmt.Sprintf("@additem %d 1", id))
		time.Sleep(100 * time.Millisecond)
	}
}

// addPetsInRange 循环使用 GM 指令 @additem 获取宠物，id 从 startID 到 endID（含），每个间隔 500ms
func addPetsInRange(r *module.Robot, startID, endID int32) {
	if r == nil {
		return
	}
	if startID > endID {
		startID, endID = endID, startID
	}
	for id := startID; id <= endID; id++ {
		r.CSendChatMsg(0, fmt.Sprintf("@func pet_add %d", id))
		time.Sleep(10 * time.Millisecond)
	}
}

// waitPetUpdate 等待 s_pet_update_1 回调填充 Pets 缓存（serise_id 已由服务端生成），超时 5 秒
func waitPetUpdate(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	deadline := time.Now().Add(5 * time.Second)
	for len(r.Client.Pets) == 0 {
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// pickPetSeriesID 从 Pets 缓存中取一个真实的 serise_id；缓存为空返回 0
func pickPetSeriesID(r *module.Robot) int32 {
	if r == nil || r.Client == nil || len(r.Client.Pets) == 0 {
		return 0
	}
	for seriesID := range r.Client.Pets {
		return seriesID
	}
	return 0
}

func RobotInit(r *module.Robot) {
	r.TickStopSignal = make(chan struct{})
	r.ConnStopSignal = make(chan struct{})
	r.AsyncMonitSignal = make(chan struct{})
	for {
		if global.LoginRate.Allow() {
			err := r.CreateWebSocketConn()
			if err != nil {

				logger.Error(fmt.Sprintf("robot create websocket: %s", err))
				global.LoginRendezvous.Done(-1)
				return
			}
			// 21000
			//r.CReqCheckPassword(120, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 120, 120, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
			// 21001
			//r.CReqCheckPassword(121, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 121, 121, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
			//21003
			//r.CReqCheckPassword(125, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 125, 125, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
			//21004
			r.CReqCheckPassword(126, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 126, 126, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)

			//r.CReqCheckPassword(1003, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 1003, 1003, "47.96.18.95", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
			// r.CReqCheckPassword(124, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 124, 124, "10.238.20.165", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
			if r.Client.ActorID == 0 {
				r.CReqCreateActor(1, 1, fmt.Sprintf("L%d", r.Index), 1, "e10adc3949ba59abbe56e057f20f883e")
				r.CReqGetActorList()
			}
			r.CReqClientReadyLogin(r.Client.ActorID)
			r.CReqActorLogin(r.Client.AccountID, r.Client.ActorID)
			time.Sleep(1 * time.Second)
			// r.CSendChatMsg(0, "@set 10066 41")
			// r.CSendChatMsg(0, "@level 130")
			// r.CSendChatMsg(0, "@additem 950 1")
			// time.Sleep(1 * time.Second)
			// waitGiftSeries(r)
			// logger.Error(fmt.Sprintf("序列号1: %v", r.Client.GiftSeries))
			// r.CUseBagGift(3, r.Client.GiftSeries, 1)
			// // 随机选择 2~25 的英雄 id，传给后续两个协议
			// heroID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(24) + 2
			// r.CActivityHero(int32(heroID))
			// r.CSwitchHero(int32(heroID))

			// // waitMinSkillID(r)
			// //logger.Error(fmt.Sprintf("穿戴后最小技能ID: %d", r.Client.UltimateSkillID))
			// weaponID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(120) + 1001
			// r.CSendChatMsg(0, fmt.Sprintf("@additem %d 1", weaponID))
			// time.Sleep(1 * time.Second)
			// waitWeaponSeries(r)
			// r.CTakeOnEquip(0, r.Client.WeaponSeries)
			// // logger.Error(fmt.Sprintf("序列号2: %v", r.Client.WeaponSeries))
			// time.Sleep(1 * time.Second)
			// // 穿戴武器后会收到两条 s_skill_data_100，取其中 id 最小的技能
			// waitMinSkillID(r)
			// logger.Error(fmt.Sprintf("穿戴后最小技能ID: %d", r.Client.MinSkillID))
			// //示例：用最小技能 id 发起普攻/技能
			// r.CStartUseSkill(r.Client.Handle, r.Client.MinSkillID, 0, 0)
			// r.CDamageEntity(r.Client.Handle, r.Client.MinSkillID, r.Client.MinSkillID*10, 0)
			// r.CSendChatMsg(0, "@level 160")
			// time.Sleep(1 * time.Second)
			// //循环获取坐骑道具（id 范围 601~620），再激活并穿戴
			// rideID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(20) + 1
			// addItemsInRange(r, 601, 621)
			// time.Sleep(1 * time.Second)
			// r.CRideAction(int32(rideID), 1) // type=1 激活坐骑
			// time.Sleep(500 * time.Millisecond)
			// r.CRideAction(int32(rideID), 5) // type=5 穿戴坐骑
			// time.Sleep(500 * time.Millisecond)
			// r.CRideAction(int32(rideID), 4) // type=4 骑乘

			// addPetsInRange(r, 1, 35)

			// time.Sleep(5 * time.Second)
			// r.CPetSummon(3)
			// waitPetUpdate(r) // 等 s_pet_update_1 回调，服务端已生成 serise_id
			// petSeriesID := pickPetSeriesID(r)
			// logger.Error(fmt.Sprintf("上阵宠物 serise_id: %d", petSeriesID))
			// if petSeriesID != 0 {
			// 	logger.Error(fmt.Sprintf("上阵宠物1 serise_id: %d", petSeriesID))
			// 	// petID := int32(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(35) + 1)
			// 	r.CPetBattle(petSeriesID, 1) // c_pet_battle_7 穿戴宠物（index=1）
			// }
			// r.CTeamDungeonMatch(1)
			break
		}
	}
}

// func Robot_Init(r *module.Robot) {
// 	r.TickStopSignal = make(chan struct{})
// 	r.ConnStopSignal = make(chan struct{})
// 	r.AsyncMonitSignal = make(chan struct{})
// 	for {
// 		if global.LoginRate.Allow() {
// 			err := r.CreateWebSocketConn()
// 			if err != nil {

// 				logger.Error(fmt.Sprintf("robot create websocket: %s", err))
// 				global.LoginRendezvous.Done(-1)
// 				return
// 			}
// 			r.CReqCheckPassword(120, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 120, 120, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
// 			//r.CReqCheckPassword(1003, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 1003, 1003, "47.96.18.95", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
// 			// r.CReqCheckPassword(124, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 124, 124, "10.238.20.165", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
// 			if r.Client.ActorID == 0 {
// 				r.CReqCreateActor(1, 1, fmt.Sprintf("L%d", r.Index), 1, "e10adc3949ba59abbe56e057f20f883e")
// 				r.CReqGetActorList()
// 			}
// 			r.CReqClientReadyLogin(r.Client.ActorID)
// 			r.CReqActorLogin(r.Client.AccountID, r.Client.ActorID)
// 			time.Sleep(1 * time.Second)
// 			r.CSendChatMsg(0, "@set 10066 41")
// 			r.CSendChatMsg(0, "@level 130")
// 			r.CSendChatMsg(0, "@additem 950 1")
// 			time.Sleep(1 * time.Second)
// 			waitGiftSeries(r)
// 			logger.Error(fmt.Sprintf("序列号1: %v", r.Client.GiftSeries))
// 			r.CUseBagGift(3, r.Client.GiftSeries, 1)
// 			// 随机选择 2~25 的英雄 id，传给后续两个协议
// 			heroID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(24) + 2
// 			r.CActivityHero(int32(heroID))
// 			r.CSwitchHero(int32(heroID))

// 			waitMinSkillID(r)
// 			// logger.Error(fmt.Sprintf("穿戴后最小技能ID: %d", r.Client.UltimateSkillID))
// 			weaponID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(120) + 1001
// 			r.CSendChatMsg(0, fmt.Sprintf("@additem %d 1", weaponID))
// 			time.Sleep(1 * time.Second)
// 			waitWeaponSeries(r)
// 			r.CTakeOnEquip(0, r.Client.WeaponSeries)
// 			// logger.Error(fmt.Sprintf("序列号2: %v", r.Client.WeaponSeries))
// 			time.Sleep(1 * time.Second)
// 			// 穿戴武器后会收到两条 s_skill_data_100，取其中 id 最小的技能
// 			waitMinSkillID(r)
// 			logger.Error(fmt.Sprintf("穿戴后最小技能ID: %d", r.Client.MinSkillID))
// 			//示例：用最小技能 id 发起普攻/技能
// 			r.CStartUseSkill(r.Client.Handle, r.Client.MinSkillID, 0, 0)
// 			r.CDamageEntity(r.Client.Handle, r.Client.MinSkillID, r.Client.MinSkillID*10, 0)
// 			r.CSendChatMsg(0, "@level 160")
// 			time.Sleep(1 * time.Second)
// 			//循环获取坐骑道具（id 范围 601~620），再激活并穿戴
// 			// rideID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(20) + 1
// 			// addItemsInRange(r, 601, 621)
// 			// time.Sleep(1 * time.Second)
// 			// r.CRideAction(int32(rideID), 1) // type=1 激活坐骑
// 			// time.Sleep(500 * time.Millisecond)
// 			// r.CRideAction(int32(rideID), 5) // type=5 穿戴坐骑
// 			// time.Sleep(500 * time.Millisecond)
// 			// r.CRideAction(int32(rideID), 4) // type=4 骑乘

// 			// addPetsInRange(r, 1, 35)

// 			time.Sleep(1 * time.Second)
// 			r.CPetSummon(3)
// 			waitPetUpdate(r) // 等 s_pet_update_1 回调，服务端已生成 serise_id
// 			petSeriesID := pickPetSeriesID(r)
// 			//logger.Error(fmt.Sprintf("上阵宠物 serise_id: %d", petSeriesID))
// 			if petSeriesID != 0 {
// 				logger.Error(fmt.Sprintf("上阵宠物1 serise_id: %d", petSeriesID))
// 				// petID := int32(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(35) + 1)
// 				r.CPetBattle(petSeriesID, 1) // c_pet_battle_7 穿戴宠物（index=1）
// 			}
// 			//登录死斗玩法
// 			// r.CChaosBattleFieldEnter()
// 			// r.CReqClientKickSelf()
// 			// r.CReqCheckPassword(120, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 120, 120, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
// 			// //r.CReqCheckPassword(1003, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 1003, 1003, "47.96.18.95", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
// 			// // r.CReqCheckPassword(124, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 124, 124, "10.238.20.165", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
// 			// if r.Client.ActorID == 0 {
// 			// 	r.CReqCreateActor(1, 1, fmt.Sprintf("L%d", r.Index), 1, "e10adc3949ba59abbe56e057f20f883e")
// 			// 	r.CReqGetActorList()
// 			// }

// 			// r.CReqClientReadyLogin(r.Client.ActorID)
// 			// r.CReqActorLogin(r.Client.AccountID, r.Client.ActorID)
// 			// time.Sleep(1 * time.Second)
// 			break
// 		}
// 	}
// }

// syncedEntityRefreshInterval 同步实体目标的刷新间隔：每 5 秒从 Entities 缓存重新取样一次
const syncedEntityRefreshInterval = 1 * time.Second

// pickSyncedEntityHandle 从后端同步回来的实体缓存（s_entity_set_dir_2 / s_entity_move_1 /
// s_sync_entity_postion_4 等回调写入 Client.Entities）中选取一个非自己的实体 handle；
// 没有可用实体时返回 0
func pickSyncedEntityHandle(r *module.Robot) int64 {
	if r == nil || r.Client == nil || len(r.Client.Entities) == 0 {
		return 0
	}
	for handle := range r.Client.Entities {
		if handle == 0 || handle == r.Client.Handle {
			continue
		}
		return handle
	}
	return 0
}

// nextSyncedEntityHandle 每 5 秒获取一次后端同步回来的实体并返回其 handle 作为下次技能释放目标，
// 未到刷新间隔时复用上一次的结果；期间实体消失（removeEntity 已将其移出缓存）则返回 0。
// 返回的 handle 可直接用作 c_skill_frame_proxy_5 的 TargetHandler 或 c_damage_entity_3 的 handle：
//
//	targetHandle := nextSyncedEntityHandle(r)
//	if targetHandle != 0 {
//	    r.CStartUseSkill(r.Client.Handle, r.Client.MinSkillID, 0, 0)
//	    r.CSkillFrameProxy(module.SkillFrameProxyRequest{...}) // TargetHandler: fmt.Sprintf("%d", targetHandle)
//	    r.CDamageEntity(targetHandle, skillID, rangeIdx, 0)
//	}
func nextSyncedEntityHandle(r *module.Robot) int64 {
	if r == nil || r.Client == nil {
		return 0
	}
	now := time.Now()
	if r.Client.SyncedTargetHandle != 0 && now.Sub(r.Client.SyncedTargetAt) < syncedEntityRefreshInterval {
		return r.Client.SyncedTargetHandle
	}
	handle := pickSyncedEntityHandle(r)
	prev := r.Client.SyncedTargetHandle
	r.Client.SyncedTargetAt = now
	if handle != prev {
		r.Client.SyncedTargetHandle = handle
	}
	return handle
}
