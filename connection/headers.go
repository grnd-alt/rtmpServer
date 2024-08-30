package connection

import (
	"encoding/binary"
	"rtmp-new/m/v2/helper"
)

type BasicHeader struct {
	Bytes    []byte
	Fmt      int
	StreamId uint32
}

type MessageHeader struct {
	Timestamp       int64
	TimestampDelta  int64
	MessageLength   uint32
	MessageTypeId   uint8
	MessageStreamId uint32
}

func (c *connection) readBasicHeader() BasicHeader {
	bytes := c.readBytes(1)
	c.BasicHeader.Fmt = int((bytes[0] & 0b11000000) >> 6)
	basicHeader := BasicHeader{
		Bytes:    bytes,
		Fmt:      int((bytes[0] & 0b11000000) >> 6),
		StreamId: uint32(bytes[0] & 0b00111111),
	}
	if c.BasicHeader.Fmt == 3 {
		c.BasicHeader.Bytes = bytes
		c.BasicHeader.Fmt = 3
		return c.BasicHeader
	}
	if basicHeader.StreamId == 0 {
		basicHeader.StreamId = 64 + uint32(c.readBytes(1)[0])
		return basicHeader
	}

	if basicHeader.StreamId == 1 {
		basicHeader.StreamId = 64 + uint32(c.readBytes(1)[0]) + 256*uint32(c.readBytes(1)[0])
	}

	return basicHeader
}

func (c *connection) readMessageHeader(basicHeader BasicHeader) {
	switch basicHeader.Fmt {
	case 0:
		bytes := c.readBytes(11)
		c.MessageHeader.Timestamp = int64(helper.BytesToUint32(bytes[0:3]))
		c.MessageHeader.MessageLength = binary.BigEndian.Uint32(append(make([]byte, 1), bytes[3:6]...))
		c.MessageHeader.MessageTypeId = bytes[6]
		c.MessageHeader.MessageStreamId = binary.BigEndian.Uint32(bytes[7:])

		if c.MessageHeader.Timestamp == 0xFFFFFF {
			timestampBytes := c.readBytes(4)
			c.MessageHeader.Timestamp = int64(helper.BytesToUint32(timestampBytes))
		}
	case 1:
		bytes := c.readBytes(7)
		c.MessageHeader.TimestampDelta = int64(helper.BytesToUint32(bytes[0:3]))
		if c.MessageHeader.TimestampDelta == 0xFFFFFF {
			timestampDeltaBytes := c.readBytes(4)
			c.MessageHeader.TimestampDelta = int64(helper.BytesToUint32(timestampDeltaBytes))
		}
		c.MessageHeader.MessageLength = binary.BigEndian.Uint32(append(make([]byte, 1), bytes[3:6]...))
		c.MessageHeader.MessageTypeId = bytes[6]
	case 2:
		bytes := c.readBytes(3)
		c.MessageHeader.TimestampDelta = int64(helper.BytesToUint32(bytes[0:3]))
		if c.MessageHeader.TimestampDelta == 0xFFFFFF {
			timestampDeltaBytes := c.readBytes(4)
			c.MessageHeader.TimestampDelta = int64(helper.BytesToUint32(timestampDeltaBytes))
		}
	case 3:
		/*do Nothing keeps MessageHeader from preceeding chunk*/
	default:
		panic(basicHeader.Fmt)
	}
	// fmt.Printf("Chunk Type: %d \n Chunk Stream ID: %d\n Message Type ID: %d\r\n", c.BasicHeader.Fmt, c.BasicHeader.StreamId, c.MessageHeader.MessageTypeId)
}
