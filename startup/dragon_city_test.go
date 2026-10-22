package startup

import (
	"reflect"
	"testing"
	"time"

	"stressTest/global"
	"stressTest/jasmine"
)

func TestMoveSpeedConfig(t *testing.T) {
	t.Cleanup(resetMrRobotDefaults)
	config := &jasmine.YokaConfig{PodId: 0, PodNum: 1}
	for _, test := range []struct {
		value string
		want  int32
	}{
		{"200", 200}, {"600", 600}, {"1", 1}, {"0", 300}, {"-1", 300}, {"invalid", 300}, {"1.5", 300}, {"2147483648", 300},
	} {
		ApplyCustomCaseParams(config, &jasmine.TestCase{CustomParams: []jasmine.CustomParam{
			{ParamName: "--move-speed", ParamValue: test.value},
		}})
		if global.MoveSpeedPixels != test.want {
			t.Fatalf("move-speed=%q got=%d want=%d", test.value, global.MoveSpeedPixels, test.want)
		}
	}
	global.MoveSpeedPixels = 999
	ApplyCustomCaseParams(config, &jasmine.TestCase{})
	if global.MoveSpeedPixels != 300 {
		t.Fatal("missing move-speed must reset to the default")
	}
}

func TestParseDragonCityWaypoints(t *testing.T) {
	for _, test := range []struct {
		value string
		want  [][2]int32
	}{
		{"69,32", [][2]int32{{69, 32}}},
		{"72,32;69,32", [][2]int32{{72, 32}, {69, 32}}},
		{"72,32;69,32;39,32", [][2]int32{{72, 32}, {69, 32}, {39, 32}}},
		{" 72,32 ; 69,32;39,32;39,56 ", [][2]int32{{72, 32}, {69, 32}, {39, 32}, {39, 56}}},
		{"0,21474836", [][2]int32{{0, 21474836}}},
	} {
		got, err := parseDragonCityWaypoints(test.value)
		if err != nil || !reflect.DeepEqual(got, test.want) {
			t.Fatalf("parse %q = %v, %v; want %v", test.value, got, err, test.want)
		}
	}
	for _, value := range []string{"", " ", "69", "69,32,1", "69,32;", "-1,32", "1.5,32", "x,32", "21474837,0", "2147483648,0", "1,1;2,2;3,3;4,4;5,5"} {
		if points, err := parseDragonCityWaypoints(value); err == nil || points != nil {
			t.Fatalf("invalid route %q accepted: %v", value, points)
		}
	}
}

func TestDragonCityWaypointsConfig(t *testing.T) {
	t.Cleanup(resetMrRobotDefaults)
	config := &jasmine.YokaConfig{PodId: 0, PodNum: 1}
	ApplyCustomCaseParams(config, &jasmine.TestCase{CustomParams: []jasmine.CustomParam{
		{ParamName: "--dragon-city-waypoints", ParamValue: "72,32;69,32;39,32;39,56"},
	}})
	if want := [][2]int32{{72, 32}, {69, 32}, {39, 32}, {39, 56}}; !reflect.DeepEqual(global.DragonCityWaypoints, want) {
		t.Fatalf("configured points = %v, want %v", global.DragonCityWaypoints, want)
	}
	ApplyCustomCaseParams(config, &jasmine.TestCase{})
	if global.DragonCityWaypoints != nil {
		t.Fatal("missing configuration must reset waypoints")
	}
	ApplyCustomCaseParams(config, &jasmine.TestCase{CustomParams: []jasmine.CustomParam{
		{ParamName: "--dragon-city-waypoints", ParamValue: "invalid"},
	}})
	if global.DragonCityWaypoints != nil {
		t.Fatal("invalid configuration must not set a partial route")
	}
}

func TestDragonCityMoveRate(t *testing.T) {
	originalCaseName := global.CaseName
	t.Cleanup(func() {
		global.CaseName = originalCaseName
		resetMrRobotDefaults()
	})
	for _, test := range []struct {
		name     string
		caseName string
		params   []jasmine.CustomParam
		wantRate int64
	}{
		{name: "dragon city doubled default", caseName: "DragonCity", wantRate: 10},
		{name: "explicit override", caseName: "DragonCity", params: []jasmine.CustomParam{{ParamName: "--move-rate", ParamValue: "40"}}, wantRate: 40},
		{name: "unlimited", caseName: "DragonCity", params: []jasmine.CustomParam{{ParamName: "--move-rate", ParamValue: "0"}}, wantRate: 0},
		{name: "other case unchanged", caseName: "Move", wantRate: 5},
	} {
		t.Run(test.name, func(t *testing.T) {
			global.CaseName = test.caseName
			ApplyCustomCaseParams(&jasmine.YokaConfig{PodId: 0, PodNum: 1}, &jasmine.TestCase{CustomParams: test.params})
			if global.MoveThreshold != test.wantRate || global.MoveStepPixels != 12 {
				t.Fatalf("move rate=%d step=%d, want rate=%d step=12", global.MoveThreshold, global.MoveStepPixels, test.wantRate)
			}
			if test.wantRate == 0 {
				if global.MoveRate != nil {
					t.Fatal("unlimited movement must not create a limiter")
				}
				return
			}
			if global.MoveRate == nil || global.RateLimiters["move"] != global.MoveRate {
				t.Fatal("missing shared move limiter")
			}
			now := time.Now()
			if !global.MoveRate.AllowN(now, test.wantRate) || global.MoveRate.AllowN(now, 1) {
				t.Fatalf("move limiter does not enforce rate %d", test.wantRate)
			}
		})
	}
}

func TestDragonCityBossIndex(t *testing.T) {
	t.Cleanup(resetMrRobotDefaults)
	config := &jasmine.YokaConfig{PodId: 0, PodNum: 1}
	for _, test := range []struct {
		value string
		want  int
	}{
		{value: "1", want: 1},
		{value: "2", want: 2},
		{value: "3", want: 3},
		{value: "4", want: 4},
		{value: "0", want: 1},
		{value: "5", want: 1},
		{value: "-1", want: 1},
		{value: "invalid", want: 1},
	} {
		t.Run(test.value, func(t *testing.T) {
			ApplyCustomCaseParams(config, &jasmine.TestCase{CustomParams: []jasmine.CustomParam{
				{ParamName: "--dragon-city-boss-index", ParamValue: test.value},
			}})
			if global.DragonCityBossIndex != test.want {
				t.Fatalf("boss index = %d, want %d", global.DragonCityBossIndex, test.want)
			}
			ApplyCustomCaseParams(config, &jasmine.TestCase{})
			if global.DragonCityBossIndex != 1 {
				t.Fatal("missing configuration must reset the fixed boss to 1")
			}
		})
	}
}
