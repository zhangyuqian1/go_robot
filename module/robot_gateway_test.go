package module

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stressTest/gameproto/defaultmsg"
	"stressTest/gameproto/login"
	"stressTest/gameproto/property"
	"stressTest/gateway"
	"stressTest/logger"

	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
)

func init() {
	logger.Init("./logs/robot_test.log", zapcore.FatalLevel)
}

func newTestRobot() *Robot {
	return &Robot{
		PTName: "robot-test",
		Client: &GameClient{},
		Mark:   cmap.New[*Events](),
	}
}

func mustProtoBytes(t *testing.T, message proto.Message) []byte {
	t.Helper()
	body, err := proto.Marshal(message)
	if err != nil {
		t.Fatalf("proto.Marshal error: %v", err)
	}
	return body
}

func TestRouteGatewayResponseCachesKnownResponses(t *testing.T) {
	r := newTestRobot()
	r.routeGatewayResponse(loginSys, loginResult, mustProtoBytes(t, &login.SLoginResult_1{Code: 7}))
	if r.Client.LastLoginCode != 7 {
		t.Fatalf("LastLoginCode = %d, want 7", r.Client.LastLoginCode)
	}

	r.routeGatewayResponse(loginSys, loginActorList, mustProtoBytes(t, &login.SActorList_2{
		AccountId: 101,
		ActorCnt:  1,
		Actors: []*login.SActorList_2Actor{{
			ActorId: 202,
			Name:    "hero",
		}},
	}))
	if r.Client.AccountID != 101 || r.Client.ActorID != 202 || r.Client.ActorName != "hero" {
		t.Fatalf("actor list cache = account:%d actor:%d name:%q", r.Client.AccountID, r.Client.ActorID, r.Client.ActorName)
	}

	r.routeGatewayResponse(loginSys, loginCreateActorResult, mustProtoBytes(t, &login.SCreateActorResult_3{Aid: 303, Result: 0}))
	if r.Client.ActorID != 303 || r.Client.LastCreateActorResult != 0 {
		t.Fatalf("create actor cache = actor:%d result:%d", r.Client.ActorID, r.Client.LastCreateActorResult)
	}

	r.routeGatewayResponse(defaultSys, defaultActorLoginResult, mustProtoBytes(t, &defaultmsg.SActorLoginResult_1{Result: 11}))
	if r.Client.LastActorLoginResult != 11 {
		t.Fatalf("LastActorLoginResult = %d, want 11", r.Client.LastActorLoginResult)
	}

	r.routeGatewayResponse(propertySys, propertyActorInit, mustProtoBytes(t, &property.SActorInit_1{
		Handle:  404,
		Id:      505,
		RawSrv:  606,
		Name:    "cached-name",
		MapData: map[int32]int64{1: 2},
		CurSrv:  707,
	}))
	if r.Client.Handle != 404 || r.Client.ActorID != 505 || r.Client.RawServerID != 606 || r.Client.CurrentServerID != 707 {
		t.Fatalf("actor init cache = handle:%d actor:%d raw:%d cur:%d", r.Client.Handle, r.Client.ActorID, r.Client.RawServerID, r.Client.CurrentServerID)
	}
	if got := r.Client.MapData[1]; got != 2 {
		t.Fatalf("MapData[1] = %d, want 2", got)
	}

	r.routeGatewayResponse(defaultSys, defaultHeartbeat, mustProtoBytes(t, &defaultmsg.SHeartbeat_10{Tick: 808}))
	if r.Client.LastHeartbeatTick != 808 {
		t.Fatalf("LastHeartbeatTick = %d, want 808", r.Client.LastHeartbeatTick)
	}

	r.routeGatewayResponse(defaultSys, defaultActorEnterScene, mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{
		Sid:  909,
		Fid:  1001,
		Type: 3,
		Px:   1.5,
		Py:   2.5,
	}))
	if !r.Client.LoginDone || r.Client.SceneID != 909 || r.Client.FubenID != 1001 || r.Client.SceneType != 3 {
		t.Fatalf("scene cache = done:%v scene:%d fuben:%d type:%d", r.Client.LoginDone, r.Client.SceneID, r.Client.FubenID, r.Client.SceneType)
	}
	if r.Client.PX != 1.5 || r.Client.PY != 2.5 {
		t.Fatalf("position cache = %.1f/%.1f", r.Client.PX, r.Client.PY)
	}
}

func TestRouteGatewayResponseIgnoresUnknownAndMalformedResponses(t *testing.T) {
	r := newTestRobot()
	r.routeGatewayResponse(99, 88, []byte{1, 2, 3})
	r.routeGatewayResponse(loginSys, loginActorList, []byte{1, 2, 3})
	r.routeGatewayResponse(defaultSys, defaultActorEnterScene, []byte{1, 2, 3})

	if r.Client.LoginDone {
		t.Fatalf("LoginDone changed after unknown or malformed packet")
	}
}

