/******************************************************************************
MIT License

Copyright (c) 2022 jung hyun, ko

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
 *****************************************************************************/

package gosof

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// InitTcpServer
// This is an alias for InitTcpServerListenConfig that doesn't use ListenConfig .
func (h *Server) InitTcpServer(network string, ip string, port uint16) error {
	return h.InitTcpServerListenConfig(network, ip, port, nil)
}

// InitTcpServerListenConfig
// ListenConfig configuration is very platform specific.
// For example, SO_REUSEPORT does not exist on Windows.
// --> use parameter.
func (h *Server) InitTcpServerListenConfig(network string, ip string, port uint16, lc *net.ListenConfig) error {
	// network : "tcp", "tcp4", "tcp6"
	log.SetFlags(log.Llongfile)
	connStr := fmt.Sprintf("%s:%d", ip, port)
	_, h.GosofErr = net.ResolveTCPAddr(network, connStr)
	if h.GosofErr != nil {
		return h.GosofErr
	}
	if h.maxDataByteLenLimit == 0 {
		h.maxDataByteLenLimit = 1024 * 1024 * 1024 //default 1 GB
	}
	if h.readClientTimeOut == 0 {
		h.readClientTimeOut = 60 * 60 //default 1 hour
	}
	if h.calculateDataLenCb == nil {
		h.GosofErr = errors.New("error : OnCalculateDataLen not set")
		return h.GosofErr
	} else if h.completeDataCb == nil {
		h.GosofErr = errors.New("error : OnCompleteData not set")
		return h.GosofErr
	}
	if lc != nil {
		h.listener, h.GosofErr = lc.Listen(context.Background(), network, connStr)
	} else {
		h.listener, h.GosofErr = net.Listen(network, connStr)
	}
	if h.GosofErr != nil {
		log.Println("InitServer error : ", h.GosofErr.Error())
		return h.GosofErr
	}
	//log.Println("server starts : ", h.listener.Addr().String())
	if h.initCompletedCb != nil {
		h.initCompletedCb()
	}

	go func() {
		defer func() {
			_ = h.listener.Close()
			log.Println("listener closed")
		}()
		for {
			conn, err := h.listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				log.Fatal(err)
				return
			}
			ctx := Context{Conn: conn, IsDataLenCalculated: false}
			if h.newClientCb != nil {
				h.newClientCb(&ctx)
			}
			go func(ctx *Context) {
				if h.readClientTimeOut > 0 {
					deadLineErr := ctx.Conn.SetReadDeadline(time.Now().
						Add(time.Duration(maxReadTimeOutSecs) * time.Second))
					if deadLineErr != nil {
						log.Println("SetReadDeadLine error : ", deadLineErr.Error())
						return
					}
				}
				h.Common.tcpBufferWork(ctx)
			}(&ctx)
		} // for
	}()
	return nil
}

// InitTcpClient
// network : "tcp", "tcp4", "tcp6"
func (h *Client) InitTcpClient(network string, ip string, port uint16, timeout uint16) error {
	connStr := fmt.Sprintf("%s:%d", ip, port)
	var connErr error
	var svrConn net.Conn
	_, resolveErr := net.ResolveTCPAddr(network, connStr)
	if resolveErr != nil {
		return resolveErr
	}
	if timeout > 0 {
		svrConn, connErr = net.DialTimeout(network, connStr, time.Duration(timeout)*time.Second)
	} else {
		svrConn, connErr = net.Dial(network, connStr)
	}
	h.Ctx.Conn = svrConn
	h.Ctx.IsDataLenCalculated = false
	if connErr != nil {
		log.Println("InitClient error : ", connErr.Error())
		return connErr
	}
	if h.serverConnectedCb != nil {
		h.serverConnectedCb(&h.Ctx)
	}
	if h.initCompletedCb != nil {
		h.initCompletedCb()
	}
	go h.Common.tcpBufferWork(&h.Ctx)

	return nil
}

// SendToServer
// Acquire a lock and send multiple byte chunks.
func (h *Client) SendToServer(dataLen int, data ...[]byte) error {
	return h.Common.SendTcp(&h.Ctx, dataLen, data...)
}
