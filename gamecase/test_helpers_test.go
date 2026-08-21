package gamecase

import "stressTest/module"

func newPostLoginTestRobot() *module.Robot {
	return &module.Robot{
		PTName: "post-login-test",
		Client: &module.GameClient{},
	}
}
