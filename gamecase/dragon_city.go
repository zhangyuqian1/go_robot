package gamecase

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

func DragonCity(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	r.Client.LoginSceneReady = make(chan struct{}, 1)
	RobotInit(r)
	if !r.Client.IsWebConnected {
		return
	}
	rendezvousPending := true
	defer func() {
		if rendezvousPending {
			global.LoginRendezvous.Done(-1)
		}
	}()
	if !waitLoginScene(ctx, r.Client.LoginSceneReady) {
		logger.Errorf("robot %s local login did not enter scene", r.PTName)
		return
	}
	accountID, actorID := r.Client.AccountID, r.Client.ActorID
	r.Client.TransmitTarget = make(chan module.TransmitTarget, 1)
	r.CDragonCityEnter()
	var target module.TransmitTarget
	transmitTimer := time.NewTimer(30 * time.Second)
	defer transmitTimer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-transmitTimer.C:
		logger.Errorf("robot %s timed out waiting for dragon city transmit", r.PTName)
		return
	case target = <-r.Client.TransmitTarget:
	}
	address, err := crossServerWebSocketURL(r.Client.WebSocketSvr, target)
	if err != nil {
		logger.Errorf("robot %s invalid dragon city route: %v", r.PTName, err)
		return
	}
	r.CReqClientKickSelf()
	r.Close()
	if ctx.Err() != nil {
		return
	}
	r.Client.WebSocketSvr = address
	r.Client.LoginDone = false
	r.Client.LoginStarted = false
	r.Client.Handle = 0
	r.Client.AccountID = 0
	r.Client.ActorID = 0
	r.Client.LastLoginCode = 0
	r.Client.LastActorLoginResult = 0
	r.Client.TransmitTarget = nil
	r.Client.Entities = nil
	r.Client.Monsters = nil
	r.Client.BattleTargetHandle = 0
	r.Client.LoginSceneReady = make(chan struct{}, 1)
	r.Client.SceneEntries = make(chan module.SceneEntry, 1)
	r.TickStopSignal = make(chan struct{})
	r.ConnStopSignal = make(chan struct{})
	r.AsyncMonitSignal = make(chan struct{})
	for global.LoginRate != nil && !global.LoginRate.Allow() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Millisecond):
		}
	}
	if ctx.Err() != nil {
		return
	}
	logger.Infof("robot %s connecting to dragon city gateway %s", r.PTName, address)
	if err := r.CreateWebSocketConn(); err != nil {
		logger.Errorf("robot %s dragon city connection failed: %v", r.PTName, err)
		return
	}

	//r.CReqCheckPassword(120, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 120, 120, target.Addr, "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	//r.CReqCheckPassword(121, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 121, 121, target.Addr, "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	//r.CReqCheckPassword(125, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 125, 125, target.Addr, "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	r.CReqCheckPassword(126, r.PTName, "e10adc3949ba59abbe56e057f20f883e", 126, 126, "127.0.0.1", "", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	if !r.Client.IsWebConnected || r.Client.LastLoginCode != 0 || r.Client.AccountID != accountID || r.Client.ActorID == 0 {
		logger.Errorf("robot %s dragon city authentication failed, code=%d account=%d actor=%d", r.PTName, r.Client.LastLoginCode, r.Client.AccountID, r.Client.ActorID)
		return
	}
	r.CReqClientReadyLogin(actorID)
	r.CReqActorLogin(accountID, actorID)
	entry, entered := waitDragonCityScene(ctx, r.Client.SceneEntries)
	if !entered {
		logger.Errorf("robot %s dragon city target scene not received, latest sid=%d fid=%d x=%v y=%v", r.PTName, r.Client.SceneID, r.Client.FubenID, r.Client.PX, r.Client.PY)
		return
	}
	global.LoginRendezvous.Done(-1)
	rendezvousPending = false
	global.LoginRendezvous.Wait()

	for {
		spawnGroup := dragonCitySpawnGroup(entry.X)
		route := dragonCityRoute(spawnGroup)
		if len(route) == 0 {
			logger.Errorf("robot %s no dragon city route for spawn x=%v group=%d", r.PTName, entry.X, spawnGroup)
			return
		}
		logger.Infof("robot %s dragon city scene=%d fid=%d spawn x=%v y=%v distance11=%v distance68=%v group=%d selected route=%v", r.PTName, entry.SceneID, entry.FubenID, entry.X, entry.Y, math.Abs(float64(entry.X)-float64(11*gmCoordScale)), math.Abs(float64(entry.X)-float64(68*gmCoordScale)), spawnGroup, route)
		lastPoint := route[len(route)-1]
		direction := rand.Intn(360)
		scatterX, scatterY := dragonCityScatterDestination(lastPoint[0], lastPoint[1], int32(direction))
		waypointCount := len(route)
		route = append(route, [2]int32{scatterX, scatterY})
		completed := true
		r.CSendChatMsg(0, "@attr 5 1000000 11 500000 9 50000")
		r.CSetPkMode(5)
		for index, point := range route {
			if !moveToPosition(ctx, r, point[0], point[1]) {
				if !r.Client.SelfDead && len(r.Client.SceneEntries) == 0 {
					return
				}
				completed = false
				break
			}
			if index < waypointCount {
				logger.Infof("robot %s arrived at dragon city waypoint %d (%d,%d)", r.PTName, index+1, point[0], point[1])
			}
		}
		if ctx.Err() != nil || !r.Client.IsWebConnected {
			return
		}
		if completed && !r.Client.SelfDead && len(r.Client.SceneEntries) == 0 {
			logger.Infof("robot %s dragon city route completed, scattered dir=%d to (%d,%d), attacking nearby entities", r.PTName, direction, scatterX, scatterY)
			if !fightDragonCity(ctx, r) {
				return
			}
		} else {
			logger.Infof("robot %s dragon city route interrupted, waiting for automatic revival scene", r.PTName)
		}
		var revived bool
		entry, revived = waitDragonCityRevival(ctx, r, entry)
		if !revived {
			return
		}
		logger.Infof("robot %s dragon city revived at (%v,%v), selecting a new route", r.PTName, entry.X, entry.Y)
	}
}

