package module

import (
	"stressTest/gameproto/hero"
	"stressTest/logger"
)

const (
	heroSys          byte = 17 // 英雄系统号
	heroSwitchHero   byte = 1  // c_switch_hero_1 切换英雄
	heroActivityHero byte = 2  // c_activity_hero_2 激活英雄

	heroSwitchRequestName         = "CSwitchHero"
	heroActivityRequestName       = "CActivityHero"
	heroRequestTimeout      int64 = 10
)

// CSwitchHero 切换英雄
func (r *Robot) CSwitchHero(heroID int32) {
	req := &hero.CSwitchHero_1{HeroId: heroID}
	r.AsyncBegin(heroSwitchRequestName, heroRequestTimeout, false)
	if err := r.sendGatewayRequest(heroSwitchRequestName, heroSys, heroSwitchHero, req, false); err != nil {
		r.Mark.Remove(heroSwitchRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(heroSwitchRequestName)
}

// CActivityHero 激活英雄
func (r *Robot) CActivityHero(heroID int32) {
	req := &hero.CActivityHero_2{HeroId: heroID}
	r.AsyncBegin(heroActivityRequestName, heroRequestTimeout, false)
	if err := r.sendGatewayRequest(heroActivityRequestName, heroSys, heroActivityHero, req, false); err != nil {
		r.Mark.Remove(heroActivityRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(heroActivityRequestName)
}
