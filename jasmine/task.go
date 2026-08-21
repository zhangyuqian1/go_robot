package jasmine

import "context"

type Task struct {
	Weight int
	Fn     func(ctx context.Context)
	Name   string
}

// YokaConfig 命令行参数解析
type YokaConfig struct {
	TotalRobotNum   uint64
	PodNum          int
	PodId           int
	AccountOn       bool   // 账号开关
	AccountFileName string // 账号文件
	AccountPrefix   string // 账号前置
	AccountIdx      uint64 // 账号Id
	LogOn           bool   // 日志开关
	LogLevel        int    //日志等级
	Consecutively   int    // 是否顺序执行(仅对jemeter有效)
	TestCases       map[string]*TestCase
	GlobalParams    []CustomParam
}

type TestCase struct {
	CaseRobotNum uint64 // 用例总用户数
	PodRobotNum  uint64 // pod的用户数
	RampUpPeriod int    // 协程初始化时长
	Duration     int    // 测试时长
	LoopNum      int    // 执行轮次
	TimeOut      int64  //超时时间
	CustomParams []CustomParam
}

type CustomParam struct {
	ParamName  string
	ParamValue string
}

type AccountConf struct {
	AccountId string
	PassId    string
}
