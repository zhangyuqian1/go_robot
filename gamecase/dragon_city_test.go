package gamecase

import (
	"context"
	"math"
	"reflect"
	"stressTest/global"
	"stressTest/module"
	"testing"
)

func TestDragonCityCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	DragonCity(ctx)
}

func TestSelectDragonCityTargetUsesClosestSyncedEntity(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.Handle = 1
	r.Client.PX, r.Client.PY = 1000, 1000
	r.Client.Entities = map[int64]*module.EntityState{
		1: {Handle: 1, X: 1000, Y: 1000},
		2: {Handle: 2, X: 1100, Y: 1000},
		3: {Handle: 3, X: 1030, Y: 1000},
		4: {Handle: 4, X: 1010, Y: 1000},
	}
	r.Client.Monsters = map[int64]*module.MonsterState{
		4: {Handle: 4, CurHP: 0},
	}
	if target := selectDragonCityTarget(r); target == nil || target.Handle != 3 {
		t.Fatalf("nearest synced target = %+v, want entity 3", target)
	}
	delete(r.Client.Entities, 3)
	if target := selectDragonCityTarget(r); target == nil || target.Handle != 2 {
		t.Fatalf("target after disappearance = %+v, want entity 2", target)
	}
	r.Client.SelfDead = true
	if target := selectDragonCityTarget(r); target != nil {
		t.Fatalf("dead robot selected target %+v", target)
	}
}

func TestFightDragonCityStopsOnDeathSceneOrCancellation(t *testing.T) {
	for _, test := range []struct {
		name string
		stop func(*module.Robot)
	}{
		{"death", func(r *module.Robot) { r.Client.SelfDead = true }},
		{"scene", func(r *module.Robot) { r.Client.SceneEntries <- module.SceneEntry{SceneID: 1032} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := newPostLoginTestRobot()
			r.Client.IsWebConnected = true
			r.Client.SceneEntries = make(chan module.SceneEntry, 1)
			test.stop(r)
			if !fightDragonCity(context.Background(), r) {
				t.Fatal("death or new scene must end attack for revival")
			}
		})
	}
	r := newPostLoginTestRobot()
	r.Client.IsWebConnected = true
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if fightDragonCity(ctx, r) {
		t.Fatal("cancellation must stop the case")
	}
}

func TestMoveTowardsPositionConverges(t *testing.T) {
	for bossIndex, position := range [][2]int32{{6800, 500}, {6800, 3200}, {6900, 3200}, {3900, 3200}, {1000, 3200}, {3900, 5600}} {
		bossX, bossY := position[0], position[1]
		for _, step := range []int32{-1, 0, 1, 12, 100, 10000} {
			curX, curY := int32(7250), int32(7450)
			for attempt := 0; attempt < 15000 && (curX != bossX || curY != bossY); attempt++ {
				destX, destY := moveTowardsPosition(curX, curY, bossX, bossY, step)
				if destX != curX && destY != curY {
					t.Fatal("a movement step must change only one axis")
				}
				before := math.Hypot(float64(bossX-curX), float64(bossY-curY))
				after := math.Hypot(float64(bossX-destX), float64(bossY-destY))
				if after >= before {
					t.Fatalf("boss %d step %d: movement did not approach target", bossIndex, step)
				}
				curX, curY = destX, destY
			}
			if curX != bossX || curY != bossY {
				t.Fatalf("boss %d step %d: stopped at (%d,%d)", bossIndex, step, curX, curY)
			}
		}
	}
}

func TestMoveTowardsPositionNearAndAtTarget(t *testing.T) {
	for _, test := range []struct {
		position [2]int32
		want     [2]int32
	}{
		{[2]int32{6899, 3201}, [2]int32{6900, 3201}},
		{[2]int32{6900, 3201}, [2]int32{6900, 3200}},
		{[2]int32{6900, 3199}, [2]int32{6900, 3200}},
		{[2]int32{6901, 3200}, [2]int32{6900, 3200}},
		{[2]int32{6900, 3200}, [2]int32{6900, 3200}},
	} {
		destX, destY := moveTowardsPosition(test.position[0], test.position[1], 6900, 3200, 12)
		if [2]int32{destX, destY} != test.want {
			t.Fatalf("from %v: got (%d,%d), want %v", test.position, destX, destY, test.want)
		}
	}
}

