package gameproto_test

import (
	"testing"

	defaultmsg "stressTest/gameproto/defaultmsg"
	"stressTest/gameproto/login"
	"stressTest/gameproto/property"
)

func TestLoginProtocolTypesImport(t *testing.T) {
	_ = &login.CCheckPassword_1{}
	_ = &login.SActorList_2{}
	_ = &login.CCreateActor_4{}
	_ = &login.SCreateActorResult_3{}
	_ = &defaultmsg.CActorLogin_1{}
	_ = &defaultmsg.SActorLoginResult_1{}
	_ = &property.SActorInit_1{}
	_ = &defaultmsg.SActorEnterScene_13{}
	_ = &defaultmsg.CHeartbeat_2{}
	_ = &defaultmsg.SHeartbeat_10{}
}
