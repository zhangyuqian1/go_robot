package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"log"
	"strconv"

	"stressTest/global"
	jasmine "stressTest/jasmine"
	"stressTest/logger"
	"stressTest/startup"
	"stressTest/tools"

	"go.uber.org/zap/zapcore"

	yokaman "gitlab.qa.dobest.com/qa/yokamancli"
)

func init() {
	logger.Init("./logs/robot.log", zapcore.DebugLevel) //如果正式跑，可以把日志级别调高
	global.ProjectId, _ = strconv.ParseUint(tools.GetEnvVarOrDefault("PROJECT_ID", "7320"), 10, 64)
	global.ReportId, _ = strconv.Atoi(tools.GetEnvVarOrDefault("REPORT_ID", "0"))
	global.ZoneId, _ = strconv.Atoi(tools.GetEnvVarOrDefault("ZONE_ID", "0"))
}

func GetCommandlineArgs() {
	for global.CaseName, _ = range global.YokaConfig.TestCases {
		testCase := global.YokaConfig.TestCases[global.CaseName]
		startup.ApplyCustomCaseParams(global.YokaConfig, testCase)
		global.PodRobotNum = int(testCase.PodRobotNum)
		global.FirstAccount = ((int(global.YokaConfig.TotalRobotNum) / global.YokaConfig.PodNum) * global.YokaConfig.PodId) + int(global.YokaConfig.AccountIdx)
		global.LastAccount = global.FirstAccount + global.PodRobotNum - 1
		global.TimeOut = testCase.TimeOut
		global.RampUpPeriod = testCase.RampUpPeriod
		if global.YokaConfig.AccountOn {
			global.GAccountList = tools.GetAccountToFile(global.YokaConfig.AccountFileName)
		} else {
			global.AccountPrefix = global.YokaConfig.AccountPrefix
		}
		if global.YokaConfig.LogOn {
			logger.Init("./logs/robot.log", zapcore.Level(global.YokaConfig.LogLevel))
		}
	}
	printConfig()
}