func TestWaitDragonCitySceneSkipsLoginScene(t *testing.T) {
	for _, spawnX := range []float32{1150, 6850} {
		entries := make(chan module.SceneEntry, 3)
		entries <- module.SceneEntry{SceneID: 1, X: 7250, Y: 7450}
		entries <- module.SceneEntry{X: 7250, Y: 7450}
		want := module.SceneEntry{SceneID: 42, FubenID: 7, X: spawnX, Y: 550}
		entries <- want
		entry, entered := waitDragonCityScene(context.Background(), entries)
		if !entered || entry != want {
			t.Fatalf("scene = %+v, entered=%v; want %+v", entry, entered, want)
		}
		wantGroup := int32(11)
		if spawnX == 6850 {
			wantGroup = 68
		}
		if group := dragonCitySpawnGroup(entry.X); group != wantGroup {
			t.Fatalf("scene x=%v group=%d, want %d", entry.X, group, wantGroup)
		}
	}
}

func TestWaitDragonCitySceneAcceptsInstance(t *testing.T) {
	entries := make(chan module.SceneEntry, 1)
	want := module.SceneEntry{SceneID: 1, FubenID: 7, X: 1150, Y: 550}
	entries <- want
	if entry, entered := waitDragonCityScene(context.Background(), entries); !entered || entry != want {
		t.Fatalf("instance entry = %+v, entered=%v", entry, entered)
	}
}

func TestWaitDragonCitySceneCancellationAndClosedChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	entries := make(chan module.SceneEntry)
	result := make(chan bool, 1)
	go func() {
		_, entered := waitDragonCityScene(ctx, entries)
		result <- entered
	}()
	entries <- module.SceneEntry{SceneID: 1, X: 7250, Y: 7450}
	cancel()
	if <-result {
		t.Fatal("login scene must not be used as the dragon city spawn")
	}
	close(entries)
	if _, entered := waitDragonCityScene(context.Background(), entries); entered {
		t.Fatal("closed channel must not report entry")
	}
	buffered := make(chan module.SceneEntry, 1)
	buffered <- module.SceneEntry{SceneID: 42, X: 1150}
	if _, entered := waitDragonCityScene(ctx, buffered); entered {
		t.Fatal("cancelled wait must not consume an entry")
	}
}

func TestWaitDragonCityRevival(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.IsWebConnected = true
	r.Client.SceneEntries = make(chan module.SceneEntry, 3)
	previous := module.SceneEntry{SceneID: 1032, X: 6850, Y: 450}
	r.Client.SceneEntries <- module.SceneEntry{SceneID: 1, X: 7250, Y: 7450}
	r.Client.SceneEntries <- module.SceneEntry{SceneID: 1032, FubenID: 9, X: 6850}
	want := module.SceneEntry{SceneID: 1032, X: 1150, Y: 450}
	r.Client.SceneEntries <- want
	entry, revived := waitDragonCityRevival(context.Background(), r, previous)
	if !revived || entry != want || dragonCitySpawnGroup(entry.X) != 11 {
		t.Fatalf("revival = %+v, revived=%v", entry, revived)
	}
	want.X = 6850
	r.Client.SceneEntries <- want
	entry, revived = waitDragonCityRevival(context.Background(), r, entry)
	if !revived || entry != want || dragonCitySpawnGroup(entry.X) != 68 {
		t.Fatalf("second revival = %+v, revived=%v", entry, revived)
	}
}

