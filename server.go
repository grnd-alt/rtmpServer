package rtmp

import (
	"fmt"
	"log"
	"net"
	"os"
	"rtmp-new/m/v2/rtmp/connection"
	"sync"
)

type unreadContent struct {
	*sync.Mutex
	unread []byte
}

type RtmpServer struct {
	hostname string
	port     uint
	listener net.Listener
	unread   unreadContent
}

func StartServer(hostname string, port uint, onFrame connection.OnFrameFunc, onPublish connection.OnPublishFunc) (*RtmpServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return nil, err
	}

	server := &RtmpServer{
		hostname: hostname,
		port:     port,
		listener: listener,
		unread: unreadContent{
			Mutex:  &sync.Mutex{},
			unread: []byte{},
		},
	}

	server.handleTraffic(onFrame, onPublish)

	return server, nil
}

func (server RtmpServer) handleTraffic(onFrame connection.OnFrameFunc, onPublish connection.OnPublishFunc) {
	for {
		// fmt.Println("reading")
		conn, err := server.listener.Accept()
		if err != nil {
			var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
			Error.Println(err)
		}
		go connection.HandleConnection(conn, onFrame, onPublish)
	}
}

func (server RtmpServer) GetUnread() []byte {
	server.unread.Mutex.Lock()
	defer server.unread.Mutex.Unlock()
	holder := server.unread.unread
	server.unread.unread = []byte{}
	return holder
}
