package main

import (
	"stressTest/gamecase"
	"stressTest/global"
	perf "stressTest/jasmine"
)

func CaseInit() {
	global.CaseMap = map[string]*perf.Task{
		"LoginOnline":    {Name: "登入在线", Weight: 1, Fn: gamecase.LoginOnline},
		"Move":           {Name: "移动", Weight: 1, Fn: gamecase.Move},
		"Battle":         {Name: "战斗", Weight: 1, Fn: gamecase.Battle},
		"PostLoginMixed": {Name: "登录后混合", Weight: 1, Fn: gamecase.PostLoginMixed},
		"混合场景":           {Name: "登录后混合", Weight: 1, Fn: gamecase.PostLoginMixed},
		"Draw":           {Name: "技能抽卡", Weight: 1, Fn: gamecase.Draw},
		"Friend":         {Name: "加好友", Weight: 1, Fn: gamecase.Friend},
		"Pray":           {Name: "循环活动祈福", Weight: 1, Fn: gamecase.Pray},
	}

	global.MixedScene = []*perf.Task{
		global.CaseMap["Move"],
		global.CaseMap["Battle"],
		global.CaseMap["Pet"],
		global.CaseMap["Equip"],
		global.CaseMap["Arena"],
		global.CaseMap["Draw"],
	}
}
