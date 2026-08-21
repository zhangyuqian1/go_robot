package module

import (
	"fmt"
	"net/http"
	"stressTest/gameproto/arena"
	"stressTest/gameproto/common"
	"stressTest/gateway"
	"stressTest/global"
	"stressTest/logger"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	cmap "github.com/orcaman/concurrent-map/v2"
	yokaman "gitlab.qa.dobest.com/qa/yokamancli"
)

const (
	STEP_SUCC = iota
	STEP_FAIL
	STEP_TIMEOUT
)

type webSocketConn struct {
	Conn *websocket.Conn
}

type GameClient struct {
	WebSocketSvr  string
	WebSocketConn *webSocketConn
	PingSending   bool

	WriteMutex sync.Mutex

	ClientSalt   uint32
	ServerSalt   uint32
	GatewayKey   gateway.Key
	GatewayCodec *gateway.Codec

	Token string

	OpenTLS    bool
	HttpClient *http.Client

	FriendIDList []uint64

	AccountID             int32
	ActorID               int32
	ActorName             string
	Handle                int64
	RawServerID           int32
	CurrentServerID       int32
	MapData               map[int32]int64
	SceneID               int32
	FubenID               int32
	SceneType             int32
	PX                    float32
	PY                    float32
	LoginDone             bool
	CreatePending         bool
	LastLoginCode         int32
	LastCreateActorResult int32
	LastActorLoginResult  int32
	LastHeartbeatTick     int32
	LoginStarted          bool

	Entities            map[int64]*EntityState
	Monsters            map[int64]*MonsterState
	MoveAngle           int32
	MoveBlocked         bool
	LastMoveTick        int32
	BattleTargetHandle  int64
	BattleBosses        map[int32]*BossState
	BattleSkillRandVal  int32
	SelfDead            bool
	Pets                map[int32]*PetState
	SelectedPetSeriesID int32
	PetBattlePending    bool
	PetCooldownUntil    time.Time
	PendingEquipItems   []*common.UserItem
	GiftSeries          *common.Series
	WeaponSeries        *common.Series
	ArenaEnemies        []*arena.ArenaEnemyInfo
	ArenaFightPending   bool
	ArenaFightReadyAt   time.Time
	ArenaInFight        bool
	ArenaScore          int32
	ArenaRank           int32
	ArenaLastWin        int32
	ArenaCooldownUntil  time.Time
	Skills              map[int32]*SkillState

	IsWebConnected bool
}

type Robot struct {
	Index    int // 序号
	ID       int
	PTName   string // 账号
	PassId   string // 通行证数字ID
	Password string // 密码

	Client *GameClient // 与GameServer的连接

	Mark cmap.ConcurrentMap[string, *Events]

	ConnStopSignal   chan struct{}
	TickStopSignal   chan struct{} //用于通知定时程序安全退出
	AsyncMonitSignal chan struct{} //用于通知定时程序安全退出
	TickStopOnce     sync.Once

	Ticking       bool
	AsyncMoniting bool

	AsyncMonitSyncGroup sync.WaitGroup
	TickSyncGroup       sync.WaitGroup
	RecvSyncGroup       sync.WaitGroup
}

//type RecordEvents struct {
//	Lf       *lockfree.Lockfree[*Events]
//	Handler  lockfree.EventHandler[*Events]
//	producer *lockfree.Producer[*Events]
//}
//
//type eventHandler[T Events] struct {
//	signal chan struct{}
//	now    time.Time
//	events *Events
//}

type Events struct {
	StartTime    int64
	EndTime      int64
	TimeOut      int64
	PacketLength int64
	Code         uint16
	stopChan     chan bool
	ErrCode      any
	permits      int32 //用原子信号量代表request请求是否已经被response响应
	stopOnce     sync.Once
}

func (e *Events) finish() {
	if e.stopChan != nil {
		e.stopOnce.Do(func() {
			close(e.stopChan)
		})
	}
}

func (r *Robot) StartRobotConn() {
	r.RecvSyncGroup.Add(1)
	go func() {
		defer r.RecvSyncGroup.Done()
		r.AesRobotConnRecv()
		logger.Info(fmt.Sprintf("Robot %s RobotConnRecv goroutine exited", r.PTName))
	}()
}

func (r *Robot) StopRobotConn() {
	if r.ConnStopSignal != nil {
		r.CloseSocket()
		close(r.ConnStopSignal) // 关闭ConnStopSignal通道，通知接收协程退出
		r.RecvSyncGroup.Wait()  // 等待接收协程退出
		logger.Info(fmt.Sprintf("Robot %s StopRobotConn", r.PTName))
	} else {
		logger.Error(fmt.Sprintf("Robot %s ConnStopSignal is nil, can't stop RobotConnRecv", r.PTName))
	}
	r.ConnStopSignal = nil
}

func (r *Robot) StartAsyncMonit() {
	r.AsyncMonitSyncGroup.Add(1) // 增加等待计数
	go func() {
		defer r.AsyncMonitSyncGroup.Done() // 协程退出时减少计数
		r.AsyncMonit()
		logger.Info(fmt.Sprintf("Robot %s AsyncWait goroutine exited", r.PTName))
	}()
}

func (r *Robot) CloseAsyncMonit() {
	if r.AsyncMonitSignal != nil {
		close(r.AsyncMonitSignal)    // 关闭AsyncWaitSignal通道，通知异步等待协程退出
		r.AsyncMonitSyncGroup.Wait() // 等待异步等待协程退出
		logger.Info(fmt.Sprintf("Robot %s CloseAsyncMonit", r.PTName))
	} else {
		logger.Error(fmt.Sprintf("Robot %s AsyncMonitSignal is nil, can't stop AsyncMonit", r.PTName))
	}
	r.AsyncMonitSignal = nil
}

