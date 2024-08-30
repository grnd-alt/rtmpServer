package connection

import (
	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-flv"
)

func (c *connection) handleVideoData(payload []byte) error {
	if c.demuxer == nil {
		c.demuxer = flv.CreateFlvVideoTagHandle(flv.GetFLVVideoCodecId(payload))
		c.demuxer.OnFrame(func(codecid codec.CodecID, frame []byte, cts int) {
			c.onFrame(int(codecid), frame, c.streamName)
		})
	}
	return c.demuxer.Decode(payload)
}

func (c *connection) handleAudioData(payload []byte) error {
	if c.audioDemuxer == nil {
		c.audioDemuxer = flv.CreateAudioTagDemuxer(flv.FLV_SOUND_FORMAT((payload[0] >> 4) & 0x0F))
		c.audioDemuxer.OnFrame(func(codecid codec.CodecID, frame []byte) {
			c.onFrame(int(codecid), frame, c.streamName)
		})
	}
	return c.audioDemuxer.Decode(payload)
}
