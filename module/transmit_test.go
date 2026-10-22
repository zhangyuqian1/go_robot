package module

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stressTest/gameproto/defaultmsg"

	"github.com/gorilla/websocket"
)

func TestRouteStartTransmit(t *testing.T) {
	r := newTestRobot()
	r.Client.TransmitTarget = make(chan TransmitTarget, 1)
	r.routeGatewayResponse(0, 5, mustProtoBytes(t, &defaultmsg.SStartTransmit_5{Addr: "10.238.20.96", Port: 21100}))
	select {
	case target := <-r.Client.TransmitTarget:
		if target.Addr != "10.238.20.96" || target.Port != 21100 {
			t.Fatalf("unexpected target: %+v", target)
		}
	default:
		t.Fatal("transmit target was not delivered")
	}
}

func TestStartTransmitRejectsInvalidTarget(t *testing.T) {
	for _, response := range []*defaultmsg.SStartTransmit_5{
		{Port: 21100},
		{Addr: "127.0.0.1", Port: 0},
		{Addr: "127.0.0.1", Port: -1},
		{Addr: "127.0.0.1", Port: 65536},
	} {
		r := newTestRobot()
		r.Client.TransmitTarget = make(chan TransmitTarget, 1)
		r.handleStartTransmit(mustProtoBytes(t, response))
		if len(r.Client.TransmitTarget) != 0 {
			t.Fatalf("invalid target delivered: %+v", response)
		}
	}
}

func TestStartTransmitMalformedAndUnwatched(t *testing.T) {
	r := newTestRobot()
	r.handleStartTransmit([]byte{0xff})
	r.handleStartTransmit(mustProtoBytes(t, &defaultmsg.SStartTransmit_5{Addr: "127.0.0.1", Port: 21100}))
	r.Client.TransmitTarget = make(chan TransmitTarget, 1)
	for attempt := 0; attempt < 2; attempt++ {
		r.handleStartTransmit(mustProtoBytes(t, &defaultmsg.SStartTransmit_5{Addr: "127.0.0.1", Port: 21100}))
	}
	if len(r.Client.TransmitTarget) != 1 {
		t.Fatal("expected one buffered target")
	}
}

func TestActorEnterSceneNotifiesLogin(t *testing.T) {
	r := newTestRobot()
	r.Client.LoginSceneReady = make(chan struct{}, 1)
	for attempt := 0; attempt < 2; attempt++ {
		r.routeGatewayResponse(defaultSys, defaultActorEnterScene, mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{Sid: 42}))
	}
	select {
	case <-r.Client.LoginSceneReady:
		if !r.Client.LoginDone || r.Client.SceneID != 42 {
			t.Fatal("login notification arrived before state was cached")
		}
	default:
		t.Fatal("missing login notification")
	}
}

func TestActorEnterSceneDeliversLatestSnapshot(t *testing.T) {
	r := newTestRobot()
	r.Client.SceneEntries = make(chan SceneEntry, 1)
	r.handleActorEnterScene(mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{Sid: 1, Px: 7250, Py: 7450}))
	r.handleActorEnterScene(mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{Sid: 42, Fid: 7, Px: 1150, Py: 550}))
	r.Client.PX, r.Client.PY = 6850, 3200
	select {
	case entry := <-r.Client.SceneEntries:
		want := SceneEntry{SceneID: 42, FubenID: 7, X: 1150, Y: 550}
		if entry != want {
			t.Fatalf("snapshot = %+v, want %+v", entry, want)
		}
	default:
		t.Fatal("missing target scene snapshot")
	}
	if len(r.Client.SceneEntries) != 0 {
		t.Fatal("stale login scene snapshot retained")
	}
	r.handleActorEnterScene([]byte{0xff})
	if len(r.Client.SceneEntries) != 0 {
		t.Fatal("malformed scene must not notify the waiter")
	}
}

func TestReconnectAfterStoppingOldConnection(t *testing.T) {
	upgrader := websocket.Upgrader{}
	packets := make(chan struct{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{1, 2, 3, 4}); err != nil {
			return
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
			packets <- struct{}{}
		}
	}))
	defer server.Close()
	r := newTestRobot()
	r.Client.WebSocketSvr = "ws" + strings.TrimPrefix(server.URL, "http")
	defer r.Close()
	for attempt := 0; attempt < 2; attempt++ {
		r.TickStopSignal = make(chan struct{})
		r.ConnStopSignal = make(chan struct{})
		r.AsyncMonitSignal = make(chan struct{})
		if err := r.CreateWebSocketConn(); err != nil {
			t.Fatal(err)
		}
		if err := r.HeartBeatReq(); err != nil {
			t.Fatalf("connection %d: %v", attempt, err)
		}
		select {
		case <-packets:
		case <-time.After(time.Second):
			t.Fatal("new connection did not receive request")
		}
		r.Close()
	}
}