func TestLoginResultErrorUnblocksLoginWait(t *testing.T) {
	r := newTestRobot()
	r.Client.LoginStarted = true
	r.AsyncBegin(loginRequestCheckPassword, 30, true)
	mark, ok := r.Mark.Get(loginRequestCheckPassword)
	if !ok {
		t.Fatalf("%s mark missing", loginRequestCheckPassword)
	}
	done := make(chan struct{})
	go func() {
		<-mark.stopChan
		close(done)
	}()

	r.routeGatewayResponse(loginSys, loginResult, mustProtoBytes(t, &login.SLoginResult_1{Code: 12}))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("login error did not unblock waiter")
	}
	finished, ok := r.Mark.Get(loginRequestCheckPassword)
	if ok && finished.Code != STEP_FAIL {
		t.Fatalf("%s code = %d, want STEP_FAIL", loginRequestCheckPassword, finished.Code)
	}
	if r.Client.LastLoginCode != 12 {
		t.Fatalf("LastLoginCode = %d, want 12", r.Client.LastLoginCode)
	}
}

func TestLoginRouteSupportsGetActorListAfterCreate(t *testing.T) {
	r := newTestRobot()
	r.Client.LoginStarted = true
	r.Client.CreatePending = true
	r.AsyncBegin(loginRequestGetActorList, 30, true)

	r.routeGatewayResponse(loginSys, loginActorList, mustProtoBytes(t, &login.SActorList_2{
		AccountId: 101,
		ActorCnt:  1,
		Actors: []*login.SActorList_2Actor{{
			ActorId: 202,
			Name:    "created",
		}},
	}))

	if r.Client.AccountID != 101 || r.Client.ActorID != 202 || r.Client.ActorName != "created" {
		t.Fatalf("actor list after create cache = account:%d actor:%d name:%q", r.Client.AccountID, r.Client.ActorID, r.Client.ActorName)
	}
	if r.Client.CreatePending {
		t.Fatalf("CreatePending should be cleared after c_get_actor_list_3 response")
	}
}

func TestActorLoginSendsReadyAndLoginWithoutWaiting(t *testing.T) {
	r := newTestRobot()
	r.Client.AccountID = 101
	r.Client.ActorID = 202
	key := [4]byte{1, 2, 3, 4}
	conn, packets, cleanup := newTestWebSocketPacketCollector(t, 2)
	t.Cleanup(cleanup)
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.IsWebConnected = true
	r.Client.GatewayCodec = gateway.NewCodec(key)

	r.CReqClientReadyLogin(r.Client.ActorID)
	r.CReqActorLogin(r.Client.AccountID, r.Client.ActorID)
	readyPacket := readGatewayPacket(t, packets, key)
	if readyPacket.Sys != defaultSys || readyPacket.MsgID != defaultClientReadyLogin {
		t.Fatalf("first login packet = sys:%d msg:%d, want %d/%d", readyPacket.Sys, readyPacket.MsgID, defaultSys, defaultClientReadyLogin)
	}
	actorPacket := readGatewayPacket(t, packets, key)
	if actorPacket.Sys != defaultSys || actorPacket.MsgID != defaultActorLoginResult {
		t.Fatalf("second login packet = sys:%d msg:%d, want %d/%d", actorPacket.Sys, actorPacket.MsgID, defaultSys, defaultActorLoginResult)
	}
	r.AsyncBegin(loginRequestEnterScene, loginRequestTimeout, true)
	r.routeGatewayResponse(defaultSys, defaultActorLoginResult, mustProtoBytes(t, &defaultmsg.SActorLoginResult_1{Result: 9}))
	if !r.Mark.Has(loginRequestEnterScene) {
		t.Fatalf("%s wait mark should remain until s_actor_enter_scene_13", loginRequestEnterScene)
	}
}

func TestActorLoginSuccessWaitsForEnterScene(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin(loginRequestActorLogin, loginRequestTimeout, false)

	r.routeGatewayResponse(defaultSys, defaultActorLoginResult, mustProtoBytes(t, &defaultmsg.SActorLoginResult_1{Result: 0}))
	if !r.Mark.Has(loginRequestActorLogin) {
		t.Fatalf("%s should wait for s_actor_enter_scene_13 after successful actor login result", loginRequestActorLogin)
	}

	r.routeGatewayResponse(defaultSys, defaultActorEnterScene, mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{Sid: 1}))
	if r.Mark.Has(loginRequestActorLogin) {
		t.Fatalf("%s should be fired by s_actor_enter_scene_13", loginRequestActorLogin)
	}
}

func TestActorEnterSceneFiresReadyAndActorLogin(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin(loginRequestReadyLogin, loginRequestTimeout, false)
	r.AsyncBegin(loginRequestActorLogin, loginRequestTimeout, false)

	r.routeGatewayResponse(defaultSys, defaultActorEnterScene, mustProtoBytes(t, &defaultmsg.SActorEnterScene_13{Sid: 1}))

	if r.Mark.Has(loginRequestReadyLogin) {
		t.Fatalf("%s should be fired by s_actor_enter_scene_13", loginRequestReadyLogin)
	}
	if r.Mark.Has(loginRequestActorLogin) {
		t.Fatalf("%s should be fired by s_actor_enter_scene_13", loginRequestActorLogin)
	}
}

