package jasmine

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"stressTest/logger"
	"time"
)

func NewYokaConfig() (yokaConfig *YokaConfig) {
	yokaConfig = &YokaConfig{
		TestCases: make(map[string]*TestCase),
	}
	return yokaConfig
}

func NewYokaTestCase() (yokaTestCase *TestCase) {
	yokaTestCase = &TestCase{}
	return yokaTestCase
}

func (yokaConfig *YokaConfig) CommandParse() {
	yokaTestCase := NewYokaTestCase()
	args := os.Args[1:]
	if len(args) == 0 {
		logger.Error("No arguments provided.")
		return
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--total-robot-num":
			i++
			if i < len(args) {
				yokaConfig.TotalRobotNum, _ = strconv.ParseUint(args[i], 10, 64)
			}
		case "--pod-num":
			i++
			if i < len(args) {
				yokaConfig.PodNum, _ = strconv.Atoi(args[i])
			}
		case "--pod-id":
			i++
			if i < len(args) {
				yokaConfig.PodId, _ = strconv.Atoi(args[i])
			}
		case "--account-on":
			i++
			if i < len(args) {
				yokaConfig.AccountOn, _ = strconv.ParseBool(args[i])
			}
		case "--account-filename":
			i++
			if i < len(args) {
				yokaConfig.AccountFileName = args[i]
			}
		case "--account-prefix":
			i++
			if i < len(args) {
				yokaConfig.AccountPrefix = args[i]
			}
		case "--account-idx":
			i++
			if i < len(args) {
				yokaConfig.AccountIdx, _ = strconv.ParseUint(args[i], 10, 64)
			}
		case "--log-on":
			i++
			if i < len(args) {
				yokaConfig.LogOn, _ = strconv.ParseBool(args[i])
			}
		case "--log-level":
			i++
			if i < len(args) {
				yokaConfig.LogLevel, _ = strconv.Atoi(args[i])
			}
		case "--consecutively":
			i++
			if i < len(args) {
				yokaConfig.Consecutively, _ = strconv.Atoi(args[i])
			}
		case "--case":
			i++
			if i < len(args) {
				caseName := args[i]
				yokaConfig.TestCases[caseName] = yokaTestCase
			}
		case "--case-robot-num":
			i++
			if i < len(args) {
				yokaTestCase.CaseRobotNum, _ = strconv.ParseUint(args[i], 10, 64)
			}
		case "--pod-robot-num":
			i++
			if i < len(args) {
				yokaTestCase.PodRobotNum, _ = strconv.ParseUint(args[i], 10, 64)
			}
		case "--ramp-up-period":
			i++
			if i < len(args) {
				yokaTestCase.RampUpPeriod, _ = strconv.Atoi(args[i])
			}
		case "--duration":
			i++
			if i < len(args) {
				yokaTestCase.Duration, _ = strconv.Atoi(args[i])
			}
		case "--loop-num":
			i++
			if i < len(args) {
				yokaTestCase.LoopNum, _ = strconv.Atoi(args[i])
			}
		case "--timeout":
			i++
			if i < len(args) {
				yokaTestCase.TimeOut, _ = strconv.ParseInt(args[i], 10, 64)
			}
		default:
			customParam := CustomParam{
				ParamName: args[i],
			}
			i++
			if i < len(args) {
				customParam.ParamValue = args[i]
				if yokaTestCase != nil {
					yokaTestCase.CustomParams = append(yokaTestCase.CustomParams, customParam)
				} else {
					yokaConfig.GlobalParams = append(yokaConfig.GlobalParams, customParam)
				}
			}
		}
	}
	logger.Info("The command line parameters are parsed successfully!")
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func GetPodCount(podId int, podNum int, threshold int64) int64 {
	base := threshold / int64(podNum)
	remainder := threshold % int64(podNum)
	for pod := 0; pod <= podNum; pod++ {
		// 每个 Pod 初始分配的数量是 base
		podThreshold := base
		// 如果当前 Pod 应该获得一个额外的余数
		if int64(pod) < remainder {
			podThreshold++
		}
		if pod == podId {
			logger.Debug(fmt.Sprintf("Pod: %d, gets %d", pod, podThreshold))
			return podThreshold
		}
	}
	return 0
}

func CreateRateLimiter(podId int, podNum int, threshold int64) (rateLimiter *Limiter, err error) {
	if threshold < int64(podNum) {
		rateLimiter, _ = NewLimiter(time.Second, 1, func() (Window, StopFunc) {
			return NewLocalWindow()
		})
		return rateLimiter, nil
	} else {
		podThreshold := GetPodCount(podId, podNum, threshold)
		logger.Debug(fmt.Sprintf("podThreshold: %d", podThreshold))
		rateLimiter, _ = NewLimiter(time.Second, podThreshold, func() (Window, StopFunc) {
			return NewLocalWindow()
		})
		return rateLimiter, nil
	}
}
