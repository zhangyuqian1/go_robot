package startup

import (
	"testing"
	"time"

	"stressTest/global"
	"stressTest/jasmine"
	"stressTest/logger"

	"go.uber.org/zap/zapcore"
)

func init() {
	logger.Init("./logs/startup_test.log", zapcore.FatalLevel)
}

func TestApplyCustomCaseParamsUsesMrRobotDefaults(t *testing.T) {
	config := &jasmine.YokaConfig{PodId: 0, PodNum: 1}
	testCase := &jasmine.TestCase{}

	ApplyCustomCaseParams(config, testCase)

	if global.WebSocketSvr != "ws://127.0.0.1:21000" {
		t.Fatalf("WebSocketSvr = %q, want default local gateway", global.WebSocketSvr)
	}
	if global.HeartbeatInterval != 10*time.Second {
		t.Fatalf("HeartbeatInterval = %s, want 10s", global.HeartbeatInterval)
	}
	if global.LoginThreshold != 1 {
		t.Fatalf("LoginThreshold = %d, want 1", global.LoginThreshold)
	}
	if global.LoginRate == nil {
		t.Fatalf("LoginRate should be built from default --login-rate")
	}
	if global.MoveStepPixels != 12 || global.BattleNormalSkillID != 4209 || global.PetBattleIndex != 1 || global.ArenaChallengeIndex != 1 {
		t.Fatalf("post-login defaults = move-step:%d skill:%d pet-index:%d arena-index:%d", global.MoveStepPixels, global.BattleNormalSkillID, global.PetBattleIndex, global.ArenaChallengeIndex)
	}
	if global.DrawThreshold != 1 {
		t.Fatalf("DrawThreshold = %d, want 1", global.DrawThreshold)
	}
	if global.FriendThreshold != 1 || global.FriendAid != 16783880 {
		t.Fatalf("Friend config = rate:%d aid:%d", global.FriendThreshold, global.FriendAid)
	}
	if global.PrayThreshold != 1 {
		t.Fatalf("PrayThreshold = %d, want 1", global.PrayThreshold)
	}
	if global.MoveRate == nil || global.BattleRate == nil || global.PetRate == nil || global.EquipRate == nil || global.ArenaRate == nil || global.DrawRate == nil || global.FriendRate == nil || global.PrayRate == nil {
		t.Fatalf("post-login rate limiters should be built from defaults")
	}
}

