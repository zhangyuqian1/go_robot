package module

import (
	"fmt"
	"stressTest/gameproto/chat"
	"stressTest/logger"

	"google.golang.org/protobuf/proto"
)

const (
	gmSys            byte = 7
	gmSendChatMsg    byte = 1
	gmTipMsgWithStr  byte = 2
	gmSendChatResult byte = 5

	gmSendChatRequestName       = "CSendChatMsg"
	gmRequestTimeout      int64 = 10
)

func (r *Robot) CSendChatMsg(channel int32, text string) {
	if text == "" {
		return
	}
	req := &chat.CSendChatMsg_1{Channel: channel, Text: text}
	r.AsyncBegin(gmSendChatRequestName, gmRequestTimeout, false)
	if err := r.sendGatewayRequest(gmSendChatRequestName, gmSys, gmSendChatMsg, req, false); err != nil {
		r.Mark.Remove(gmSendChatRequestName)
		logger.Error(err.Error())
		return
	}
	r.AsyncWait(gmSendChatRequestName)
}

func GMMoveToText(sceneID int32, x int32, y int32) string {
	return fmt.Sprintf("@MoveTo %d %d %d", sceneID, x, y)
}

func (r *Robot) handleGMGatewayResponse(sys byte, msgID byte, body []byte) bool {
	if sys != gmSys {
		return false
	}
	switch msgID {
	case gmTipMsgWithStr:
		r.handleGMTipMsgWithStr(body)
	case gmSendChatResult:
		r.handleGMSendChatMsg(body)
	default:
		return false
	}
	return true
}

func (r *Robot) handleGMTipMsgWithStr(body []byte) {
	var resp chat.STipmsgWithStr_2
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_tipmsg_with_str_2 error: %v", err)
		r.AsyncFire(gmSendChatRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.AsyncFire(gmSendChatRequestName, int64(len(body)), STEP_SUCC, 0)
}

func (r *Robot) handleGMSendChatMsg(body []byte) {
	var resp chat.SSendChatMsg_5
	if err := proto.Unmarshal(body, &resp); err != nil {
		logger.Errorf("unmarshal s_send_chat_msg_5 error: %v", err)
		r.AsyncFire(gmSendChatRequestName, 0, STEP_FAIL, 0)
		return
	}
	r.AsyncFire(gmSendChatRequestName, int64(len(body)), STEP_SUCC, 0)
}
