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
	"os"
	"time"
)

// InitUnixServer
// network :  "unix", "unixgram", "unixpacket"
func (h *Server) InitUnixServer(network string, address string, maxMsgLen uint) error {
	raddr, resolveErr := net.ResolveUnixAddr(network, address)
	if resolveErr != nil {
		h.GosofErr = errors.New("error : invalid network : " + network)
		return resolveErr
	}
	if h.readClientTimeOut == 0 {
		h.readClientTimeOut = 60 * 60 //default 1 hour
	}
	if h.completeDataCb == nil {
		h.GosofErr = errors.New("error : OnCompleteData not set")
		return h.GosofErr
	}
	if maxMsgLen == 0 {
		h.GosofErr = errors.New(fmt.Sprintf("error : invalid max msg len : %d", maxMsgLen))
		return h.GosofErr
	}
	if _, h.GosofErr = os.Stat(address); h.GosofErr == nil {
		h.GosofErr = os.Remove(address)
		if h.GosofErr != nil {
			return h.GosofErr
		}
	}

	h.unixListener, h.GosofErr = net.ListenUnix(network, raddr)
	if h.GosofErr != nil {
		log.Println("InitServer error : ", h.GosofErr.Error())
		return h.GosofErr
	}
	if h.initCompletedCb != nil {
		h.initCompletedCb()
	}
	go func() {
		defer func() {
			_ = h.unixListener.Close()
			log.Println("listener closed")
		}()
		for {
			conn, err := h.unixListener.AcceptUnix()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				log.Fatal(err)
				return
			}
			ctx := Context{UnixConn: conn}
			if h.newClientCb != nil {
				h.newClientCb(&ctx)
			}
			go func(clientCtx *Context) {
				if h.readClientTimeOut > 0 {
					deadLineErr := clientCtx.UnixConn.SetReadDeadline(time.Now().Add(time.Duration(maxReadTimeOutSecs) * time.Second))
					if deadLineErr != nil {
						log.Println("SetReadDeadLine error : ", deadLineErr.Error())
						return
					}
				}
				recvBuf := make([]byte, maxMsgLen)
				defer func() {
					if clientCtx.UnixConn != nil {
						_ = clientCtx.UnixConn.Close()
					}
				}()
				for {
					recvedLen, _, readErr := clientCtx.UnixConn.ReadFromUnix(recvBuf)
					if recvedLen > 0 {
						h.completeDataCb(clientCtx, recvBuf[:recvedLen], recvedLen)
					}
					if nil != readErr {
						if h.disConnectedCb != nil {
							h.disConnectedCb(clientCtx, readErr)
						}
						return
					}
				} // for
			}(&ctx)
		} // for
	}()

	return nil
}

// InitUnixClient
// network :  "unix", "unixgram", "unixpacket"
func (h *Client) InitUnixClient(network string, svrAddr string, cliAddr string, maxMsgLen uint) error {
	var connErr error
	var svrConn *net.UnixConn
	raddr, resolveErr := net.ResolveUnixAddr(network, svrAddr)
	if resolveErr != nil {
		h.GosofErr = errors.New("error : invalid network : " + network)
		return resolveErr
	}
	laddr := net.UnixAddr{Name: cliAddr, Net: network}
	svrConn, connErr = net.DialUnix(network, &laddr, raddr)
	h.Ctx.UnixConn = svrConn
	if connErr != nil {
		log.Println("InitUnixClient error : ", connErr.Error())
		return connErr
	}
	if h.initCompletedCb != nil {
		h.initCompletedCb()
	}
	go func(conn *net.UnixConn, cliSockFile string) {
		defer func() {
			_ = conn.Close()
			_ = os.Remove(cliSockFile)
		}()
		recvBuf := make([]byte, maxMsgLen)
		for {
			recvedLen, _, readErr := conn.ReadFromUnix(recvBuf)
			if recvedLen > 0 {
				ctx := Context{UnixConn: conn}
				h.completeDataCb(&ctx, recvBuf[:recvedLen], recvedLen)
			}
			if nil != readErr {
				if h.disConnectedCb != nil {
					ctx := Context{UnixConn: conn}
					h.disConnectedCb(&ctx, readErr)
				}
				return
			}
		} // for
	}(svrConn, cliAddr)

	return nil
}

func (h *Client) SendToUnixServer(data []byte) error {
	_, writeErr := h.Ctx.UnixConn.Write(data)
	if writeErr != nil {
		log.Println(writeErr.Error())
		return writeErr
	}
	return nil
}
