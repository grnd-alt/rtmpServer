package helper

import "time"

func BytesToUint32(data []byte) uint32 {
	res := uint32(0)
	for i := 0; i < len(data); i++ {
		res += uint32(data[i]) << (8 * (len(data) - i - 1))
	}
	return res
}

func CreateBasicHeader(message_format int, chunk_stream_id int) byte {
	if chunk_stream_id > 63 {
		panic("NOT YET IMPLEMENTED")
	}
	return ((byte(message_format) << 6) & 192) | (byte(chunk_stream_id) & 63)
}

func CreateMessageHeader(chunk_type int, payload []byte, message_type_id byte, message_stream_id uint32) (header []byte, timestamp int64) {
	timestamp = time.Now().Unix()
	switch chunk_type {
	case 0:
		header = make([]byte, 11)
		header[3] = byte(len(payload) >> 16)
		header[4] = byte(len(payload) >> 8)
		header[5] = byte(len(payload))
		header[6] = message_type_id
		header[7] = byte(message_stream_id >> 0)
		header[8] = byte(message_stream_id >> 8)
		header[9] = byte(message_stream_id >> 16)
		header[10] = byte(message_stream_id >> 24)
	case 1:
		// header := make([]byte, 7)
	case 2:
		// header := make([]byte, 3)
	default:
	}
	return
}
