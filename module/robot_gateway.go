package module

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"stressTest/gateway"
	"stressTest/logger"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func (r *Robot) CreateConn() error {
	return r.CreateWebSocketConn()
}

func (r *Robot) CreateWebSocketConn() error {
	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
	}

	conn, _, err := dialer.Dial(r.Client.WebSocketSvr, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to %s: %s", r.Client.WebSocketSvr, err))
		return err
	}

	if err := r.gatewayHandshake(conn); err != nil {
		_ = conn.Close()
		return err
	}

	r.Client.IsWebConnected = true
	r.Client.WebSocketConn = &webSocketConn{Conn: conn}
	r.Client.PingSending = true

	r.StartRobotConn()
	r.StartAsyncMonit()
	return nil
}

func (r *Robot) gatewayHandshake(conn *websocket.Conn) error {
	var saltBytes [4]byte
	if _, err := rand.Read(saltBytes[:]); err != nil {
		return fmt.Errorf("generate client salt: %w", err)
	}
	clientSalt := binary.LittleEndian.Uint32(saltBytes[:])
	if err := conn.WriteMessage(websocket.BinaryMessage, saltBytes[:]); err != nil {
		return fmt.Errorf("send client salt: %w", err)
	}

	messageType, serverSaltBytes, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read server salt: %w", err)
	}
	if messageType != websocket.BinaryMessage {
		return fmt.Errorf("read server salt: unexpected websocket message type %d", messageType)
	}
	if len(serverSaltBytes) < 4 {
		return fmt.Errorf("read server salt: got %d bytes", len(serverSaltBytes))
	}
	serverSalt := binary.LittleEndian.Uint32(serverSaltBytes[:4])
	key := gateway.DeriveKey(clientSalt, serverSalt)

	var checkBytes [2]byte
	binary.LittleEndian.PutUint16(checkBytes[:], gateway.CRC16(key.Bytes()))
	if err := conn.WriteMessage(websocket.BinaryMessage, checkBytes[:]); err != nil {
		return fmt.Errorf("send key check: %w", err)
	}

	r.Client.ClientSalt = clientSalt
	r.Client.ServerSalt = serverSalt
	r.Client.GatewayKey = key
	r.Client.GatewayCodec = gateway.NewCodec(key.Buff)
	return nil
}

func (r *Robot) AesRobotConnRecv() error {
	for {
		select {
		case <-r.ConnStopSignal:
			logger.Info(fmt.Sprintf("Robot %s Conn Recv Stop Signal", r.PTName))
			return nil
		default:
			if r.Client == nil || r.Client.WebSocketConn == nil || r.Client.WebSocketConn.Conn == nil || !r.Client.IsWebConnected {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			messageType, packetBytes, err := r.Client.WebSocketConn.Conn.ReadMessage()
			if err != nil {
				r.CloseSocket()
				return err
			}
			if messageType != websocket.BinaryMessage {
				logger.Debug(fmt.Sprintf("Robot %s ignore websocket message type %d", r.PTName, messageType))
				continue
			}
			if r.Client.GatewayCodec == nil {
				logger.Error(fmt.Sprintf("Robot %s gateway codec is not initialized", r.PTName))
				r.CloseSocket()
				return fmt.Errorf("gateway codec is not initialized")
			}
			packet, err := r.Client.GatewayCodec.UnmarshalResponse(packetBytes)
			if err != nil {
				logger.Error(fmt.Sprintf("Gateway packet decode error:%s", err))
				r.CloseSocket()
				return err
			}
			r.routeGatewayResponse(packet.Sys, packet.MsgID, packet.ProtoBytes)
		}
	}
}

func (r *Robot) sendGatewayRequest(requestName string, sys byte, msgID byte, pb proto.Message, wait bool) error {
	if r.Client == nil || r.Client.WebSocketConn == nil || r.Client.WebSocketConn.Conn == nil || !r.Client.IsWebConnected {
		return fmt.Errorf("robot %s websocket not connected, can't send %s", r.PTName, requestName)
	}
	if r.Client.GatewayCodec == nil {
		return fmt.Errorf("robot %s gateway codec is not initialized", r.PTName)
	}

	protoBytes, err := proto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", requestName, err)
	}
	packet, err := r.Client.GatewayCodec.MarshalRequest(sys, msgID, protoBytes)
	if err != nil {
		return fmt.Errorf("marshal gateway packet %s: %w", requestName, err)
	}

	r.Client.WriteMutex.Lock()
	writeErr := r.Client.WebSocketConn.Conn.WriteMessage(websocket.BinaryMessage, packet)
	r.Client.WriteMutex.Unlock()
	if writeErr != nil {
		r.CloseSocket()
		return fmt.Errorf("send gateway packet %s: %w", requestName, writeErr)
	}
	logger.Debug(fmt.Sprintf("send gateway sys:%d msgId:%d account:%s", sys, msgID, r.PTName))
	return nil
}
