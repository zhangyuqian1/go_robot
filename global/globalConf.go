package global

import (
	"math/rand"
	"stressTest/jasmine"
	"sync"
	"time"
)

var (
	// 环境
	MetricsSvrChat string
	MetricsSvrData string
	ReportId       int
	ProjectId      uint64
	ZoneId         int

	//自定义
	CaseName     string
	WebSocketSvr = "ws://127.0.0.1:21000"
	UserName     string

	//Case
	PodRobotNum  int
	FirstAccount int
	LastAccount  int // 账号
	TimeOut      int64
	RampUpPeriod int

	//全局
	AccountPrefix string
	GAccountList  []string = nil
	PlayerNKList  []string = nil

	GlobalLock   sync.Mutex
	CaseMap      map[string]*jasmine.Task
	MixedScene   []*jasmine.Task
	RateLimiters map[string]*jasmine.Limiter

	LoginRate           *jasmine.Limiter
	MoveRate            *jasmine.Limiter
	BattleRate          *jasmine.Limiter
	PetRate             *jasmine.Limiter
	EquipRate           *jasmine.Limiter
	ArenaRate           *jasmine.Limiter
	DrawRate            *jasmine.Limiter
	FriendRate          *jasmine.Limiter
	PrayRate            *jasmine.Limiter
	HeartbeatInterval         = 10 * time.Second
	MoveStepPixels      int32 = 12
	MoveSpeedPixels     int32 = 300
	DragonCityBossIndex int   = 1
	DragonCityWaypoints [][2]int32
	BattleNormalSkillID int32 = 1101
	PetBattleIndex      int32 = 1
	PetCooldown               = 10 * time.Second
	ArenaListInterval         = 10 * time.Second
	ArenaFightDelay           = 3 * time.Second
	ArenaCooldown             = 10 * time.Second
	ArenaGMCommand            = ""
	ArenaChallengeIndex int32 = 1
	FriendAid           int32 = 16783890

	LoginThreshold  int64 = 1
	MoveThreshold   int64 = 5
	BattleThreshold int64 = 2
	PetThreshold    int64 = 1
	EquipThreshold  int64 = 1
	ArenaThreshold  int64 = 1
	DrawThreshold   int64 = 1
	FriendThreshold int64 = 1
	PrayThreshold   int64 = 1
	TotalCurrency   int64

	LoginRendezvous *jasmine.RendezvousPoint
	YokaConfig      *jasmine.YokaConfig
	CaseParam       jasmine.TestCase
	Jasmine         *jasmine.Perf

	PodCasterNum int
	PodCasters   []PodCaster
	PcmParams    string
	PdFstId      int
)

func MustGetRateLimiter(name string) *jasmine.Limiter {
	limiter, ok := RateLimiters[name]
	if !ok {
		panic("rate limiter not found: " + name)
	}
	return limiter
}

type PodCaster struct {
	State     int    //0未在房间内 1房间内待开局 2对局中
	TableName string //房间口令
	TableId   uint32 //房间id
	PtName    string
}

func AccountDispatcher() int {
	GlobalLock.Lock()
	var account int
	if FirstAccount > LastAccount {
		FirstAccount = int(YokaConfig.AccountIdx)
		account = FirstAccount
	} else {
		account = FirstAccount
		FirstAccount++
	}

	GlobalLock.Unlock()
	return account
}

func GenerateNickname(chars []string) string {
	rand.Seed(time.Now().UnixNano()) // 设置随机种子

	// 随机选择昵称长度，范围是 4 到 6
	length := rand.Intn(3) + 4

	// 构建昵称
	var nickname string
	for i := 0; i < length; i++ {
		nickname += chars[rand.Intn(len(chars))] // 随机选择一个汉字
	}

	return nickname
}
