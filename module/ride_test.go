package module

import (
	"testing"

	"stressTest/gameproto/misc"
	"stressTest/gateway"

	"google.golang.org/protobuf/proto"
)

func TestCRideActionSendsWithoutResponse(t *testing.T) {
	key := [4]byte{1, 2, 3, 4}
	actions := []int32{1, 5, 4}
	conn, packets, cleanup := newTestWebSocketPacketCollector(t, len(actions)+1)
	t.Cleanup(cleanup)
	r := newTestRobot()
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.IsWebConnected = true
	r.Client.GatewayCodec = gateway.NewCodec(key)

	for _, actionType := range actions {
		r.CRideAction(2, actionType)
		packet := readGatewayPacket(t, packets, key)
		if packet.Sys != 15 || packet.MsgID != 14 {
			t.Fatalf("ride packet route = %d/%d, want 15/14", packet.Sys, packet.MsgID)
		}
		var req misc.CRideAction_14
		if err := proto.Unmarshal(packet.ProtoBytes, &req); err != nil {
			t.Fatalf("unmarshal ride request: %v", err)
		}
		if req.GetId() != 2 || req.GetType() != actionType {
			t.Fatalf("ride request = id:%d type:%d, want 2/%d", req.GetId(), req.GetType(), actionType)
		}
		if !r.Mark.IsEmpty() {
			t.Fatal("ride action must not leave a pending response mark")
		}
	}
}

func TestCRideActionWithoutConnection(t *testing.T) {
	r := newTestRobot()
	r.CRideAction(2, 1)
	if !r.Mark.IsEmpty() {
		t.Fatal("failed ride action must not leave a pending response mark")
	}
}
