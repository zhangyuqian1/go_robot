package module

import (
	"stressTest/gameproto/misc"
	"stressTest/logger"
)

const (
	rideSys    byte = 15
	rideAction byte = 14

	rideActionRequestName = "CRideAction"
)

func (r *Robot) CRideAction(rideID int32, actionType int32) {
	req := &misc.CRideAction_14{Id: rideID, Type: actionType}
	if err := r.sendGatewayRequest(rideActionRequestName, rideSys, rideAction, req, false); err != nil {
		logger.Error(err.Error())
		return
	}
}
