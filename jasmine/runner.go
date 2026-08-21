package jasmine

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

//const (
//	stateInit     = "ready"
//	stateSpawning = "spawning"
//	stateRunning  = "running"
//	stateStopped  = "stopped"
//	stateQuitting = "quitting"
//)

const (
	DefaultMode = iota
	TimerMode
)

type runner struct {
	state string

	tasks           []*Task
	totalTaskWeight int
	runTask         []*Task

	rateLimiter      RateLimiter
	rateLimitEnabled bool

	numClients int32
	spawnRate  float64

	executeMode int
	durationStr time.Duration

	// Cancellation method for all running workers(goroutines)
	cancelFuncs []context.CancelFunc
}

func (r *runner) safeRun(fn func(ctx context.Context), ctx context.Context) {
	defer func() {
		// don't panic
		err := recover()
		if err != nil {
			stackTrace := debug.Stack()
			errMsg := fmt.Sprintf("%v", err)
			os.Stderr.Write([]byte(errMsg))
			os.Stderr.Write([]byte("\n"))
			os.Stderr.Write(stackTrace)
		}
	}()
	fn(ctx)
}

func (r *runner) getTaskToScale() *Task {
	//if len(r.tasks) == 1 {
	//	return r.tasks[0]
	//} else {
	//	weightSum := 0
	//	for _, task := range r.tasks {
	//		if task.Weight <= 0 { //确保用户输入的值合法
	//			task.Weight = 1
	//		}
	//		weightSum += task.Weight
	//	}
	//	r.totalTaskWeight = weightSum
	//
	//	r.runTask = make([]*Task, r.totalTaskWeight)
	//	rate := rand.Float64() * float64(r.totalTaskWeight)
	//	accumPercentage := 0.0
	//	for _, task := range r.tasks {
	//		if task.Weight > 0 {
	//			accumPercentage += float64(task.Weight)
	//			if rate < accumPercentage {
	//				return task
	//			}
	//		}
	//	}
	//}
	//return nil
	if len(r.tasks) == 1 {
		return r.tasks[0]
	}

	weightSum := 0
	for _, task := range r.tasks {
		if task.Weight <= 0 { //确保用户输入的值合法
			task.Weight = 1
		}
		weightSum += task.Weight
	}
	r.totalTaskWeight = weightSum

	randomWeight := rand.Intn(r.totalTaskWeight)
	currentWeight := 0
	for _, task := range r.tasks {
		currentWeight += task.Weight
		if currentWeight > randomWeight {
			return task
		}
	}
	return nil // 理论上不会到达这里
}

func (r *runner) setExecuteMode(gapCount int) {
	switch r.executeMode {
	case TimerMode:
		durationTimer := time.NewTimer(r.durationStr)
		r.addTimerWorkers(durationTimer, gapCount)
	case DefaultMode:
		r.addCycleWorkers(gapCount)
	}
}

func (r *runner) addTimerWorkers(durationTimer *time.Timer, gapCount int) {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFuncs = append(r.cancelFuncs, cancel)

	for i := 0; i < gapCount; i++ {
		go func(ctx context.Context) {
			if r.rateLimitEnabled {
				blocked := r.rateLimiter.Acquire()
				if !blocked {
					task := r.getTaskToScale()
					r.safeRun(task.Fn, ctx)
				}
			} else {
				task := r.getTaskToScale()
				r.safeRun(task.Fn, ctx)
			}
		}(ctx)
		runtime.Gosched()
	}
	<-durationTimer.C
	cancel()
	time.Sleep(5 * time.Second) // 停止后使用
	os.Exit(0)
}

func (r *runner) addCycleWorkers(gapCount int) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFuncs = append(r.cancelFuncs, cancel)
	wg.Add(gapCount)
	for i := 0; i < gapCount; i++ {
		go func(ctx context.Context) {
			if r.rateLimitEnabled {
				blocked := r.rateLimiter.Acquire()
				if !blocked {
					task := r.getTaskToScale()
					r.safeRun(task.Fn, ctx)
				}
			} else {
				task := r.getTaskToScale()
				r.safeRun(task.Fn, ctx)
			}
			wg.Done()
		}(ctx)
		runtime.Gosched()
	}
	wg.Wait()
	cancel()
	time.Sleep(10 * time.Second) // 停止后使用
}

func (r *runner) reduceWorkers(gapCount int) {
	if gapCount == 0 {
		num := len(r.cancelFuncs) - gapCount
		for _, cancelFunc := range r.cancelFuncs[num:] {
			cancelFunc()
		}
		r.cancelFuncs = r.cancelFuncs[:num]
	}
}

func (r *runner) spawnWorkers(spawnCount int, spawnCompleteFunc func()) {
	var gapCount int
	if spawnCount > int(r.numClients) {
		gapCount = spawnCount - int(r.numClients)
		r.setExecuteMode(gapCount)

	} else {
		gapCount = int(r.numClients) - spawnCount
		r.reduceWorkers(gapCount)

	}
	r.numClients = int32(spawnCount)

	if spawnCompleteFunc != nil {
		go spawnCompleteFunc() //For faster time
	}
}

func (r *runner) startSpawning(spawnCount int, spawnRate float64, spawnCompleteFunc func()) {
	r.spawnWorkers(spawnCount, spawnCompleteFunc)
}

type localRunner struct {
	runner
	spawnCount int
}

func newLocalRunner(tasks []*Task, rateLimiter RateLimiter, spawnCount int, spawnRate float64, executeMode int, executeCtl string) (r *localRunner) {
	r = &localRunner{}
	r.tasks = tasks
	r.spawnRate = spawnRate
	r.spawnCount = spawnCount
	r.executeMode = executeMode
	if rateLimiter != nil {
		r.rateLimitEnabled = true
		r.rateLimiter = rateLimiter
	}
	switch r.executeMode {
	case TimerMode:
		duration, err := time.ParseDuration(executeCtl)
		if err != nil {
			fmt.Println("解析时间段失败：", err)
		}
		r.durationStr = duration
	}
	return r
}

func (r *localRunner) run() {
	//r.state = stateInit
	r.startSpawning(r.spawnCount, r.spawnRate, nil)
}