func TestLoginResponseFailuresUseAsyncFireCode(t *testing.T) {
	tests := []struct {
		name        string
		requestName string
		route       func(*Robot)
		wantCode    uint16
		wantErrCode any
	}{
		{
			name:        "create actor result",
			requestName: loginRequestCreateActor,
			route: func(r *Robot) {
				r.routeGatewayResponse(loginSys, loginCreateActorResult, mustProtoBytes(t, &login.SCreateActorResult_3{Aid: 303, Result: 5}))
			},
			wantCode:    STEP_FAIL,
			wantErrCode: int32(5),
		},
		{
			name:        "actor login result",
			requestName: loginRequestActorLogin,
			route: func(r *Robot) {
				r.routeGatewayResponse(defaultSys, defaultActorLoginResult, mustProtoBytes(t, &defaultmsg.SActorLoginResult_1{Result: 6}))
			},
			wantCode:    STEP_FAIL,
			wantErrCode: int32(6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRobot()
			r.AsyncBegin(tt.requestName, 30, false)
			tt.route(r)
			mark, ok := r.Mark.Get(tt.requestName)
			if ok && (mark.Code != tt.wantCode || mark.ErrCode != tt.wantErrCode) {
				t.Fatalf("%s mark = code:%d err:%v, want code:%d err:%v", tt.requestName, mark.Code, mark.ErrCode, tt.wantCode, tt.wantErrCode)
			}
		})
	}
}

func TestActorListEmptyAfterCreateKeepsDirectCaseFlow(t *testing.T) {
	r := newTestRobot()
	r.Client.LoginStarted = true
	r.Client.CreatePending = true
	r.AsyncBegin(loginRequestGetActorList, 30, false)

	r.routeGatewayResponse(loginSys, loginActorList, mustProtoBytes(t, &login.SActorList_2{
		AccountId: 101,
		ActorCnt:  0,
	}))

	if r.Client.AccountID != 101 || r.Client.ActorID != 0 || r.Client.ActorName != "" {
		t.Fatalf("empty actor list cache = account:%d actor:%d name:%q", r.Client.AccountID, r.Client.ActorID, r.Client.ActorName)
	}
	mark, ok := r.Mark.Get(loginRequestGetActorList)
	if ok && mark.Code != STEP_SUCC {
		t.Fatalf("%s code = %d, want STEP_SUCC", loginRequestGetActorList, mark.Code)
	}
}

func TestStopPendingGatewayWaitsUnblocksWaiters(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin("WaitForRoute", 30, true)
	mark, ok := r.Mark.Get("WaitForRoute")
	if !ok {
		t.Fatalf("WaitForRoute mark missing")
	}
	done := make(chan struct{})
	go func() {
		<-mark.stopChan
		close(done)
	}()

	r.stopPendingGatewayWaits(STEP_FAIL, "closed")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("AsyncWait did not unblock after stopPendingGatewayWaits")
	}
	if r.Mark.Has("WaitForRoute") {
		t.Fatalf("pending wait was not removed")
	}
}

func TestAsyncWaitStillWaitsWhenMetricsRequestFails(t *testing.T) {
	r := newTestRobot()
	r.AsyncBegin("WaitWithMetricsError", 30, true)

	returned := make(chan struct{})
	go func() {
		r.AsyncWait("WaitWithMetricsError")
		close(returned)
	}()

	select {
	case <-returned:
		t.Fatalf("AsyncWait returned before response")
	case <-time.After(100 * time.Millisecond):
	}

	r.AsyncFire("WaitWithMetricsError", 0, STEP_SUCC, 0)

	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatalf("AsyncWait did not return after response")
	}
}

func TestSendGatewayRequestRequiresReadyWebSocket(t *testing.T) {
	r := newTestRobot()
	if err := r.sendGatewayRequest("missing", defaultSys, defaultHeartbeatRequest, &defaultmsg.CHeartbeat_2{}, false); err == nil {
		t.Fatalf("sendGatewayRequest returned nil without websocket")
	}
}

func newTestWebSocketPacketCollector(t *testing.T, packetCount int) (*websocket.Conn, <-chan []byte, func()) {
	t.Helper()
	packets := make(chan []byte, packetCount)
	errs := make(chan error, 1)
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close()
		for i := 0; i < packetCount; i++ {
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
		}
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

func readGatewayPacket(t *testing.T, packets <-chan []byte, key [4]byte) *gateway.Packet {
	t.Helper()
	select {
	case packetBytes := <-packets:
		packet, err := gateway.NewCodec(key).UnmarshalResponse(packetBytes)
		if err != nil {
			t.Fatalf("unmarshal gateway packet: %v", err)
		}
		return packet
	case <-time.After(time.Second):
		t.Fatalf("gateway packet was not sent")
		return nil
	}
}