func fightDragonCity(ctx context.Context, r *module.Robot) bool {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil || !r.Client.IsWebConnected {
			return false
		}
		if r.Client.SelfDead || len(r.Client.SceneEntries) > 0 {
			return true
		}
		if target := selectDragonCityTarget(r); target != nil && (global.BattleRate == nil || global.BattleRate.Allow()) {
			skillID := r.Client.MinSkillID
			if skillID <= 0 {
				skillID = battleDefaultSkillID
			}
			monsterTarget := &module.MonsterState{Handle: target.Handle, X: target.X, Y: target.Y}
			dir := battleDirection(r, monsterTarget)
			r.CStartUseSkill(r.Client.Handle, skillID, dir, 0)
			if ctx.Err() != nil || r.Client.SelfDead || len(r.Client.SceneEntries) > 0 {
				continue
			}
			r.CSkillFrameProxy(buildBattleSkillFrameRequest(r, monsterTarget, skillID, dir))
			if ctx.Err() != nil || r.Client.SelfDead || len(r.Client.SceneEntries) > 0 {
				continue
			}
			if r.Client.Entities[target.Handle] != nil {
				r.CDamageEntity(target.Handle, skillID, battleRangeIndex(skillID), 0)
			}
		}
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
		}
	}
}

func selectDragonCityTarget(r *module.Robot) *module.EntityState {
	if r == nil || r.Client == nil || r.Client.SelfDead {
		return nil
	}
	var nearest *module.EntityState
	var nearestDistance float64
	for handle, entity := range r.Client.Entities {
		if handle == r.Client.Handle || handle == 0 || entity == nil {
			continue
		}
		if monster := r.Client.Monsters[handle]; monster != nil && monster.CurHP == 0 {
			continue
		}
		dx := float64(entity.X) - float64(r.Client.PX)
		dy := float64(entity.Y) - float64(r.Client.PY)
		distance := dx*dx + dy*dy
		if nearest == nil || distance < nearestDistance || (distance == nearestDistance && handle < nearest.Handle) {
			nearest, nearestDistance = entity, distance
		}
	}
	return nearest
}

func waitDragonCityRevival(ctx context.Context, r *module.Robot, previous module.SceneEntry) (module.SceneEntry, bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil || !r.Client.IsWebConnected {
			return module.SceneEntry{}, false
		}
		select {
		case <-ctx.Done():
			return module.SceneEntry{}, false
		case <-ticker.C:
		case entry, open := <-r.Client.SceneEntries:
			if !open {
				return module.SceneEntry{}, false
			}
			if entry.SceneID == previous.SceneID && entry.FubenID == previous.FubenID {
				return entry, true
			}
		}
	}
}

func waitDragonCityScene(ctx context.Context, entries <-chan module.SceneEntry) (module.SceneEntry, bool) {
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	for {
		if ctx.Err() != nil {
			return module.SceneEntry{}, false
		}
		select {
		case <-ctx.Done():
			return module.SceneEntry{}, false
		case <-timer.C:
			return module.SceneEntry{}, false
		case entry, open := <-entries:
			if !open {
				return module.SceneEntry{}, false
			}
			if entry.SceneID > 0 && (entry.SceneID != mainCitySceneID || entry.FubenID != 0) {
				return entry, true
			}
		}
	}
}

func dragonCitySpawnGroup(spawnX float32) int32 {
	distanceTo68 := math.Abs(float64(spawnX) - float64(68*gmCoordScale))
	distanceTo11 := math.Abs(float64(spawnX) - float64(11*gmCoordScale))
	if distanceTo68 < distanceTo11 {
		return 68
	}
	return 11
}

func dragonCityRoute(spawnX int32) [][2]int32 {
	waypoints := global.DragonCityWaypoints
	if len(waypoints) == 0 {
		routes := dragonCityRoutes(spawnX)
		if len(routes) == 0 {
			return nil
		}
		waypoints = routes[rand.Intn(len(routes))]
	}
	points := make([][2]int32, len(waypoints))
	for index, point := range waypoints {
		points[index] = [2]int32{point[0] * gmCoordScale, point[1] * gmCoordScale}
	}
	return points
}

func dragonCityRoutes(spawnX int32) [][][2]int32 {
	switch spawnX {
	case 68:
		return [][][2]int32{
			{{68, 5}, {68, 32}, {39, 32}},
			{{68, 5}, {68, 53}, {39, 53}},
			{{68, 5}, {68, 32}},
		}
	case 11:
		return [][][2]int32{
			{{11, 5}, {11, 33}, {39, 32}},
			{{11, 5}, {11, 54}, {39, 56}},
			{{11, 5}, {11, 33}},
		}
	default:
		return nil
	}
}

func dragonCityScatterDestination(curX, curY, direction int32) (int32, int32) {
	return moveDestination(curX, curY, direction, 2*gmCoordScale)
}
