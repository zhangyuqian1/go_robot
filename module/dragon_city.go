package module

import (
	"stressTest/gameproto/dragoncity"
	"stressTest/logger"
)

const (
	dragonCitySys              byte = 106
	dragonCityEnter            byte = 2
	dragonCityEnterRequestName      = "CDragonCityEnter"
)

func (r *Robot) CDragonCityEnter() {
	r.AsyncBegin(dragonCityEnterRequestName, 30, false)
	if err := r.sendGatewayRequest(dragonCityEnterRequestName, dragonCitySys, dragonCityEnter, &dragoncity.CDragonCityEnter_2{}, false); err != nil {
		r.Mark.Remove(dragonCityEnterRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(dragonCityEnterRequestName)
}