func TestWaitDragonCityRevivalStops(t *testing.T) {
	r := newPostLoginTestRobot()
	r.Client.IsWebConnected = true
	r.Client.SceneEntries = make(chan module.SceneEntry)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan bool, 1)
	go func() {
		_, revived := waitDragonCityRevival(ctx, r, module.SceneEntry{SceneID: 1032})
		result <- revived
	}()
	r.Client.SceneEntries <- module.SceneEntry{SceneID: 1}
	cancel()
	if <-result {
		t.Fatal("cancelled revival must stop")
	}
	r.Client.IsWebConnected = false
	if _, revived := waitDragonCityRevival(context.Background(), r, module.SceneEntry{}); revived {
		t.Fatal("disconnected revival must stop")
	}
	r.Client.IsWebConnected = true
	close(r.Client.SceneEntries)
	if _, revived := waitDragonCityRevival(context.Background(), r, module.SceneEntry{}); revived {
		t.Fatal("closed revival channel must stop")
	}
}

func TestDragonCitySpawnGroup(t *testing.T) {
	for _, test := range []struct {
		spawnX float32
		want   int32
	}{
		{-100, 11},
		{0, 11},
		{1100, 11},
		{1150, 11},
		{3900, 11},
		{3949.5, 11},
		{3950, 11},
		{3950.5, 68},
		{4000, 68},
		{6800, 68},
		{6850, 68},
		{7250, 68},
		{10000, 68},
	} {
		group := dragonCitySpawnGroup(test.spawnX)
		if group != test.want {
			t.Fatalf("spawn x=%v group=%d, want %d", test.spawnX, group, test.want)
		}
		routes := dragonCityRoutes(group)
		if len(routes) != 3 {
			t.Fatalf("spawn x=%v must have three candidate routes", test.spawnX)
		}
		for _, route := range routes {
			if route[0][0] != test.want {
				t.Fatalf("spawn x=%v selected wrong group route %v", test.spawnX, route)
			}
		}
	}
}

func TestDragonCityRoute(t *testing.T) {
	originalPoints, originalIndex := global.DragonCityWaypoints, global.DragonCityBossIndex
	t.Cleanup(func() {
		global.DragonCityWaypoints, global.DragonCityBossIndex = originalPoints, originalIndex
	})
	global.DragonCityWaypoints = nil
	global.DragonCityBossIndex = 4
	for _, spawnX := range []int32{11, 68} {
		for attempt := 0; attempt < 100; attempt++ {
			route := dragonCityRoute(spawnX)
			matched := false
			for _, candidate := range dragonCityRoutes(spawnX) {
				pixels := make([][2]int32, len(candidate))
				for index, point := range candidate {
					pixels[index] = [2]int32{point[0] * 100, point[1] * 100}
				}
				if reflect.DeepEqual(route, pixels) {
					matched = true
				}
			}
			if !matched {
				t.Fatalf("spawn %d selected unexpected route %v", spawnX, route)
			}
		}
	}
	for _, spawnX := range []int32{0, 10, 12, 67, 69, 72} {
		if route := dragonCityRoute(spawnX); route != nil {
			t.Fatalf("unsupported spawn %d selected route %v", spawnX, route)
		}
	}
	points := [][2]int32{{72, 32}, {69, 32}, {39, 32}, {39, 56}}
	for count := 1; count <= len(points); count++ {
		global.DragonCityWaypoints = points[:count]
		route := dragonCityRoute(0)
		if len(route) != count {
			t.Fatalf("route has %d points, want %d", len(route), count)
		}
		curX, curY := int32(7250), int32(7450)
		for index, point := range route {
			if point != [2]int32{points[index][0] * 100, points[index][1] * 100} {
				t.Fatalf("route point %d = %v", index, point)
			}
			for attempt := 0; attempt < 2000 && [2]int32{curX, curY} != point; attempt++ {
				curX, curY = moveTowardsPosition(curX, curY, point[0], point[1], 12)
			}
			if [2]int32{curX, curY} != point {
				t.Fatalf("did not reach route point %d", index)
			}
		}
		if !reflect.DeepEqual(points, [][2]int32{{72, 32}, {69, 32}, {39, 32}, {39, 56}}) {
			t.Fatal("route conversion modified configured points")
		}
	}
}

