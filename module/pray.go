package module

import (
	"stressTest/gameproto/activity"
	"stressTest/logger"
)

const (
	activitySys       byte  = 102 // 循环活动系统号
	loopActPraySearch byte  = 4   // c_loop_act_pray_search_4 祈福抽卡

	prayRequestName    = "CLoopActPraySearch"
	prayRequestTimeout int64 = 10
)

// CPraySearch 祈福抽卡
func (r *Robot) CPraySearch(count int32) {
	req := &activity.CLoopActPraySearch_4{
		Count: count,
	}
	r.AsyncBegin(prayRequestName, prayRequestTimeout, false)
	if err := r.sendGatewayRequest(prayRequestName, activitySys, loopActPraySearch, req, false); err != nil {
		r.Mark.Remove(prayRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(prayRequestName)
}