func TestApplyCustomCaseParamsOverridesMrRobotValues(t *testing.T) {
	config := &jasmine.YokaConfig{PodId: 0, PodNum: 1}
	testCase := &jasmine.TestCase{CustomParams: []jasmine.CustomParam{
		{ParamName: "--ws", ParamValue: "ws://10.0.0.1:22000"},
		{ParamName: "--heartbeat-interval", ParamValue: "7"},
		{ParamName: "--login-rate", ParamValue: "9"},
		{ParamName: "--move-rate", ParamValue: "10"},
		{ParamName: "--move-step", ParamValue: "18"},
		{ParamName: "--battle-rate", ParamValue: "11"},
		{ParamName: "--battle-skill-id", ParamValue: "101"},
		{ParamName: "--pet-rate", ParamValue: "12"},
		{ParamName: "--pet-cooldown", ParamValue: "13"},
		{ParamName: "--equip-rate", ParamValue: "14"},
		{ParamName: "--arena-rate", ParamValue: "15"},
		{ParamName: "--arena-list-interval", ParamValue: "16"},
		{ParamName: "--arena-fight-delay", ParamValue: "17"},
		{ParamName: "--arena-cooldown", ParamValue: "18"},
		{ParamName: "--arena-gm", ParamValue: "open arena"},
		{ParamName: "--arena-index", ParamValue: "2"},
		{ParamName: "--draw-rate", ParamValue: "5"},
		{ParamName: "--friend-rate", ParamValue: "6"},
		{ParamName: "--friend-aid", ParamValue: "16888"},
		{ParamName: "--pray-rate", ParamValue: "7"},
	}}

	ApplyCustomCaseParams(config, testCase)

	if global.WebSocketSvr != "ws://10.0.0.1:22000" {
		t.Fatalf("WebSocketSvr = %q", global.WebSocketSvr)
	}
	if global.HeartbeatInterval != 7*time.Second || global.LoginThreshold != 9 {
		t.Fatalf("timing/rate config = heartbeat:%s login-rate:%d", global.HeartbeatInterval, global.LoginThreshold)
	}
	if global.LoginRate == nil {
		t.Fatalf("LoginRate should be built from explicit --login-rate")
	}
	if global.MoveThreshold != 10 || global.MoveStepPixels != 18 || global.BattleThreshold != 11 || global.BattleNormalSkillID != 101 {
		t.Fatalf("move/battle config = move-rate:%d move-step:%d battle-rate:%d skill:%d", global.MoveThreshold, global.MoveStepPixels, global.BattleThreshold, global.BattleNormalSkillID)
	}
	if global.PetThreshold != 12 || global.PetCooldown != 13*time.Second || global.EquipThreshold != 14 {
		t.Fatalf("pet/equip config = pet-rate:%d cooldown:%s equip-rate:%d", global.PetThreshold, global.PetCooldown, global.EquipThreshold)
	}
	if global.ArenaThreshold != 15 || global.ArenaListInterval != 16*time.Second || global.ArenaFightDelay != 17*time.Second || global.ArenaCooldown != 18*time.Second || global.ArenaGMCommand != "open arena" || global.ArenaChallengeIndex != 2 {
		t.Fatalf("arena config = rate:%d list:%s delay:%s cooldown:%s gm:%q index:%d", global.ArenaThreshold, global.ArenaListInterval, global.ArenaFightDelay, global.ArenaCooldown, global.ArenaGMCommand, global.ArenaChallengeIndex)
	}
	if global.DrawThreshold != 5 {
		t.Fatalf("DrawThreshold = %d, want 5", global.DrawThreshold)
	}
	if global.FriendThreshold != 6 || global.FriendAid != 16888 {
		t.Fatalf("Friend config = rate:%d aid:%d", global.FriendThreshold, global.FriendAid)
	}
	if global.PrayThreshold != 7 {
		t.Fatalf("PrayThreshold = %d, want 7", global.PrayThreshold)
	}
	if global.MoveRate == nil || global.BattleRate == nil || global.PetRate == nil || global.EquipRate == nil || global.ArenaRate == nil || global.DrawRate == nil || global.FriendRate == nil || global.PrayRate == nil {
		t.Fatalf("post-login rate limiters should be built from explicit rates")
	}
}

func TestApplyCustomCaseParamsCanDisableLoginLimiterWithoutPanic(t *testing.T) {
	config := &jasmine.YokaConfig{PodId: 0, PodNum: 1}
	testCase := &jasmine.TestCase{CustomParams: []jasmine.CustomParam{
		{ParamName: "--login-rate", ParamValue: "0"},
		{ParamName: "--move-rate", ParamValue: "0"},
		{ParamName: "--battle-rate", ParamValue: "0"},
		{ParamName: "--pet-rate", ParamValue: "0"},
		{ParamName: "--equip-rate", ParamValue: "0"},
		{ParamName: "--arena-rate", ParamValue: "0"},
		{ParamName: "--draw-rate", ParamValue: "0"},
		{ParamName: "--friend-rate", ParamValue: "0"},
		{ParamName: "--pray-rate", ParamValue: "0"},
	}}

	ApplyCustomCaseParams(config, testCase)

	if global.LoginThreshold != 0 {
		t.Fatalf("LoginThreshold = %d, want explicit disable", global.LoginThreshold)
	}
	if global.LoginRate != nil {
		t.Fatalf("LoginRate = %#v, want nil when disabled", global.LoginRate)
	}
	if global.DrawThreshold != 0 {
		t.Fatalf("DrawThreshold = %d, want explicit disable", global.DrawThreshold)
	}
	if global.DrawRate != nil {
		t.Fatalf("DrawRate = %#v, want nil when disabled", global.DrawRate)
	}
	if global.PrayThreshold != 0 {
		t.Fatalf("PrayThreshold = %d, want explicit disable", global.PrayThreshold)
	}
	if global.PrayRate != nil {
		t.Fatalf("PrayRate = %#v, want nil when disabled", global.PrayRate)
	}
	if global.MoveRate != nil || global.BattleRate != nil || global.PetRate != nil || global.EquipRate != nil || global.ArenaRate != nil || global.DrawRate != nil || global.FriendRate != nil || global.PrayRate != nil {
		t.Fatalf("post-login rate limiters should be nil when disabled")
	}
}