func TestDragonCityRoutesMatchDocument(t *testing.T) {
	for _, test := range []struct {
		spawnX int32
		want   [][][2]int32
	}{
		{68, [][][2]int32{
			{{68, 5}, {68, 32}, {39, 32}},
			{{68, 5}, {68, 53}, {39, 53}},
			{{68, 5}, {68, 32}},
		}},
		{11, [][][2]int32{
			{{11, 5}, {11, 33}, {39, 32}},
			{{11, 5}, {11, 54}, {39, 56}},
			{{11, 5}, {11, 33}},
		}},
	} {
		if got := dragonCityRoutes(test.spawnX); !reflect.DeepEqual(got, test.want) {
			t.Fatalf("spawn %d routes = %v, want %v", test.spawnX, got, test.want)
		}
		for routeIndex, route := range test.want {
			curX, curY := test.spawnX*100+50, int32(550)
			for pointIndex, point := range route {
				targetX, targetY := point[0]*100, point[1]*100
				for attempt := 0; attempt < 2000 && (curX != targetX || curY != targetY); attempt++ {
					curX, curY = moveTowardsPosition(curX, curY, targetX, targetY, 12)
				}
				if curX != targetX || curY != targetY {
					t.Fatalf("spawn %d route %d did not reach point %d", test.spawnX, routeIndex, pointIndex)
				}
			}
		}
	}
}

func TestDragonCityScatterDestination(t *testing.T) {
	for _, center := range [][2]int32{{3900, 3200}, {3900, 5300}, {6800, 3200}, {3900, 5600}, {1100, 3300}} {
		for direction := int32(0); direction < 360; direction++ {
			destX, destY := dragonCityScatterDestination(center[0], center[1], direction)
			distance := math.Hypot(float64(destX-center[0]), float64(destY-center[1]))
			if math.Abs(distance-200) > 1 {
				t.Fatalf("direction %d scatter distance=%v, want 200 pixels", direction, distance)
			}
		}
	}
	for _, test := range []struct {
		direction int32
		want      [2]int32
	}{
		{0, [2]int32{6900, 3400}},
		{90, [2]int32{7100, 3200}},
		{180, [2]int32{6900, 3000}},
		{270, [2]int32{6700, 3200}},
	} {
		destX, destY := dragonCityScatterDestination(6900, 3200, test.direction)
		if [2]int32{destX, destY} != test.want {
			t.Fatalf("direction %d = (%d,%d), want %v", test.direction, destX, destY, test.want)
		}
	}
}

func TestMoveToPositionEarlyExit(t *testing.T) {
	if moveToPosition(context.Background(), nil, 6900, 3200) {
		t.Fatal("nil robot must not report arrival")
	}
	r := newPostLoginTestRobot()
	if moveToPosition(context.Background(), r, 6900, 3200) {
		t.Fatal("disconnected robot must not move")
	}
	r.Client.IsWebConnected = true
	r.Client.PX, r.Client.PY = 6900, 3200
	r.Client.SelfDead = true
	if moveToPosition(context.Background(), r, 6900, 3200) {
		t.Fatal("dead robot must not report arrival or move")
	}
	r.Client.SelfDead = false
	r.Client.SceneEntries = make(chan module.SceneEntry, 1)
	r.Client.SceneEntries <- module.SceneEntry{SceneID: 1032, X: 1150, Y: 450}
	if moveToPosition(context.Background(), r, 6900, 3200) {
		t.Fatal("new scene must interrupt the previous route")
	}
	<-r.Client.SceneEntries
	if !moveToPosition(context.Background(), r, 6900, 3200) {
		t.Fatal("robot already at destination must report arrival without sending")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if moveToPosition(ctx, r, 6900, 3200) {
		t.Fatal("cancelled movement must stop")
	}
}
