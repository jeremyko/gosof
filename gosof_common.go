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
	"bytes"
	"log"
	"net"
	"sync"
)

// What the server and the client use in common.

// network :  "ip", "ip4", "ip6", "unix", "unixgram", "unixpacket"
var maxReadTimeOutSecs = 60 * 60

type SocketOpFlag uint

const (
	NeedMoreInfo SocketOpFlag = 1 + iota
	AnalyzedCompleted
)

type Context struct {
	lock                sync.Mutex
	Conn                net.Conn
	IsDataLenCalculated bool
	TotalPacketLen      int
	UdpConn             *net.UDPConn
	UnixConn            *net.UnixConn
	UdpAddr             *net.UDPAddr
}

type Common struct {
	GosofErr            error
	maxDataByteLenLimit uint
	calculateDataLenCb  func(data []byte, receivedAccumulatedLen int) (SocketOpFlag, int)
	completeDataCb      func(ctx *Context, data []byte, packetLen int)
	disConnectedCb      func(ctx *Context, err error)
	initCompletedCb     func()
}

func (h *Common) GetLastErrMsg() string {
	if h.GosofErr != nil {
		return h.GosofErr.Error()
	}
	return ""
}

func (h *Common) tcpBufferWork(ctx *Context) {
	recvBuf := make([]byte, 4096)
	receivedTotalLen := 0
	var buffer bytes.Buffer

	defer func() {
		if ctx.Conn != nil {
			_ = ctx.Conn.Close()
		}
	}()
	for {
		readLen, readErr := ctx.Conn.Read(recvBuf)
		if nil != readErr {
			if h.disConnectedCb != nil {
				h.disConnectedCb(ctx, readErr)
			}
			return
		}
		receivedTotalLen += readLen
		//log.Println("read len =", readLen, " / receivedTotalLen= ", receivedTotalLen)
		buffer.Write(recvBuf[:readLen])

		var sockOp SocketOpFlag
		for {
			// Multiple data can be received in one chunk.
			if ctx.IsDataLenCalculated == false {
				// This callback is only called when the user does not know the packet information.
				sockOp, ctx.TotalPacketLen = h.calculateDataLenCb(buffer.Bytes(), receivedTotalLen)
				if sockOp == NeedMoreInfo {
					//log.Println("need more info -> wait")
					break // read again
				}
			}
			ctx.IsDataLenCalculated = true
			if receivedTotalLen >= ctx.TotalPacketLen {
				h.completeDataCb(ctx, buffer.Next(ctx.TotalPacketLen), ctx.TotalPacketLen)
				receivedTotalLen -= ctx.TotalPacketLen
				ctx.IsDataLenCalculated = false
				ctx.TotalPacketLen = 0
				if receivedTotalLen == 0 {
					break // All data processing complete.
				}
			} else {
				break
			}
		} // for
	} // for
}

func (h *Common) SendTcp(ctx *Context, totalLen int, datas ...[]byte) error {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()
	var totalSent = 0
	// Multiple goroutines may invoke methods on a Conn simultaneously.
	// --> However, this is not the case if multiple method calls are required.
	for _, data := range datas {
		dataLen := len(data)
		dataSent := 0
		dataOffset := 0
		var writeErr error
		for {
			dataSent, writeErr = ctx.Conn.Write(data[dataOffset:])
			if writeErr != nil {
				log.Println(writeErr.Error())
				return writeErr
			}
			totalSent += dataSent
			dataOffset += dataSent
			if totalSent == totalLen {
				//log.Println("sent OK : len : ", totalSent, " / ", len)
				return nil
			}
			if dataLen == dataOffset {
				break
			}
		} // for
	} // for
	return nil
}

func (h *Common) SendToClientUDP(ctx *Context, data []byte) error {
	// udp server --> client
	_, writeErr := ctx.UdpConn.WriteToUDP(data, ctx.UdpAddr)
	if writeErr != nil {
		log.Println(writeErr.Error())
		return writeErr
	}
	return nil
}

func (h *Common) SendUnix(ctx *Context, data []byte) error {
	_, writeErr := ctx.UnixConn.Write(data)
	if writeErr != nil {
		log.Println(writeErr.Error())
		return writeErr
	}
	return nil
}

func (h *Common) SetInitCompletedCb(cb func()) {
	h.initCompletedCb = cb
}

func (h *Common) SetCalculateDataLenCb(cb func(data []byte, receivedLen int) (SocketOpFlag, int)) {
	h.calculateDataLenCb = cb
}

func (h *Common) SetCompleteDataCb(cb func(ctx *Context, data []byte, packetLen int)) {
	h.completeDataCb = cb
}

func (h *Common) SetMaxDataByteLenLimit(maxByteLenLimit uint) {
	h.maxDataByteLenLimit = maxByteLenLimit
}

func (h *Common) SetDisConnectedCB(cb func(ctx *Context, err error)) {
	h.disConnectedCb = cb
}
