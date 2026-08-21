package gamecase

import (
	"testing"

	"stressTest/global"
	"stressTest/module"
)

func TestFriendOnceSkipsWhenNoTarget(t *testing.T) {
	old := global.FriendAid
	global.FriendAid = 0
	defer func() { global.FriendAid = old }()

	r := newPostLoginTestRobot()
	friendOnce(r) // 目标 aid 为空时直接返回，不发包
}

func TestFriendOnceSkipsNilRobot(t *testing.T) {
	friendOnce(nil)
}

func TestFriendOnceSkipsNilClient(t *testing.T) {
	r := &module.Robot{}
	friendOnce(r)
}
