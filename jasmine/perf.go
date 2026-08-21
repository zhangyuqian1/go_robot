package jasmine

type Mode int

const (
	StandaloneMode Mode = iota
)

type Perf struct {
	mode        Mode
	rateLimiter RateLimiter

	localRunner *localRunner
	spawnCount  int
	spawnRate   float64
}

func NewStandaloneBoomer(spawnCount int, spawnRate float64, rateLimiter RateLimiter) *Perf {
	return &Perf{
		spawnCount:  spawnCount, // 协程数
		spawnRate:   spawnRate,  // 协程生成
		mode:        StandaloneMode,
		rateLimiter: rateLimiter,
	}
}

func (p *Perf) Run(executeMode int, executeCtl string, tasks ...*Task) {
	switch p.mode {
	case StandaloneMode:
		p.localRunner = newLocalRunner(tasks, p.rateLimiter, p.spawnCount, p.spawnRate, executeMode, executeCtl)
		p.localRunner.run()
	}
}
