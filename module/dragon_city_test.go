package module

import (
	"testing"

	"stressTest/gameproto/defaultmsg"
	"stressTest/gameproto/dragoncity"
	"stressTest/gameproto/observer"
	"stressTest/gateway"

	"google.golang.org/protobuf/proto"
)

func TestDragonCityDeathAndAutomaticRevival(t *testing.T) {
	r := newTestRobot()
	r.Client.Handle = 123
	r.Client.SceneEntries = make(chan SceneEntry, 1)
	r.Client.BattleTargetHandle = 456
	r.handleEntityDeath(mustProtoBytes(t, &observer.SEntityDeath_8{Handle: 999}))
	if r.Client.SelfDead {
		t.Fatal("another entity death must not stop our route")
	}
	r.AsyncBegin(movementRequestName, 10, true)
	mark, _ := r.Mark.Get(movementRequestName)
	r.routeGatewayResponse(16, 8, mustProtoBytes(t, &observer.SEntityDeath_8{Handle: 123}))
	if !r.Client.SelfDead || r.Client.BattleTargetHandle != 0 || mark.Code != STEP_FAIL || r.Mark.Has(movementRequestName) {
		t.Fatal("self death must mark dead and release the pending movement")
	}
	select {
	case <-mark.stopChan:
	default:
		t.Fatal("movement waiter not released on death")
	}
	r.routeGatewayResponse(0, 13, mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{Sid: 1032, Px: 1150, Py: 450}))
	if r.Client.SelfDead || r.Client.PX != 1150 || r.Client.PY != 450 {
		t.Fatal("automatic revival must reset death state and cache spawn")
	}
	select {
	case entry := <-r.Client.SceneEntries:
		if entry != (SceneEntry{SceneID: 1032, X: 1150, Y: 450}) {
			t.Fatalf("revival snapshot = %+v", entry)
		}
	default:
		t.Fatal("missing automatic revival snapshot")
	}
}

func TestDragonCityEnterAndTransmit(t *testing.T) {
	key := [4]byte{1, 2, 3, 4}
	conn, packets, cleanup := newTestWebSocketPacketCollector(t, 1)
	t.Cleanup(cleanup)
	r := newTestRobot()
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.IsWebConnected = true
	r.Client.GatewayCodec = gateway.NewCodec(key)
	r.Client.TransmitTarget = make(chan TransmitTarget, 1)
	r.CDragonCityEnter()
	packet := readGatewayPacket(t, packets, key)
	if packet.Sys != 106 || packet.MsgID != 2 {
		t.Fatalf("enter route = %d/%d, want 106/2", packet.Sys, packet.MsgID)
	}
	var req dragoncity.CDragonCityEnter_2
	if err := proto.Unmarshal(packet.ProtoBytes, &req); err != nil {
		t.Fatal(err)
	}
	if len(packet.ProtoBytes) != 0 {
		t.Fatal("dragon city enter must be an empty request")
	}
	mark, exists := r.Mark.Get(dragonCityEnterRequestName)
	if !exists {
		t.Fatal("enter must wait for transmit push")
	}
	r.routeGatewayResponse(0, 5, mustProtoBytes(t, &defaultmsg.SStartTransmit_5{Addr: "10.238.20.96", Port: 21100}))
	if r.Mark.Has(dragonCityEnterRequestName) || mark.Code != STEP_SUCC {
		t.Fatal("transmit push did not complete dragon city enter")
	}
	select {
	case target := <-r.Client.TransmitTarget:
		if target.Addr != "10.238.20.96" || target.Port != 21100 {
			t.Fatalf("unexpected target: %+v", target)
		}
	default:
		t.Fatal("missing transmit notification")
	}
}

func TestDragonCityEnterInvalidTransmit(t *testing.T) {
	for _, body := range [][]byte{
		{0xff},
		nil,
		mustProtoBytes(t, &defaultmsg.SStartTransmit_5{Addr: "localhost", Port: 65536}),
	} {
		r := newTestRobot()
		r.Client.TransmitTarget = make(chan TransmitTarget, 1)
		r.AsyncBegin(dragonCityEnterRequestName, 30, false)
		mark, _ := r.Mark.Get(dragonCityEnterRequestName)
		r.routeGatewayResponse(0, 5, body)
		if r.Mark.Has(dragonCityEnterRequestName) || mark.Code != STEP_FAIL {
			t.Fatal("invalid transmit must fail dragon city enter")
		}
		if len(r.Client.TransmitTarget) != 0 {
			t.Fatal("invalid transmit must not trigger reconnect")
		}
	}
}

func TestDragonCityEnterSendFailure(t *testing.T) {
	r := newTestRobot()
	r.CDragonCityEnter()
	if r.Mark.Has(dragonCityEnterRequestName) {
		t.Fatal("send failure must remove enter mark")
	}
}
