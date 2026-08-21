package module

import (
	"testing"
	"time"

	"stressTest/gameproto/common"
	"stressTest/gameproto/friend"
)

func TestHandleSocialListCachesFriendIDs(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin(friendRequestName, friendRequestTimeout, false)
	r.routeGatewayResponse(friendSys, friendSocialList, mustProtoBytes(t, &friend.SSocialList_1{
		List: []*friend.SSocialList_1Info{
			{Basic: &common.ActorBasic{Aid: 111}},
			{Basic: &common.ActorBasic{Aid: 222}},
		},
	}))

	if len(r.Client.FriendIDList) != 2 {
		t.Fatalf("FriendIDList = %v, want 2 entries", r.Client.FriendIDList)
	}
	if r.Client.FriendIDList[0] != 111 || r.Client.FriendIDList[1] != 222 {
		t.Fatalf("FriendIDList = %v, want [111 222]", r.Client.FriendIDList)
	}
}

func TestHandleSocialListUnblocksAddSocialWait(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin(friendRequestName, friendRequestTimeout, true)
	mark, ok := r.Mark.Get(friendRequestName)
	if !ok {
		t.Fatalf("%s mark missing", friendRequestName)
	}
	done := make(chan struct{})
	go func() {
		<-mark.stopChan
		close(done)
	}()

	r.routeGatewayResponse(friendSys, friendSocialList, mustProtoBytes(t, &friend.SSocialList_1{}))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("CAddSocial wait not unblocked by s_social_list_1")
	}
}

func TestHandleSocialListMalformedUnblocksAddSocialWait(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin(friendRequestName, friendRequestTimeout, true)
	mark, ok := r.Mark.Get(friendRequestName)
	if !ok {
		t.Fatalf("%s mark missing", friendRequestName)
	}
	done := make(chan struct{})
	go func() {
		<-mark.stopChan
		close(done)
	}()

	r.routeGatewayResponse(friendSys, friendSocialList, []byte{1, 2, 3})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("malformed s_social_list_1 should still fire CAddSocial")
	}
}

func TestFriendRouteIgnoresUnknownSysAndMsg(t *testing.T) {
	r := newTestRobot()
	if r.handleFriendGatewayResponse(15, friendSocialList, nil) {
		t.Fatalf("handleFriendGatewayResponse should ignore non-friend sys")
	}
	if r.handleFriendGatewayResponse(friendSys, 9, nil) {
		t.Fatalf("handleFriendGatewayResponse should ignore unknown msgID")
	}
}
