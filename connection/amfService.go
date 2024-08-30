package connection

import (
	"bytes"
	"fmt"
	"rtmp-new/m/v2/helper"

	"github.com/moggle-mog/goav/amf"
)

func (c *connection) handleAMFConnect() error {
	c.sendWindowAckSize(int(c.windowAckSize))
	c.sendPeerBandwith(5000000)

	connectResp := []interface{}{"_result", 1, amf.Object{"fmsVer": "FMS/3,0,1,123", "capabilities": 31, "mode": 1.0}, 10}
	encoded := bytes.NewBuffer([]byte{})
	err := amf.NewEnDecAMF0().EncodeBatch(encoded, connectResp...)
	if err != nil {
		return err
	}
	encPayload := encoded.Bytes()
	basicHeader := helper.CreateBasicHeader(0, 3)
	messageHeader, _ := helper.CreateMessageHeader(0, encPayload, 20, 0)
	c.conn.Write(append(append([]byte{basicHeader}, messageHeader...), encPayload...))
	return nil
}

func (c *connection) handleAMFRelease(decoded []interface{}) error {
	resp := []interface{}{"_result", decoded[1], nil, c.BasicHeader.StreamId}
	encoded := bytes.NewBuffer([]byte{})
	err := amf.NewEnDecAMF0().EncodeBatch(encoded, resp...)
	if err != nil {
		return err
	}
	encPayload := encoded.Bytes()
	basicHeader := helper.CreateBasicHeader(0, 3)
	messageHeader, _ := helper.CreateMessageHeader(0, encPayload, 20, 0)
	c.conn.Write(append(append([]byte{basicHeader}, messageHeader...), encPayload...))
	return nil
}

func (c *connection) handleAMFFCPublish(decoded []interface{}) error {
	c.streamName = decoded[3].(string)
	c.onPublish(c.streamName)
	resp := []interface{}{"_result", decoded[1], nil}
	encoded := bytes.NewBuffer([]byte{})
	err := amf.NewEnDecAMF0().EncodeBatch(encoded, resp...)
	if err != nil {
		return err
	}
	encPayload := encoded.Bytes()
	basicHeader := helper.CreateBasicHeader(0, 3)
	messageHeader, _ := helper.CreateMessageHeader(0, encPayload, 20, 0)
	c.conn.Write(append(append([]byte{basicHeader}, messageHeader...), encPayload...))
	return nil
}

func (c *connection) handleAMFCreateStream(decoded []interface{}) error {
	resp := []interface{}{"_result", decoded[1], nil}
	encoded := bytes.NewBuffer([]byte{})
	err := amf.NewEnDecAMF0().EncodeBatch(encoded, resp...)
	if err != nil {
		return err
	}
	encPayload := encoded.Bytes()
	basicHeader := helper.CreateBasicHeader(0, 3)
	messageHeader, _ := helper.CreateMessageHeader(0, encPayload, 20, 0)
	c.conn.Write(append(append([]byte{basicHeader}, messageHeader...), encPayload...))
	return nil
}

func (c *connection) publishStream() error {
	payload := []interface{}{"onStatus", 0, nil, amf.Object{"level": "status", "code": "NetStream.Publish.Start", "description": "Stream is now published"}}
	encoded := bytes.NewBuffer([]byte{})
	err := amf.NewEnDecAMF0().EncodeBatch(encoded, payload...)
	if err != nil {
		return err
	}
	encPayload := encoded.Bytes()
	basicHeader := helper.CreateBasicHeader(0, 3)
	messageHeader, _ := helper.CreateMessageHeader(0, encPayload, 20, 0)
	c.conn.Write(append(append([]byte{basicHeader}, messageHeader...), encPayload...))
	return nil
}

func (c *connection) handleMetaData(payload []byte) error {
	decoded, err := amf.NewEnDecAMF0().DecodeBatch(bytes.NewReader(payload))
	if err != nil {
		return err
	}
	if decoded[0] == "@setDataFrame" {
		if decoded[1] == "onMetaData" {
			// c.channel <- payload
			return nil
		}
	}
	return nil
}

func (c *connection) handleAMFCommand(payload []byte) error {
	decoded, err := amf.NewEnDecAMF0().DecodeBatch(bytes.NewReader(payload))
	if err != nil {
		return err
	}
	switch decoded[0] {
	case "connect":
		return c.handleAMFConnect()
	case "releaseStream":
		return c.handleAMFRelease(decoded)
	case "FCPublish":
		return c.handleAMFFCPublish(decoded)
	case "createStream":
		return c.handleAMFCreateStream(decoded)
	case "publish":
		return c.publishStream()
	default:
		return fmt.Errorf("%s is unknown", decoded[0])
	}
}
