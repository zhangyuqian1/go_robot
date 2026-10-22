package gamecase

import (
	"context"
	"math"
	"math/rand"
	"stressTest/global"
	"stressTest/logger"
	"stressTest/module"
	"time"
)

const (
	mainCitySceneID int32 = 1
	mainCityMinX    int32 = 6400
	mainCityMaxX    int32 = 9100
	mainCityMinY    int32 = 6200
	mainCityMaxY    int32 = 8700
	gmCoordScale    int32 = 100
)

func Move(ctx context.Context) {
	r := StartFunc()
	if r == nil {
		return
	}
	defer r.Close()
	RobotInit(r)
	time.Sleep(1 * time.Second)

	// r.CSendChatMsg(0, "@set 10066 41")
	// r.CSendChatMsg(0, "@set 2007 100000")
	// r.CSendChatMsg(0, "@set 2005 10000000")
	// r.CSendChatMsg(0, "@set 2008 100000")
	// r.CSendChatMsg(0, "@set 2009 100000")
	// r.CSendChatMsg(0, "@set 2010 100000")
	// r.CSendChatMsg(0, "@additem 950 1")
	// time.Sleep(1 * time.Second)

	// waitGiftSeries(r)
	// logger.Error(fmt.Sprintf("序列号1: %v", r.Client.GiftSeries))
	// r.CUseBagGift(3, r.Client.GiftSeries, 1)
	// // 随机选择 2~25 的英雄 id，传给后续两个协议
	// heroID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(24) + 2
	// r.CActivityHero(int32(heroID))
	// r.CSwitchHero(int32(heroID))

	// weaponID := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(120) + 1001
	// r.CSendChatMsg(0, fmt.Sprintf("@additem %d 1", weaponID))
	// time.Sleep(1 * time.Second)
	// waitWeaponSeries(r)
	// logger.Error(fmt.Sprintf("序列号2: %v", r.Client.WeaponSeries))
	// r.CTakeOnEquip(1, r.Client.WeaponSeries)

	scatterToMainCity(r)
	global.LoginRendezvous.Done(-1)
	global.LoginRendezvous.Wait()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("登出")
			return
		default:
			if global.MoveRate.Allow() {
				actorMove(r)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func scatterToMainCity(r *module.Robot) {
	if r == nil {
		return
	}
	r.CSendChatMsg(0, mainCityScatterMoveToText(r))
}

func mainCityScatterMoveToText(r *module.Robot) string {
	x, y := mainCityScatterPosition(r)
	return module.GMMoveToText(mainCitySceneID, x, y)
}

func mainCityScatterPosition(r *module.Robot) (int32, int32) {
	minX := mainCityMinX / gmCoordScale
	maxX := mainCityMaxX / gmCoordScale
	minY := mainCityMinY / gmCoordScale
	maxY := mainCityMaxY / gmCoordScale
	width := maxX - minX + 1
	height := maxY - minY + 1
	if width <= 0 || height <= 0 {
		return minX, minY
	}
	index := 0
	if r != nil {
		index = r.Index
	}
	if index < 0 {
		index = -index
	}
	x := minX + int32((index*97)%int(width))
	y := minY + int32((index*193)%int(height))
	return x, y
}

func actorMove(r *module.Robot) {
	if r == nil || r.Client == nil {
		return
	}
	curX := int32(math.Round(float64(r.Client.PX)))
	curY := int32(math.Round(float64(r.Client.PY)))
	dir := int32(rand.Intn(360))
	destX, destY := moveDestination(curX, curY, dir, global.MoveStepPixels)
	destX = clampInt32(destX, mainCityMinX, mainCityMaxX)
	destY = clampInt32(destY, mainCityMinY, mainCityMaxY)
	dir = moveDirection(curX, curY, destX, destY)
	r.Client.MoveAngle = dir
	r.Client.MoveBlocked = false
	r.CActorSetDir(dir)
	r.CActorMove(curX, curY, destX, destY, int32(time.Now().Unix()))
}

func moveToPosition(ctx context.Context, r *module.Robot, targetX, targetY int32) bool {
	if r == nil || r.Client == nil {
		return false
	}
	stalled := 0
	for {
		if ctx.Err() != nil || !r.Client.IsWebConnected || r.Client.SelfDead || len(r.Client.SceneEntries) > 0 {
			return false
		}
		if remaining := time.Until(r.Client.MoveReadyAt); remaining > 0 {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(min(remaining, 5*time.Millisecond)):
			}
			continue
		}
		curX := int32(math.Round(float64(r.Client.PX)))
		curY := int32(math.Round(float64(r.Client.PY)))
		if curX == targetX && curY == targetY {
			return true
		}
		if global.MoveRate == nil || global.MoveRate.Allow() {
			destX, destY := moveTowardsPosition(curX, curY, targetX, targetY, global.MoveStepPixels)
			dir := moveDirection(curX, curY, destX, destY)
			r.Client.MoveAngle = dir
			r.Client.MoveBlocked = false
			r.CActorSetDir(dir)
			if ctx.Err() != nil || r.Client.SelfDead || len(r.Client.SceneEntries) > 0 {
				return false
			}
			r.CActorMove(curX, curY, destX, destY, int32(time.Now().Unix()))
			r.Client.MoveReadyAt = time.Now().Add(moveTravelDuration(curX, curY, destX, destY, global.MoveSpeedPixels))
			if r.Client.SelfDead || len(r.Client.SceneEntries) > 0 {
				return false
			}
			if r.Client.MoveBlocked || (r.Client.PX == float32(curX) && r.Client.PY == float32(curY)) {
				stalled++
			} else {
				stalled = 0
			}
			if stalled >= 10 {
				logger.Errorf("robot %s cannot reach (%d,%d), stopped at (%v,%v)", r.PTName, targetX, targetY, r.Client.PX, r.Client.PY)
				return false
			}
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func moveTravelDuration(curX, curY, destX, destY, speed int32) time.Duration {
	if speed <= 0 {
		speed = 1
	}
	distance := math.Hypot(float64(destX)-float64(curX), float64(destY)-float64(curY))
	return time.Duration(math.Ceil(distance / float64(speed) * float64(time.Second)))
}

func moveTowardsPosition(curX, curY, targetX, targetY, step int32) (int32, int32) {
	if step <= 0 {
		step = 1
	}
	if curX != targetX {
		if curX < targetX {
			return curX + min(step, targetX-curX), curY
		}
		return curX - min(step, curX-targetX), curY
	}
	if curY < targetY {
		return curX, curY + min(step, targetY-curY)
	}
	return curX, curY - min(step, curY-targetY)
}

func moveDestination(curX int32, curY int32, angle int32, step int32) (int32, int32) {
	if step <= 0 {
		step = global.MoveStepPixels
	}
	radians := float64(angle) * math.Pi / 180
	destX := curX + int32(math.Round(math.Sin(radians)*float64(step)))
	destY := curY + int32(math.Round(math.Cos(radians)*float64(step)))
	return destX, destY
}

func moveDirection(curX int32, curY int32, destX int32, destY int32) int32 {
	dx := float64(destX - curX)
	dy := float64(destY - curY)
	degrees := math.Atan2(dx, dy) * 180 / math.Pi
	if degrees < 0 {
		degrees += 360
	}
	return int32(math.Round(degrees)) % 360
}

func clampInt32(value int32, min int32, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
