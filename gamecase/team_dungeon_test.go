package gamecase

import (
	"context"
	"testing"
)

func TestTeamDungeonCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	TeamDungeon(ctx)
}
