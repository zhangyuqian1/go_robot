package module

import (
	"stressTest/gameproto/friend"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	friendSys        byte  = 14
	friendAddSocial  byte  = 1
	friendSocialList byte  = 1 // s_social_list_1 与 c_add_social_1 同消息号

	friendRequestName    = "CAddSocial"
	friendRequestTimeout int64 = 10
)

func (r *Robot) CAddSocial(aid int32, socialType int32) {
	req := &friend.CAddSocial_1{
		Aid:  aid,
		Type: socialType,
	}
	r.AsyncBegin(friendRequestName, friendRequestTimeout, true)
	if err := r.sendGatewayRequest(friendRequestName, friendSys, friendAddSocial, req, false); err != nil {
		r.Mark.Remove(friendRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(friendRequestName)
}

func (r *Robot) handleFriendGatewayResponse(sys byte, msgID byte, body []byte) bool {
	if sys != friendSys {
		return false
	}
	switch msgID {
	case friendSocialList:
		r.handleSocialList(body)
		return true
	default:
		return false
	}
}

func (r *Robot) handleSocialList(body []byte) {
	var resp friend.SSocialList_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		r.AsyncFire(friendRequestName, 0, STEP_FAIL, 0)
		return
	}

	if r.Client != nil {
		list := make([]uint64, 0, len(resp.GetList()))
		for _, info := range resp.GetList() {
			if info == nil || info.GetBasic() == nil {
				continue
			}
			list = append(list, uint64(info.GetBasic().GetAid()))
		}
		r.Client.FriendIDList = list
	}
	r.AsyncFire(friendRequestName, int64(len(body)), STEP_SUCC, 0)
}
