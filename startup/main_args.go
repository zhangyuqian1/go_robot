package startup

import (
	"strconv"
	"time"

	"stressTest/global"
	"stressTest/jasmine"
)

const (
	defaultWebSocketSvr        = "ws://127.0.0.1:21000"
	defaultHeartbeatInterval   = 10 * time.Second
	defaultLoginRate           = 1
	defaultMoveRate            = 5
	defaultMoveStepPixels      = 12
	defaultBattleRate          = 2
	defaultBattleNormalSkillID = 4209
	defaultPetRate             = 1
	defaultPetBattleIndex      = 1
	defaultPetCooldown         = 10 * time.Second
	defaultEquipRate           = 1
	defaultArenaRate           = 1
	defaultArenaListInterval   = 10 * time.Second
	defaultArenaFightDelay     = 3 * time.Second
	defaultArenaCooldown       = 10 * time.Second
	defaultArenaGMCommand      = ""
	defaultArenaChallengeIndex = 1
	defaultDrawRate           = 1
	defaultFriendRate         = 1
	defaultFriendAid    int32 = 16783880
	defaultPrayRate           = 1
)

func ApplyCustomCaseParams(config *jasmine.YokaConfig, testCase *jasmine.TestCase) {
	resetMrRobotDefaults()
	if config == nil || testCase == nil {
		buildLoginRateLimiter(config)
		return
	}

	for _, custom := range testCase.CustomParams {
		switch custom.ParamName {
		case "--creator":
			global.UserName = custom.ParamValue
		case "--ws":
			global.WebSocketSvr = custom.ParamValue
		case "--heartbeat-interval":
			if seconds, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil && seconds > 0 {
				global.HeartbeatInterval = time.Duration(seconds) * time.Second
			}
		case "--login-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.LoginThreshold = value
			}
		case "--move-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.MoveThreshold = value
			}
		case "--move-step":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 32); err == nil && value > 0 {
				global.MoveStepPixels = int32(value)
			}
		case "--battle-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.BattleThreshold = value
			}
		case "--battle-skill-id":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 32); err == nil {
				global.BattleNormalSkillID = int32(value)
			}
		case "--pet-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.PetThreshold = value
			}
		case "--pet-index":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 32); err == nil && value > 0 {
				global.PetBattleIndex = int32(value)
			}
		case "--pet-cooldown":
			if seconds, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil && seconds >= 0 {
				global.PetCooldown = time.Duration(seconds) * time.Second
			}
		case "--equip-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.EquipThreshold = value
			}
		case "--arena-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.ArenaThreshold = value
			}
		case "--arena-list-interval":

			if seconds, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil && seconds > 0 {
				global.ArenaListInterval = time.Duration(seconds) * time.Second
			}
		case "--arena-fight-delay":
			if seconds, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil && seconds >= 0 {
				global.ArenaFightDelay = time.Duration(seconds) * time.Second
			}
		case "--arena-cooldown":
			if seconds, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil && seconds >= 0 {
				global.ArenaCooldown = time.Duration(seconds) * time.Second
			}
		case "--arena-gm":
			global.ArenaGMCommand = custom.ParamValue
		case "--arena-index":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 32); err == nil && value > 0 {
				global.ArenaChallengeIndex = int32(value)
			}
		case "--draw-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.DrawThreshold = value
			}
		case "--friend-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.FriendThreshold = value
			}
		case "--friend-aid":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 32); err == nil {
				global.FriendAid = int32(value)
			}
		case "--pray-rate":
			if value, err := strconv.ParseInt(custom.ParamValue, 10, 64); err == nil {
				global.PrayThreshold = value
			}
		case "--project-id":
			if value, err := strconv.ParseUint(custom.ParamValue, 10, 64); err == nil {
				global.ProjectId = value
			}
		}
	}
	buildLoginRateLimiter(config)
}

func resetMrRobotDefaults() {
	global.WebSocketSvr = defaultWebSocketSvr
	global.HeartbeatInterval = defaultHeartbeatInterval
	global.LoginThreshold = defaultLoginRate
	global.MoveThreshold = defaultMoveRate
	global.MoveStepPixels = defaultMoveStepPixels
	global.BattleThreshold = defaultBattleRate
	global.BattleNormalSkillID = defaultBattleNormalSkillID
	global.PetThreshold = defaultPetRate
	global.PetBattleIndex = defaultPetBattleIndex
	global.PetCooldown = defaultPetCooldown
	global.EquipThreshold = defaultEquipRate
	global.ArenaThreshold = defaultArenaRate
	global.ArenaListInterval = defaultArenaListInterval
	global.ArenaFightDelay = defaultArenaFightDelay
	global.ArenaCooldown = defaultArenaCooldown
	global.ArenaGMCommand = defaultArenaGMCommand
	global.ArenaChallengeIndex = defaultArenaChallengeIndex
	global.FriendAid = defaultFriendAid
	global.DrawThreshold = defaultDrawRate
	global.FriendThreshold = defaultFriendRate
	global.PrayThreshold = defaultPrayRate
	global.LoginRate = nil
	global.MoveRate = nil
	global.BattleRate = nil
	global.PetRate = nil
	global.EquipRate = nil
	global.ArenaRate = nil
	global.DrawRate = nil
	global.FriendRate = nil
	global.PrayRate = nil
	global.RateLimiters = nil
}

func buildLoginRateLimiter(config *jasmine.YokaConfig) {
	global.LoginRate = buildRateLimiter(config, global.LoginThreshold)
	global.MoveRate = buildRateLimiter(config, global.MoveThreshold)
	global.BattleRate = buildRateLimiter(config, global.BattleThreshold)
	global.PetRate = buildRateLimiter(config, global.PetThreshold)
	global.EquipRate = buildRateLimiter(config, global.EquipThreshold)
	global.ArenaRate = buildRateLimiter(config, global.ArenaThreshold)
	global.DrawRate = buildRateLimiter(config, global.DrawThreshold)
	global.FriendRate = buildRateLimiter(config, global.FriendThreshold)
	global.PrayRate = buildRateLimiter(config, global.PrayThreshold)
	global.RateLimiters = map[string]*jasmine.Limiter{
		"login":  global.LoginRate,
		"move":   global.MoveRate,
		"battle": global.BattleRate,
		"pet":    global.PetRate,
		"equip":  global.EquipRate,
		"arena":  global.ArenaRate,
		"draw":   global.DrawRate,
		"friend": global.FriendRate,
		"pray":   global.PrayRate,
	}

}

func buildRateLimiter(config *jasmine.YokaConfig, threshold int64) *jasmine.Limiter {
	if config == nil || threshold <= 0 {
		return nil
	}
	limiter, _ := jasmine.CreateRateLimiter(config.PodId, config.PodNum, threshold)
	return limiter
}
