package gamecase

import (
	"context"
	"testing"
	"time"
)

// TestDrawSignature 验证 Draw 案例函数签名和上下文取消
func TestDrawSignature(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		Draw(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Draw 正确退出了
	case <-time.After(3 * time.Second):
		t.Fatal("Draw did not exit within 3 seconds after cancel")
	}
}
