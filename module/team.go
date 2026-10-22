package module

import (
	"stressTest/gameproto/team"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	teamSys                 byte = 20
	teamDungeonMatch        byte = 16
	teamDungeonMatchSuccess byte = 14
	teamDungeonReplay       byte = 19

	teamDungeonMatchRequestName  = "CTeamDungeonMatch"
	teamDungeonReplayRequestName = "CTeamDungeonReplay"
)

func (r *Robot) CTeamDungeonMatch(lv int32) {
	req := &team.CTeamDungeonMatch_16{Lv: lv}
	if err := r.sendGatewayRequest(teamDungeonMatchRequestName, teamSys, teamDungeonMatch, req, false); err != nil {
		logger.Error(err.Error())
		return
	}
}

func (r *Robot) CTeamDungeonReplay(id, replay int32) {
	req := &team.CTeamDungeonReplay_19{Id: id, Replay: replay}
	r.AsyncBegin(teamDungeonReplayRequestName, 30, false)
	if err := r.sendGatewayRequest(teamDungeonReplayRequestName, teamSys, teamDungeonReplay, req, false); err != nil {
		r.Mark.Remove(teamDungeonReplayRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(teamDungeonReplayRequestName)
}

func (r *Robot) handleTeamGatewayResponse(sys, msgID byte, body []byte) bool {
	if sys != teamSys || msgID != teamDungeonMatchSuccess {
		return false
	}
	var resp team.STeamDungeonMatchSuccess_14
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_team_dungeon_match_success_14: %v", err)
		return true
	}
	if resp.GetId() <= 0 {
		logger.Errorf("invalid team dungeon match id: %d", resp.GetId())
		return true
	}
	if r.Client.TeamDungeonMatchReady != nil {
		select {
		case r.Client.TeamDungeonMatchReady <- resp.GetId():
		default:
		}
	}
	return true
}
