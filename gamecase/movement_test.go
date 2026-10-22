package gamecase

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestMoveTravelDuration(t *testing.T) {
	for _, test := range []struct {
		curX, curY, destX, destY, speed int32
		want                            time.Duration
	}{
		{0, 0, 12, 0, 300, 40 * time.Millisecond},
		{0, 0, 0, -12, 300, 40 * time.Millisecond},
		{0, 0, 3, 4, 100, 50 * time.Millisecond},
		{0, 0, 1, 0, 300, 3333334 * time.Nanosecond},
		{0, 0, 0, 0, 300, 0},
		{0, 0, 12, 0, 600, 20 * time.Millisecond},
		{0, 0, 1, 0, 0, time.Second},
		{0, 0, 1, 0, -1, time.Second},
	} {
		if got := moveTravelDuration(test.curX, test.curY, test.destX, test.destY, test.speed); got != test.want {
			t.Fatalf("travel duration %+v = %v, want %v", test, got, test.want)
		}
	}
}

func TestMoveToPositionWaitsForArrivalAcrossRevival(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.IsWebConnected = true
	for _, position := range [][2]int32{{3900, 3200}, {1150, 450}, {6850, 450}} {
		r.Client.PX, r.Client.PY = float32(position[0]), float32(position[1])
		r.Client.SelfDead = false
		r.Client.MoveReadyAt = time.Now().Add(20 * time.Millisecond)
		deadline := r.Client.MoveReadyAt
		if !moveToPosition(context.Background(), r, position[0], position[1]) {
			t.Fatal("arrival must complete after the movement deadline")
		}
		if time.Now().Before(deadline) {
			t.Fatal("arrival or revival must not bypass the movement deadline")
		}
	}
}

func TestMoveToPositionPacingCancellation(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.IsWebConnected = true
	r.Client.MoveReadyAt = time.Now().Add(time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if moveToPosition(ctx, r, 0, 0) {
		t.Fatal("cancelled movement must not report arrival")
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Fatal("movement should wait until cancellation")
	}
	r.Client.SelfDead = true
	if moveToPosition(context.Background(), r, 0, 0) {
		t.Fatal("death must interrupt movement pacing")
	}
}

func TestMainCityScatterMoveToTextUsesMainCityBounds(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Index = 42

	text := mainCityScatterMoveToText(r)

	if !strings.HasPrefix(text, "@MoveTo 1 ") {
		t.Fatalf("mainCityScatterMoveToText() = %q, want main city MoveTo", text)
	}
	x, y, ok := parseMoveToXY(t, text)
	if !ok {
		t.Fatalf("mainCityScatterMoveToText() = %q, want parsable coordinates", text)
	}
	if sceneID := parseMoveToSceneID(t, text); sceneID != mainCitySceneID {
		t.Fatalf("mainCityScatterMoveToText() scene = %d, want %d", sceneID, mainCitySceneID)
	}
	if x < mainCityGMMinX() || x > mainCityGMMaxX() || y < mainCityGMMinY() || y > mainCityGMMaxY() {
		t.Fatalf("mainCityScatterMoveToText() = %q, coordinates out of GM main city bounds", text)
	}
	if x >= mainCityMinX || y >= mainCityMinY {
		t.Fatalf("mainCityScatterMoveToText() = %q, GM coordinates should use main city bounds divided by 100", text)
	}
}

func TestMainCityScatterMoveToTextScattersByRobotIndex(t *testing.T) {
	r1 := newPostLoginTestRobot()
	r1.Index = 1
	r2 := newPostLoginTestRobot()
	r2.Index = 2

	text1 := mainCityScatterMoveToText(r1)
	text2 := mainCityScatterMoveToText(r2)

	if text1 == text2 {
		t.Fatalf("mainCityScatterMoveToText() did not scatter robots: %q", text1)
	}
}

func TestMainCityScatterMoveToTextFallsBackWhenRobotNil(t *testing.T) {
	text := mainCityScatterMoveToText(nil)
	x, y, ok := parseMoveToXY(t, text)
	if !ok {
		t.Fatalf("mainCityScatterMoveToText(nil) = %q, want parsable coordinates", text)
	}
	if x < mainCityGMMinX() || x > mainCityGMMaxX() || y < mainCityGMMinY() || y > mainCityGMMaxY() {
		t.Fatalf("mainCityScatterMoveToText(nil) = %q, coordinates out of GM main city bounds", text)
	}
}

func TestMainCityScatterPositionUsesPixelBoundsDividedBy100(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Index = 0
	x, y := mainCityScatterPosition(r)

	if x != mainCityGMMinX() || y != mainCityGMMinY() {
		t.Fatalf("mainCityScatterPosition(index 0) = (%d,%d), want (%d,%d)", x, y, mainCityGMMinX(), mainCityGMMinY())
	}
}

func mainCityGMMinX() int32 { return mainCityMinX / gmCoordScale }
func mainCityGMMaxX() int32 { return mainCityMaxX / gmCoordScale }
func mainCityGMMinY() int32 { return mainCityMinY / gmCoordScale }
func mainCityGMMaxY() int32 { return mainCityMaxY / gmCoordScale }

func parseMoveToXY(t *testing.T, text string) (int32, int32, bool) {
	t.Helper()
	var sceneID int32
	var x int32
	var y int32
	if _, err := fmt.Sscanf(text, "@MoveTo %d %d %d", &sceneID, &x, &y); err != nil {
		return 0, 0, false
	}
	return x, y, true
}

func parseMoveToSceneID(t *testing.T, text string) int32 {
	t.Helper()
	var sceneID int32
	var x int32
	var y int32
	if _, err := fmt.Sscanf(text, "@MoveTo %d %d %d", &sceneID, &x, &y); err != nil {
		t.Fatalf("parse scene id from %q: %v", text, err)
	}
	return sceneID
}
