package connection

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/yapingcat/gomedia/go-flv"
)

type OnFrameFunc func(codecId int, frame []byte, streamId string)
type OnPublishFunc func(streamId string)

type connection struct {
	conn          net.Conn
	bytesRead     uint32
	windowAckSize uint32
	MessageHeader MessageHeader
	BasicHeader   BasicHeader
	chunkSize     uint32
	streamName    string
	demuxer       flv.VideoTagDemuxer
	audioDemuxer  flv.AudioTagDemuxer
	onFrame       OnFrameFunc
	onPublish     OnPublishFunc
}

func (c *connection) doHandShake() {
	buffer := c.readBytes(1537)

	// todo: Add error handling here
	if buffer[0] == 0b0011 {
		c.conn.Write(buffer[0:1])
		s1rand := make([]byte, 1528)
		rand.Read(s1rand)
		theirTime := buffer[1:5]
		ownTime := make([]byte, 4)
		binary.LittleEndian.PutUint32(ownTime, uint32(time.Now().Unix()))
		s1 := append(ownTime, append(buffer[5:9], s1rand...)...)
		c.conn.Write(s1)
		c2 := c.readBytes(1536)
		s2 := append(theirTime, append(ownTime, c2[8:]...)...)
		c.conn.Write(s2)
	}
}
func (c *connection) readBytes(length uint32) []byte {
	if length <= 0 {
		return []byte{}
	}
	c.bytesRead += length
	bytes := make([]byte, length)
	io.ReadAtLeast(c.conn, bytes, int(length))

	if c.bytesRead > c.windowAckSize {
		// c.sendAck()
		c.bytesRead = 0
	}
	return bytes
}

func (c *connection) getRealChunkSize(messageLength int, chunkSize int) int {
	if chunkSize == 0 {
		return messageLength
	}
	res := messageLength + int(messageLength/chunkSize)
	if messageLength%chunkSize != 0 {
		return res
	} else {
		return res - 1
	}
}

func (c *connection) readPayload() []byte {
	if c.chunkSize == 0 {
		return c.readBytes(c.MessageHeader.MessageLength)
	}
	rtmpBody := []byte{}
	// messageLength + 1 byte * num of type 3 headers that will be included
	rtmpBodySize := c.MessageHeader.MessageLength
	chunkBodySize := c.getRealChunkSize(int(c.MessageHeader.MessageLength), int(c.chunkSize))
	// for 100 chunkBody and chunk size 10 len(chunkBody) = 110
	chunkBody := c.readBytes(uint32(chunkBodySize))

	chunkBodypos := 0

	for {
		// more than one chunk left: store entire chunk
		if rtmpBodySize > c.chunkSize {
			rtmpBody = append(rtmpBody, chunkBody[chunkBodypos:chunkBodypos+int(c.chunkSize)]...)
			rtmpBodySize -= c.chunkSize
			chunkBodypos += int(c.chunkSize)
			chunkBodypos++
			// else: store all bytes that are left
		} else {
			rtmpBody = append(rtmpBody, chunkBody[chunkBodypos:chunkBodypos+int(rtmpBodySize)]...)
			rtmpBodySize = 0
		}
		if !(rtmpBodySize > 0) {
			break
		}
	}
	if len(rtmpBody) != int(c.MessageHeader.MessageLength) {
		panic("NOT CORRECTLY READ")
		// fmt.Printf("%d, %d", c.MessageHeader.MessageLength, bytesToReadOriginal)
	}
	return rtmpBody
}

func (c *connection) readMessage() []byte {
	c.BasicHeader = c.readBasicHeader()
	c.readMessageHeader(c.BasicHeader)
	return c.readPayload()
}

func (c *connection) respondToMessage(payload []byte) (bool, error) {
	switch c.MessageHeader.MessageTypeId {
	case 0:
		// END OF STREAM
		return true, nil

	// Set Chunk Size (5.4.1)
	case 1:
		c.setChunkSize(payload)
		// fmt.Println(c.chunkSize)
	case 8:
		c.handleAudioData(payload)
	case 9:
		c.handleVideoData(payload)
	case 18:
		// fmt.Println("DATA COMMAND")
		err := c.handleMetaData(payload)
		return false, err
	case 20:
		err := c.handleAMFCommand(payload)
		if err != nil {
			panic(err)
		}
		return false, err
	default:
		return false, fmt.Errorf("UNKNOWN TYPE ID %d", c.MessageHeader.MessageTypeId)
	}
	return false, nil
}

func (c *connection) handleTraffic() {
	for {
		// fmt.Println("hello again")
		payload := c.readMessage()
		end, err := c.respondToMessage(payload)
		if err != nil {
			return
		}
		if end {
			return
		}
		// fmt.Println(payload)
	}
}

func HandleConnection(c net.Conn, onFrame OnFrameFunc, onPublish OnPublishFunc) connection {
	defer c.Close()
	// fmt.Println("hello")
	connection := connection{conn: c, windowAckSize: 100000, chunkSize: 128, onFrame: onFrame, onPublish: onPublish}
	connection.doHandShake()
	connection.handleTraffic()
	return connection
}