func (r *Robot) AsyncFire(requestName string, packetLength int64, code uint16, errCode any) {
	endTime := time.Now().UnixMilli()
	mark, ok := r.Mark.Get(requestName)
	if !ok {
		//logger.Error(fmt.Sprintf("account: %s Not Found proto:%s", r.PTName, requestName))
		return
	}
	mark.EndTime = endTime
	mark.PacketLength = packetLength
	mark.Code = code
	mark.ErrCode = errCode
	if endTime-mark.StartTime >= mark.TimeOut*1000 {
		mark.Code = STEP_TIMEOUT
	}
	r.Mark.Set(requestName, mark)
	r.RecordSuccess(requestName, 0, endTime, mark.Code)
}

func (r *Robot) AsyncBegin(requestName string, timeOut int64, isWait bool) {
	// 生产者
	startTime := time.Now().UnixMilli()
	mark := &Events{
		StartTime: startTime,
		TimeOut:   timeOut,
	}
	if isWait {
		mark.stopChan = make(chan bool, 0)
	} else {
		mark.stopChan = nil
	}
	r.Mark.Set(requestName, mark)
}

func (r *Robot) AsyncWait(requestName string) {
	logger.Debug(fmt.Sprintf("account:%s, request_name: %s", r.PTName, requestName))
	mark, exists := r.Mark.Get(requestName)
	if !exists {
		return
	}
	req := yokaman.Request{
		Name:    requestName,
		ReqTime: mark.StartTime,
	}
	err := yokaman.YoKaManCli().StatRequest(req)
	if err != nil {
		logger.Error(fmt.Sprintf("StatRequest error: %s", err))
	}
	// 等待消费者消费结束后，才可以继续执行
	if mark.stopChan != nil {
		<-mark.stopChan
		if exists := r.Mark.Has(requestName); exists {
			r.Mark.Remove(requestName)
		}
	}
}

func (r *Robot) RecordSuccess(requestName string, responseLength int64, endTime int64, code uint16) {
	mark, ok := r.Mark.Get(requestName)
	if !ok {
		//logger.Error(fmt.Sprintf("Request %s not in Mark.", requestName))
		return
	}

	if endTime < mark.StartTime {
		logger.Error(fmt.Sprintf("Request %s startTime > endTime", requestName))
		mark.finish()
		return
	}
	req := yokaman.Response{
		Name:     requestName,
		ReqTime:  mark.StartTime,
		RespTime: endTime,
		Code:     code,
		RobotId:  uint(0),
	}
	if req.Code == STEP_FAIL {
		req.ErrCode = mark.ErrCode
	}

	if req.Code == STEP_TIMEOUT {
		logger.Error(fmt.Sprintf("TIMEOUT metrics:%s, %s, reqTime:%d, respTime:%d", req.Name, r.PTName, req.ReqTime, req.RespTime))
	}

	//logger.Error(fmt.Sprintf("yokaman.Response: %s, %d", req.Name, req.RespTime-req.ReqTime))
	err := yokaman.YoKaManCli().StatResponse(req)
	if err != nil {
		logger.Error(fmt.Sprintf("ReqMetrics error: %s", err))
	}
	if mark.stopChan != nil {
		mark.finish()
	}
	r.Mark.Remove(requestName)
}

func (r *Robot) stopPendingGatewayWaits(code uint16, errCode any) {
	if r.Mark.IsEmpty() {
		return
	}
	endTime := time.Now().UnixMilli()
	for item := range r.Mark.IterBuffered() {
		mark := item.Val
		mark.EndTime = endTime
		mark.Code = code
		mark.ErrCode = errCode
		mark.finish()
		r.Mark.Remove(item.Key)
	}
}

func (r *Robot) AsyncMonit() {
	sleepDuration := time.Duration(global.TimeOut) * time.Second
	if sleepDuration <= 0 {
		sleepDuration = time.Second
	}
	for {
		select {
		case <-r.AsyncMonitSignal:
			logger.Info(fmt.Sprintf("Robot %s Async Wait Recv Stop Signal", r.PTName))
			return
		default:
			tupleChan := r.Mark.IterBuffered() // 获取已经聚合的通道
			for item := range tupleChan {
				// 处理每个 item
				val := item.Val
				nowTime := time.Now().UnixMilli()
				if nowTime-val.StartTime >= val.TimeOut*1000 {
					r.RecordSuccess(item.Key, 0, nowTime, STEP_TIMEOUT)
				}
			}
			time.Sleep(sleepDuration) // 休眠协议的默认超时时间即可
		}
	}
}

func (r *Robot) CloseSocket() {
	if r.Client == nil {
		return
	}
	r.Client.PingSending = false
	if r.Ticking {
		r.signalHeartbeatStop()
	}
	if r.Client.IsWebConnected && r.Client.WebSocketConn != nil && r.Client.WebSocketConn.Conn != nil {
		r.Client.IsWebConnected = false
		err := r.Client.WebSocketConn.Conn.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("WebSocket close err: %s", err))
		}
	}
	r.stopPendingGatewayWaits(STEP_FAIL, "websocket closed")
}

func (r *Robot) Close() {
	r.CloseSocket()
	r.StopRobotConn()
	r.StopHeartBeat()
	r.CloseAsyncMonit()
}

type HttpResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GmResponse struct {
	Code  int `json:"code"`
	Items struct {
		Items []Item `json:"items"`
	} `json:"items"`
	Msg string `json:"msg"`
}

type Item struct {
	ItemID int `json:"itemID"`
	Num    int `json:"num"`
}
