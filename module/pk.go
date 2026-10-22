package module

import (
	"stressTest/gameproto/pk"
	"stressTest/logger"
)

const (
	pkSys                byte = 18
	pkSetMode            byte = 1
	pkSetModeRequestName      = "CSetPkMode"
)

func (r *Robot) CSetPkMode(mode int32) {
	if err := r.sendGatewayRequest(pkSetModeRequestName, pkSys, pkSetMode, &pk.CSetPkMode_1{Mode: mode}, false); err != nil {
		logger.Error(err.Error())
	}
}
