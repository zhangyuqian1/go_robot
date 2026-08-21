package gamecase

import (
	"context"
	"testing"
	"time"
)

// TestPraySignature 验证 Pray 案例函数签名和上下文取消
func TestPraySignature(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		Pray(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Pray 正确退出了
	case <-time.After(3 * time.Second):
		t.Fatal("Pray did not exit within 3 seconds after cancel")
	}
}
