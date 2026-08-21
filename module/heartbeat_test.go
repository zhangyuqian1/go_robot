package module

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stressTest/gameproto/defaultmsg"
	"stressTest/gateway"
	"stressTest/global"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func withHeartbeatInterval(t *testing.T, interval time.Duration) {
	t.Helper()
	original := global.HeartbeatInterval
	global.HeartbeatInterval = interval
	t.Cleanup(func() {
		global.HeartbeatInterval = original
	})
}

func TestStartHeartBeatRequiresLoginDone(t *testing.T) {
	withHeartbeatInterval(t, time.Hour)
	r := newTestRobot()
	r.TickStopSignal = make(chan struct{})

	r.StartHeartBeat()
	t.Cleanup(r.StopHeartBeat)

	if r.Ticking || r.Client.PingSending {
		t.Fatalf("heartbeat started before login completion")
	}
}

func TestActorEnterSceneStartsHeartbeatAfterLoginDone(t *testing.T) {
	withHeartbeatInterval(t, time.Hour)
	r := newTestRobot()
	r.TickStopSignal = make(chan struct{})
	r.Client.IsWebConnected = true

	r.routeGatewayResponse(defaultSys, defaultActorEnterScene, mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{Sid: 1}))
	t.Cleanup(r.StopHeartBeat)

	if !r.Client.LoginDone {
		t.Fatalf("LoginDone was not set by enter scene response")
	}
	if !r.Ticking || !r.Client.PingSending {
		t.Fatalf("heartbeat did not start after login completion")
	}
}

func TestCloseSocketSignalsHeartbeatStop(t *testing.T) {
	withHeartbeatInterval(t, time.Hour)
	r := newTestRobot()
	r.TickStopSignal = make(chan struct{})
	r.Client.LoginDone = true
	r.Client.IsWebConnected = true

	r.StartHeartBeat()
	t.Cleanup(r.StopHeartBeat)
	if !r.Ticking {
		t.Fatalf("heartbeat did not start")
	}

	r.CloseSocket()
	waitForHeartbeatStopped(t, r)
}

func TestStartHeartBeatRecreatesClosedStopSignal(t *testing.T) {
	withHeartbeatInterval(t, time.Hour)
	r := newTestRobot()
	r.TickStopSignal = make(chan struct{})
	r.Client.LoginDone = true
	r.Client.IsWebConnected = true

	r.StartHeartBeat()
	r.CloseSocket()
	waitForHeartbeatStopped(t, r)

	r.Client.IsWebConnected = true
	r.StartHeartBeat()
	t.Cleanup(r.StopHeartBeat)

	if !r.Ticking || !r.Client.PingSending {
		t.Fatalf("heartbeat did not restart after recreating stop signal")
	}
}

func TestBuildHeartbeatRequestUsesUnixTick(t *testing.T) {
	req := buildHeartbeatRequest(time.Unix(1234, 999999999))
	if req.GetTick() != 1234 {
		t.Fatalf("heartbeat tick = %d, want 1234", req.GetTick())
	}
}

func TestHeartBeatReqSendsDefaultHeartbeatPacket(t *testing.T) {
	key := [4]byte{1, 2, 3, 4}
	conn, packets, cleanup := newTestWebSocketPair(t)
	defer cleanup()

	r := newTestRobot()
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.IsWebConnected = true
	r.Client.GatewayCodec = gateway.NewCodec(key)

	if err := r.HeartBeatReq(); err != nil {
		t.Fatalf("HeartBeatReq error: %v", err)
	}

	var packetBytes []byte
	select {
	case packetBytes = <-packets:
	case <-time.After(time.Second):
		t.Fatalf("server did not receive heartbeat packet")
	}

	packet, err := gateway.NewCodec(key).UnmarshalResponse(packetBytes)
	if err != nil {
		t.Fatalf("decode heartbeat packet: %v", err)
	}
	if packet.Sys != defaultSys || packet.MsgID != defaultHeartbeatRequest {
		t.Fatalf("heartbeat packet route = %d/%d, want %d/%d", packet.Sys, packet.MsgID, defaultSys, defaultHeartbeatRequest)
	}
	var req defaultmsg.CHeartbeat_2
	if err := proto.Unmarshal(packet.ProtoBytes, &req); err != nil {
		t.Fatalf("unmarshal heartbeat request: %v", err)
	}
	if req.GetTick() == 0 {
		t.Fatalf("heartbeat tick was not populated")
	}
}

func waitForHeartbeatStopped(t *testing.T, r *Robot) {
	t.Helper()
	deadline := time.After(time.Second)
	for r.Ticking {
		select {
		case <-deadline:
			t.Fatalf("heartbeat did not stop")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func newTestWebSocketPair(t *testing.T) (*websocket.Conn, <-chan []byte, func()) {
	t.Helper()
	packets := make(chan []byte, 1)
	errs := make(chan error, 1)
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close()
		messageType, packet, err := conn.ReadMessage()
		if err != nil {
			errs <- err
			return
		}
		if messageType != websocket.BinaryMessage {
			errs <- fmt.Errorf("message type = %d, want binary", messageType)
			return
		}
		packets <- packet
	}))

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		server.Close()
		t.Fatalf("dial test websocket: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		server.Close()
		select {
		case err := <-errs:
			if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				t.Errorf("test websocket server error: %v", err)
			}
		default:
		}
	}
	return conn, packets, cleanup
}