// printConfig 打印关键配置参数
func printConfig() {
	logger.Info("========== 配置参数信息 ==========")
	logger.Info(fmt.Sprintf("CaseName: %s", global.CaseName))
	logger.Info(fmt.Sprintf("PodNum: %d", global.YokaConfig.PodNum))
	logger.Info(fmt.Sprintf("PodId: %d", global.YokaConfig.PodId))
	logger.Info(fmt.Sprintf("TotalRobotNum: %d", global.YokaConfig.TotalRobotNum))
	logger.Info(fmt.Sprintf("PodRobotNum: %d", global.PodRobotNum))
	logger.Info(fmt.Sprintf("FirstAccount: %d", global.FirstAccount))
	logger.Info(fmt.Sprintf("LastAccount: %d", global.LastAccount))
	logger.Info(fmt.Sprintf("AccountPrefix: %s", global.AccountPrefix))
	logger.Info(fmt.Sprintf("AccountOn: %v", global.YokaConfig.AccountOn))
	logger.Info(fmt.Sprintf("LogOn: %v", global.YokaConfig.LogOn))
	logger.Info(fmt.Sprintf("LogLevel: %d", global.YokaConfig.LogLevel))
	logger.Info(fmt.Sprintf("TimeOut: %d", global.TimeOut))
	logger.Info(fmt.Sprintf("RampUpPeriod: %d", global.RampUpPeriod))
	logger.Info(fmt.Sprintf("Duration: %d", global.YokaConfig.TestCases[global.CaseName].Duration))
	logger.Info(fmt.Sprintf("WebSocketSvr: %s", global.WebSocketSvr))
	logger.Info(fmt.Sprintf("LoginThreshold: %d", global.LoginThreshold))
	logger.Info(fmt.Sprintf("HeartbeatInterval: %s", global.HeartbeatInterval))
	logger.Info(fmt.Sprintf("Move: rate=%d step=%d", global.MoveThreshold, global.MoveStepPixels))
	logger.Info(fmt.Sprintf("Battle: rate=%d normal_skill_id=%d", global.BattleThreshold, global.BattleNormalSkillID))
	logger.Info(fmt.Sprintf("Pet: rate=%d index=%d cooldown=%s", global.PetThreshold, global.PetBattleIndex, global.PetCooldown))
	logger.Info(fmt.Sprintf("Equip: rate=%d", global.EquipThreshold))
	logger.Info(fmt.Sprintf("Arena: rate=%d list_interval=%s fight_delay=%s cooldown=%s index=%d gm=%q", global.ArenaThreshold, global.ArenaListInterval, global.ArenaFightDelay, global.ArenaCooldown, global.ArenaChallengeIndex, global.ArenaGMCommand))
	logger.Info(fmt.Sprintf("Draw: rate=%d", global.DrawThreshold))
	logger.Info(fmt.Sprintf("Pray: rate=%d", global.PrayThreshold))
	logger.Info(fmt.Sprintf("Friend: rate=%d aid=%d", global.FriendThreshold, global.FriendAid))
	logger.Info(fmt.Sprintf("ProjectId: %d", global.ProjectId))
	logger.Info(fmt.Sprintf("ReportId: %d", global.ReportId))
	logger.Info(fmt.Sprintf("UserName: %s", global.UserName))
	logger.Info(fmt.Sprintf("LoginRateLimiter Enabled: %v", global.LoginRate != nil))
	logger.Info("======================================")
}
func YokaMetric() (yokacli *yokaman.YoKaMan) {
	yokacli = yokaman.YoKaManCli()

	if global.ReportId == 0 {
		yokacli.SetProjectInfo(uint(global.ProjectId), global.UserName)
		yokacli.SetCaseName(global.CaseName)
		yokacli.SetPodInfo(uint(global.PodRobotNum), int8(global.YokaConfig.PodId)) //设置该节点的人数

	}

	yokacli.SetTestInfo(int32(global.ReportId), uint(global.YokaConfig.PodId))

	yokacli.SetEnvOptions(yokaman.SvrAddrOptions{
		// 配置服务器地址和环境信息，谨慎切换测试环境和生产环境
		Env: yokaman.Prod,
		//Region: yokaman.Default,
	})
	err := yokacli.Start() //启动数据上报客户端，在后台会启动线程上传
	if err != nil {
		logger.Error(fmt.Sprintf("Start metrics client failed err %s", err))
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		logger.Debug("系统已退出")
		yokacli.Stop()
		os.Exit(0)
	}()

	return yokacli
}

func main() {
	global.YokaConfig = jasmine.NewYokaConfig()
	global.YokaConfig.CommandParse()

	GetCommandlineArgs()

	tools.MkdirPath()

	CaseInit()

	// 检查 CaseName 是否有效
	if global.CaseName == "" {
		logger.Error("CaseName is empty, please provide --case parameter")
		os.Exit(1)
	}

	// 检查 Case 是否存在
	if _, exists := global.CaseMap[global.CaseName]; !exists {
		logger.Error(fmt.Sprintf("Case '%s' not found in CaseMap", global.CaseName))
		logger.Error(fmt.Sprintf("Available cases: %v", getAvailableCases()))
		os.Exit(1)
	}

	_ = YokaMetric()

	global.LoginRendezvous = jasmine.NewRendezvousPoint("LoginRendezvous", int64(global.PodRobotNum))

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	global.Jasmine = jasmine.NewStandaloneBoomer(global.PodRobotNum, 0, nil)

	logger.Debug("Start stresstest")

	if global.CaseName == "混合场景" {
		global.Jasmine.Run(jasmine.TimerMode, strconv.Itoa(global.YokaConfig.TestCases[global.CaseName].Duration)+"s", global.MixedScene...)
	} else {
		global.Jasmine.Run(jasmine.TimerMode, strconv.Itoa(global.YokaConfig.TestCases[global.CaseName].Duration)+"s", global.CaseMap[global.CaseName])
	}
}

func getAvailableCases() []string {
	cases := make([]string, 0, len(global.CaseMap))
	for k := range global.CaseMap {
		cases = append(cases, k)
	}
	return cases
}
