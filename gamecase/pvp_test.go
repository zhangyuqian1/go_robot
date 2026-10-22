package gamecase

import (
	"context"
	"testing"

	"stressTest/module"
)

func TestPVPCrossServerWebSocketURL(t *testing.T) {
	for _, test := range []struct {
		name    string
		current string
		target  module.TransmitTarget
		want    string
	}{
		{"plain", "ws://127.0.0.1:10000", module.TransmitTarget{Addr: "10.238.20.96", Port: 21100}, "ws://10.238.20.96:21100"},
		{"preserve TLS path query", "wss://local.example:10000/gateway?token=abc", module.TransmitTarget{Addr: "cross.example", Port: 21100}, "wss://cross.example:21100/gateway?token=abc"},
		{"IPv6", "ws://127.0.0.1:10000/ws", module.TransmitTarget{Addr: "::1", Port: 21100}, "ws://[::1]:21100/ws"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := crossServerWebSocketURL(test.current, test.target)
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestPVPCrossServerWebSocketURLRejectsInvalidAddress(t *testing.T) {
	for _, current := range []string{"", "http://localhost", "ws:///missing", "://invalid"} {
		if _, err := crossServerWebSocketURL(current, module.TransmitTarget{Addr: "localhost", Port: 21100}); err == nil {
			t.Fatalf("accepted invalid address %q", current)
		}
	}
	for _, target := range []module.TransmitTarget{{}, {Addr: "localhost", Port: -1}, {Addr: "localhost", Port: 65536}} {
		if _, err := crossServerWebSocketURL("ws://localhost", target); err == nil {
			t.Fatalf("accepted invalid target %+v", target)
		}
	}
}

func TestPVPWaitLoginScene(t *testing.T) {
	ready := make(chan struct{}, 1)
	ready <- struct{}{}
	if !waitLoginScene(context.Background(), ready) {
		t.Fatal("did not observe scene notification")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if waitLoginScene(ctx, ready) {
		t.Fatal("did not observe context cancellation")
	}
}
