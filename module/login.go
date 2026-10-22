package module

import (
	"stressTest/gameproto/defaultmsg"
	"stressTest/gameproto/login"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	loginSys               byte  = 255
	loginCheckPassword     byte  = 1
	loginResult            byte  = 1
	loginActorList         byte  = 2
	loginGetActorList      byte  = 3
	loginCreateActorResult byte  = 3
	loginCreateActor       byte  = 4
	loginRequestTimeout    int64 = 30
)

const (
	loginRequestCheckPassword = "CReqCheckPassword"
	loginRequestCreateActor   = "CReqCreateActor"
	loginRequestGetActorList  = "CReqGetActorList"
	loginRequestReadyLogin    = "CClientReadyLogin"
	loginRequestActorLogin    = "CActorLogin"
	loginRequestKickSelf      = "CClientKickSelf"
	loginRequestEnterScene    = "SActorEnterScene"
)

func (r *Robot) CReqCheckPassword(selectSrv int32, name string, pwd string, loginSrv int32, rawSrv int32, ip string, md5 string, tick string, channedID int32) {
	req := &login.CCheckPassword_1{
		SelectSrv: selectSrv,
		Name:      name,
		Pwd:       pwd,
		LoginSrv:  loginSrv,
		RawSrv:    rawSrv,
		Ip:        ip,
		Md5:       md5,
		Tick:      tick,
		ChannedId: channedID,
	}
	r.AsyncBegin(loginRequestCheckPassword, loginRequestTimeout, true)
	if err := r.sendGatewayRequest(loginRequestCheckPassword, loginSys, loginCheckPassword, req, false); err != nil {
		r.Mark.Remove(loginRequestCheckPassword)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(loginRequestCheckPassword)
}
func (r *Robot) CReqCreateActor(icon int32, job int32, name string, sex int32, token string) {
	req := &login.CCreateActor_4{
		Name:  name,
		Sex:   sex,
		Job:   job,
		Icon:  icon,
		Token: token,
	}
	r.Client.CreatePending = true
	r.AsyncBegin(loginRequestCreateActor, loginRequestTimeout, true)
	if err := r.sendGatewayRequest(loginRequestCreateActor, loginSys, loginCreateActor, req, false); err != nil {
		r.Mark.Remove(loginRequestCreateActor)
		r.Client.CreatePending = false
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(loginRequestCreateActor)
}

func (r *Robot) CReqGetActorList() {
	req := &login.CGetActorList_3{}
	r.AsyncBegin(loginRequestGetActorList, loginRequestTimeout, true)
	if err := r.sendGatewayRequest(loginRequestGetActorList, loginSys, loginGetActorList, req, false); err != nil {
		r.Mark.Remove(loginRequestGetActorList)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(loginRequestGetActorList)
}

func (r *Robot) CReqClientReadyLogin(actorID int32) {
	req := &defaultmsg.CClientReadyLogin_7{ActorId: actorID}
	r.AsyncBegin(loginRequestReadyLogin, loginRequestTimeout, false)
	if err := r.sendGatewayRequest(loginRequestReadyLogin, defaultSys, defaultClientReadyLogin, req, false); err != nil {
		r.Mark.Remove(loginRequestReadyLogin)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(loginRequestReadyLogin)
}

func (r *Robot) CReqClientKickSelf() {
	req := &defaultmsg.CClientKickSelf_5{}
	if err := r.sendGatewayRequest(loginRequestKickSelf, defaultSys, defaultClientKickSelf, req, false); err != nil {
		logger.Error(err.Error())
	}
}

func (r *Robot) CReqActorLogin(accountID int32, aid int32) {
	req := &defaultmsg.CActorLogin_1{
		AccountId: accountID,
		Aid:       aid,
	}
	r.AsyncBegin(loginRequestActorLogin, loginRequestTimeout, false)
	if err := r.sendGatewayRequest(loginRequestActorLogin, defaultSys, defaultActorLoginResult, req, false); err != nil {
		r.Mark.Remove(loginRequestActorLogin)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(loginRequestActorLogin)
}

func (r *Robot) handleLoginGatewayResponse(sys byte, msgID byte, body []byte) bool {
	if sys != loginSys {
		return false
	}
	switch msgID {
	case loginResult:
		r.handleLoginResult(body)
		return true
	case loginActorList:
		r.handleActorList(body)
		return true
	case loginCreateActorResult:
		r.handleCreateActorResult(body)
		return true
	default:
		return false
	}
}

func (r *Robot) handleLoginResult(body []byte) {
	var resp login.SLoginResult_1
	if err := proto.Unmarshal(body, &resp); err != nil {
		r.AsyncFire(loginRequestCheckPassword, 0, STEP_FAIL, 0)
		return
	}
	r.Client.LastLoginCode = resp.GetCode()
	if resp.GetCode() != 0 {
		r.AsyncFire(loginRequestCheckPassword, 0, STEP_FAIL, resp.GetCode())
		return
	}
	r.AsyncFire(loginRequestCheckPassword, 0, STEP_SUCC, 0)
}

func (r *Robot) handleActorList(body []byte) {
	requestName := r.actorListRequestName()
	var resp login.SActorList_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		r.AsyncFire(requestName, 0, STEP_FAIL, 0)
		return
	}

	r.Client.AccountID = resp.GetAccountId()
	actors := resp.GetActors()
	if len(actors) == 0 {
		r.Client.ActorID = 0
		r.Client.ActorName = ""
		r.AsyncFire(requestName, int64(len(body)), STEP_SUCC, 0)
		return
	}

	actor := actors[0]
	r.Client.ActorID = actor.GetActorId()
	r.Client.ActorName = actor.GetName()
	r.Client.CreatePending = false
	r.AsyncFire(requestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) actorListRequestName() string {
	if r.Mark.Has(loginRequestGetActorList) {
		return loginRequestGetActorList
	}
	return loginRequestCheckPassword
}

func (r *Robot) handleCreateActorResult(body []byte) {
	var resp login.SCreateActorResult_3
	if err := proto.Unmarshal(body, &resp); err != nil {
		r.AsyncFire(loginRequestCreateActor, 0, STEP_FAIL, 0)
		return
	}
	r.Client.LastCreateActorResult = resp.GetResult()
	if resp.GetResult() != 0 {
		r.AsyncFire(loginRequestCreateActor, int64(len(body)), STEP_FAIL, resp.GetResult())
		return
	}
	r.Client.ActorID = resp.GetAid()
	r.AsyncFire(loginRequestCreateActor, int64(len(body)), STEP_SUCC, 0)
}
