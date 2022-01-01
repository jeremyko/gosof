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
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// InitUdpServer
// network : "udp", "udp4", "udp6"
func (h *Server) InitUdpServer(network string, ip string, port uint16, maxMsgLen uint) error {
	h.readClientTimeOut = 60 * 60 //default 1 hour
	if h.completeDataCb == nil {
		h.GosofErr = errors.New("error : OnCompleteData not set")
		return h.GosofErr
	}
	if maxMsgLen == 0 {
		h.GosofErr = errors.New(fmt.Sprintf("error : invalid max msg len : %d", maxMsgLen))
		return h.GosofErr
	}
	connStr := fmt.Sprintf("%s:%d", ip, port)
	raddr, resolveErr := net.ResolveUDPAddr(network, connStr)
	if resolveErr != nil {
		return resolveErr
	}
	udpConn, netErr := net.ListenUDP(network, raddr)
	if netErr != nil {
		log.Println("InitServer error : ", netErr.Error())
		return netErr
	}
	if h.readClientTimeOut > 0 {
		h.GosofErr = udpConn.SetReadDeadline(time.Now().Add(time.Duration(maxReadTimeOutSecs) * time.Second))
		if h.GosofErr != nil {
			log.Println("SetReadDeadLine error : ", h.GosofErr.Error())
			return h.GosofErr
		}
	}
	//log.Println("udp server starts : ", udpConn.LocalAddr().String())
	if h.initCompletedCb != nil {
		h.initCompletedCb()
	}
	go func(conn *net.UDPConn) {
		defer func() {
			_ = conn.Close()
		}()
		recvBuf := make([]byte, maxMsgLen)
		for {
			recvedLen, clientAddress, err := conn.ReadFromUDP(recvBuf)
			if recvedLen > 0 {
				ctx := Context{UdpConn: conn, UdpAddr: clientAddress}
				h.completeDataCb(&ctx, recvBuf[:recvedLen], recvedLen)
			}
			if err != nil {
				if h.disConnectedCb != nil {
					ctx := Context{UdpConn: conn, UdpAddr: clientAddress}
					h.disConnectedCb(&ctx, err)
				}
				return
			}
		} // for
	}(udpConn)

	return nil
}

func (h *Client) InitUdpClient(network string, ip string, port uint16, maxMsgLen uint) error {
	connStr := fmt.Sprintf("%s:%d", ip, port)
	svrAddr, netErr := net.ResolveUDPAddr(network, connStr)
	if netErr != nil {
		return netErr
	}
	svrConn, connErr := net.DialUDP("udp", nil, svrAddr)
	h.Ctx.Conn = nil
	h.Ctx.UdpConn = svrConn
	h.Ctx.UdpAddr = svrAddr
	if connErr != nil {
		log.Println("InitClient error : ", connErr.Error())
		return connErr
	}
	//log.Println("InitClient : ", connStr, ", server :", svrAddr.String())
	if h.initCompletedCb != nil {
		h.initCompletedCb()
	}
	go func(conn *net.UDPConn) {
		defer func() {
			_ = conn.Close()
		}()
		recvBuf := make([]byte, maxMsgLen)
		for {
			recvedLen, _, err := conn.ReadFromUDP(recvBuf)
			if recvedLen > 0 {
				ctx := Context{UdpConn: conn}
				h.completeDataCb(&ctx, recvBuf[:recvedLen], recvedLen)
			}
			if h.disConnectedCb != nil {
				ctx := Context{UdpConn: conn}
				h.disConnectedCb(&ctx, err)
			}
		} // for
	}(svrConn)

	return nil
}

func (h *Client) SendToUdpServer(data []byte) error {
	_, writeErr := h.Ctx.UdpConn.Write(data)
	if writeErr != nil {
		log.Println(writeErr.Error())
		return writeErr
	}
	return nil
}
