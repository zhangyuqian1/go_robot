package module

import (
	"testing"

	"stressTest/gameproto/pk"
	"stressTest/gateway"

	"google.golang.org/protobuf/proto"
)

func TestSetPkModeGatewayRequest(t *testing.T) {
	key := [4]byte{1, 2, 3, 4}
	conn, packets, cleanup := newTestWebSocketPacketCollector(t, 1)
	t.Cleanup(cleanup)
	r := newTestRobot()
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.IsWebConnected = true
	r.Client.GatewayCodec = gateway.NewCodec(key)
	r.CSetPkMode(5)
	packet := readGatewayPacket(t, packets, key)
	if packet.Sys != 18 || packet.MsgID != 1 {
		t.Fatalf("PK mode route = %d/%d, want 18/1", packet.Sys, packet.MsgID)
	}
	var request pk.CSetPkMode_1
	if err := proto.Unmarshal(packet.ProtoBytes, &request); err != nil {
		t.Fatal(err)
	}
	if request.GetMode() != 5 {
		t.Fatalf("PK mode = %d, want 5", request.GetMode())
	}
	if r.Mark.Has(pkSetModeRequestName) {
		t.Fatal("request without response must not leave a pending mark")
	}
}

func TestSetPkModeDisconnected(t *testing.T) {
	r := newTestRobot()
	r.CSetPkMode(5)
	if r.Mark.Has(pkSetModeRequestName) {
		t.Fatal("failed request must not leave a pending mark")
	}
}
