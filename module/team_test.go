package module

import (
	"testing"

	"stressTest/gameproto/defaultmsg"
	"stressTest/gameproto/team"
	"stressTest/gateway"

	"google.golang.org/protobuf/proto"
)

func TestCTeamDungeonMatchSendsWithoutResponse(t *testing.T) {
	key := [4]byte{1, 2, 3, 4}
	levels := []int32{1, 2}
	conn, packets, cleanup := newTestWebSocketPacketCollector(t, len(levels)+1)
	t.Cleanup(cleanup)
	r := newTestRobot()
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.IsWebConnected = true
	r.Client.GatewayCodec = gateway.NewCodec(key)

	for _, level := range levels {
		r.CTeamDungeonMatch(level)
		packet := readGatewayPacket(t, packets, key)
		if packet.Sys != 20 || packet.MsgID != 16 {
			t.Fatalf("team dungeon match route = %d/%d, want 20/16", packet.Sys, packet.MsgID)
		}
		var req team.CTeamDungeonMatch_16
		if err := proto.Unmarshal(packet.ProtoBytes, &req); err != nil {
			t.Fatalf("unmarshal team dungeon match request: %v", err)
		}
		if req.GetLv() != level {
			t.Fatalf("match level = %d, want %d", req.GetLv(), level)
		}
		if !r.Mark.IsEmpty() {
			t.Fatal("team dungeon match must not leave a pending response mark")
		}
	}
}

func TestCTeamDungeonMatchWithoutConnection(t *testing.T) {
	r := newTestRobot()
	r.CTeamDungeonMatch(1)
	if !r.Mark.IsEmpty() {
		t.Fatal("failed team dungeon match must not leave a pending response mark")
	}
}

func TestTeamDungeonMatchReplayAndTransmit(t *testing.T) {
	key := [4]byte{1, 2, 3, 4}
	conn, packets, cleanup := newTestWebSocketPacketCollector(t, 2)
	t.Cleanup(cleanup)
	r := newTestRobot()
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.IsWebConnected = true
	r.Client.GatewayCodec = gateway.NewCodec(key)
	r.Client.TeamDungeonMatchReady = make(chan int32, 1)
	r.Client.TransmitTarget = make(chan TransmitTarget, 1)
	r.CTeamDungeonMatch(1)
	packet := readGatewayPacket(t, packets, key)
	if packet.Sys != 20 || packet.MsgID != 16 {
		t.Fatalf("match route = %d/%d", packet.Sys, packet.MsgID)
	}
	r.routeGatewayResponse(20, 14, mustProtoBytes(t, &team.STeamDungeonMatchSuccess_14{Id: 37}))
	var matchID int32
	select {
	case matchID = <-r.Client.TeamDungeonMatchReady:
	default:
		t.Fatal("missing match notification")
	}
	r.CTeamDungeonReplay(matchID, 1)
	packet = readGatewayPacket(t, packets, key)
	if packet.Sys != 20 || packet.MsgID != 19 {
		t.Fatalf("replay route = %d/%d, want 20/19", packet.Sys, packet.MsgID)
	}
	var req team.CTeamDungeonReplay_19
	if err := proto.Unmarshal(packet.ProtoBytes, &req); err != nil {
		t.Fatal(err)
	}
	if req.GetId() != 37 || req.GetReplay() != 1 {
		t.Fatalf("replay must use matched id, got %v", &req)
	}
	mark, exists := r.Mark.Get(teamDungeonReplayRequestName)
	if !exists {
		t.Fatal("replay must wait for transmit push")
	}
	r.routeGatewayResponse(0, 5, mustProtoBytes(t, &defaultmsg.SStartTransmit_5{Addr: "127.0.0.1", Port: 21100}))
	select {
	case target := <-r.Client.TransmitTarget:
		if target.Addr != "127.0.0.1" || target.Port != 21100 {
			t.Fatalf("unexpected transmit target: %+v", target)
		}
	default:
		t.Fatal("missing transmit notification")
	}
	if r.Mark.Has(teamDungeonReplayRequestName) || mark.Code != STEP_SUCC {
		t.Fatal("transmit push did not complete replay")
	}
}

func TestTeamDungeonMatchPushValidation(t *testing.T) {
	for _, body := range [][]byte{
		{0xff},
		nil,
		mustProtoBytes(t, &team.STeamDungeonMatchSuccess_14{Id: -1}),
	} {
		r := newTestRobot()
		r.Client.TeamDungeonMatchReady = make(chan int32, 1)
		r.routeGatewayResponse(20, 14, body)
		if len(r.Client.TeamDungeonMatchReady) != 0 {
			t.Fatalf("invalid match push delivered: %x", body)
		}
	}
}

func TestTeamDungeonMatchPushDoesNotBlock(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &team.STeamDungeonMatchSuccess_14{Id: 7})
	r.routeGatewayResponse(20, 14, body)
	r.Client.TeamDungeonMatchReady = make(chan int32, 1)
	r.routeGatewayResponse(20, 14, body)
	r.routeGatewayResponse(20, 14, body)
	if len(r.Client.TeamDungeonMatchReady) != 1 {
		t.Fatal("duplicate match push should not block or replace pending notification")
	}
	if r.handleTeamGatewayResponse(0, 14, body) || r.handleTeamGatewayResponse(20, 15, nil) {
		t.Fatal("team handler consumed unrelated push")
	}
}

func TestTeamDungeonReplayRejectsInvalidTransmit(t *testing.T) {
	for _, body := range [][]byte{
		{0xff},
		mustProtoBytes(t, &defaultmsg.SStartTransmit_5{Addr: "127.0.0.1"}),
	} {
		r := newTestRobot()
		r.Client.TransmitTarget = make(chan TransmitTarget, 1)
		r.AsyncBegin(teamDungeonReplayRequestName, 30, false)
		mark, _ := r.Mark.Get(teamDungeonReplayRequestName)
		r.routeGatewayResponse(0, 5, body)
		if r.Mark.Has(teamDungeonReplayRequestName) || mark.Code != STEP_FAIL {
			t.Fatal("invalid transmit must fail replay request")
		}
		if len(r.Client.TransmitTarget) != 0 {
			t.Fatal("invalid transmit target delivered")
		}
	}
}

func TestCTeamDungeonReplayWithoutConnection(t *testing.T) {
	r := newTestRobot()
	r.CTeamDungeonReplay(3, 1)
	if !r.Mark.IsEmpty() {
		t.Fatal("failed replay must not leave a pending mark")
	}
}
