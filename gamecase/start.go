package gamecase

import (
	"fmt"
	"math/rand"
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
	deadline := time.Now().Add(5 * time.Second)
	for r.Client.WeaponSeries == nil {
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
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
			r.CReqCheckPassword(120, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 120, 120, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
			if r.Client.ActorID == 0 {
				r.CReqCreateActor(1, 1, fmt.Sprintf("L%d", r.Index), 1, "e10adc3949ba59abbe56e057f20f883e")
				r.CReqGetActorList()
			}
			r.CReqClientReadyLogin(r.Client.ActorID)
			r.CReqActorLogin(r.Client.AccountID, r.Client.ActorID)
			// time.Sleep(1 * time.Second)
			// r.CSendChatMsg(0, "@set 10066 41")
			// r.CSendChatMsg(0, "@set 2007 100000")
			// r.CSendChatMsg(0, "@set 2005 10000000")
			// r.CSendChatMsg(0, "@set 2008 100000")
			// r.CSendChatMsg(0, "@set 2009 100000")
			// r.CSendChatMsg(0, "@set 2010 100000")
			// r.CSendChatMsg(0, "@additem 950 1")
			// time.Sleep(1 * time.Second)
			// waitGiftSeries(r)
			// logger.Error(fmt.Sprintf("序列号1: %v", r.Client.GiftSeries))
			// r.CUseBagGift(3, r.Client.GiftSeries, 1)
			// // 随机选择 2~25 的英雄 id，传给后续两个协议
			// heroID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(24) + 2
			// r.CActivityHero(int32(heroID))
			// r.CSwitchHero(int32(heroID))

			weaponID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(120) + 1001
			r.CSendChatMsg(0, fmt.Sprintf("@additem %d 1", weaponID))
			time.Sleep(1 * time.Second)
			waitWeaponSeries(r)
			logger.Error(fmt.Sprintf("序列号2: %v", r.Client.WeaponSeries))
			r.CTakeOnEquip(3, r.Client.WeaponSeries)
			// r.CSendChatMsg(0, "@level 160")
			// time.Sleep(1 * time.Second)
			// r.Clogin_skill_draw(0, false)
			break
		}
	}

}
