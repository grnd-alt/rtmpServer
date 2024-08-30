package connection

import (
	"encoding/binary"
	"rtmp-new/m/v2/helper"
)

func (c *connection) sendMessage(basic_header BasicHeader, message_header MessageHeader, payload []byte) {

	basic := helper.CreateBasicHeader(basic_header.Fmt, int(basic_header.StreamId))
	message, _ := helper.CreateMessageHeader(0, payload, message_header.MessageTypeId, message_header.MessageStreamId)
	c.conn.Write(append(append([]byte{basic}, message...), payload...))
}

func (c *connection) sendWindowAckSize(size int) {
	ack_window_size := make([]byte, 4)
	ack_window_size[0] = byte(size >> 24)
	ack_window_size[1] = byte(size >> 16)
	ack_window_size[2] = byte(size >> 8)
	ack_window_size[3] = byte(size)
	basic_header := BasicHeader{Fmt: 0, StreamId: 2}
	message_header := MessageHeader{MessageTypeId: 5, MessageStreamId: c.MessageHeader.MessageStreamId}
	c.sendMessage(basic_header, message_header, ack_window_size)
}

func (c *connection) sendPeerBandwith(size int) {
	peer_bandwidth := make([]byte, 5)
	peer_bandwidth[0] = byte(size >> 24)
	peer_bandwidth[1] = byte(size >> 16)
	peer_bandwidth[2] = byte(size >> 8)
	peer_bandwidth[3] = byte(size)
	peer_bandwidth[4] = 1
	basic_header := helper.CreateBasicHeader(0, 2)
	message_header, _ := helper.CreateMessageHeader(0, peer_bandwidth, 6, 2)
	c.conn.Write(append(append([]byte{basic_header}, message_header...), peer_bandwidth...))
}

func (c *connection) sendAck() {
	basic_header := BasicHeader{Fmt: 0, StreamId: 2}
	message_header := MessageHeader{MessageTypeId: 3, MessageStreamId: c.MessageHeader.MessageStreamId}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, uint32(c.bytesRead))
	c.sendMessage(basic_header, message_header, payload)
}

func (c *connection) setChunkSize(payload []byte) {
	c.chunkSize = uint32(binary.BigEndian.Uint32(payload))
}
