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
 ******************************************************************************/

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/jeremyko/gosof"
	custommsg "github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def"
	"log"
	"os"
)

var svr gosof.Server

func main() {
	svr.SetInitCompletedCb(onInitCompleted)
	svr.SetNewClientCb(onNewClient)
	svr.SetCalculateDataLenCb(onCalculateDataLen)
	svr.SetCompleteDataCb(onCompleteData)
	svr.SetDisConnectedCB(onClientDisconnected)
	// ListenConfig configuration is very platform specific.
	// For example, SO_REUSEPORT does not exist on Windows.
	// --> use parameter.
	//lc := net.ListenConfig{
	//	Control: func(network, address string, conn syscall.RawConn) error {
	//		var operr error
	//		if err := conn.Control(func(fd uintptr) {
	//			operr = syscall.SetsockoptInt(syscall.Handle(int(fd)),
	//				syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1) //for windows !!
	//		}); err != nil {
	//			return err
	//		}
	//		return operr
	//	},
	//}
	//if svr.InitTcpServerListenConfig("tcp4", "127.0.0.1", 9990, &lc) != nil {
	if svr.InitTcpServer("tcp4", "127.0.0.1", 9990) != nil {
		log.Println("error! ", svr.GetLastErrMsg())
		return
	}
	log.Println("press enter to exit")
	input := bufio.NewScanner(os.Stdin) // wait user input to terminate
	input.Scan()
}

//------------------------------------------------------------------------------
//user specific callbacks
//------------------------------------------------------------------------------
func onInitCompleted() {
	log.Println("initialized") //successfully initialized.
}

func onNewClient(ctx *gosof.Context) {
	log.Println("new connection : ", ctx.Conn.RemoteAddr().String())
}

func onClientDisconnected(ctx *gosof.Context, err error) {
	log.Println("client disconnected : ", ctx.Conn.RemoteAddr().String(), " - ", err.Error())
}

// calculate your complete packet length here
// Return 'gosof.NeedMoreInfo' if you haven't received enough information about your data.
func onCalculateDataLen(data []byte, receivedAccumulatedLen int) (gosof.SocketOpFlag, int) {
	if receivedAccumulatedLen < custommsg.FixedHeaderSize {
		// Fixed-length header data has not arrived yet.
		return gosof.NeedMoreInfo, 0
	}
	// get fixed length header
	myMsgHeader := custommsg.UserMsgHeader{}
	binBufHeader := bytes.Buffer{}
	binBufHeader.Write(data[:receivedAccumulatedLen])
	err := binary.Read(&binBufHeader, binary.LittleEndian, &myMsgHeader)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	return gosof.AnalyzedCompleted, int(myMsgHeader.MsgTotalLen)
}

// your whole data has arrived.
// 'data' contains header and actual data.
func onCompleteData(ctx *gosof.Context, data []byte, packetLen int) {
	// header
	myMsgHeader := custommsg.UserMsgHeader{}
	binBufHeader := bytes.Buffer{}
	binBufHeader.Write(data[:custommsg.FixedHeaderSize])
	err := binary.Read(&binBufHeader, binary.LittleEndian, &myMsgHeader)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	// body
	clientMsg := custommsg.UserMsgBody{}
	bufDataOnly := bytes.NewBuffer(data[custommsg.FixedHeaderSize:]) //except header data
	dec := gob.NewDecoder(bufDataOnly)                               // body of dynamic length
	err = dec.Decode(&clientMsg)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	log.Println(ctx.Conn.RemoteAddr().String(), " / client data : type = ", string(myMsgHeader.MsgType[:]),
		" etc info =", string(myMsgHeader.EtcInfo[:]),
		" / data :  ==> ", clientMsg.Field1, ", ", clientMsg.Field2, ", ", string(clientMsg.Field3))

	// this is a simple echo server
	err = svr.SendTcp(ctx, packetLen, data)
	if err != nil {
		log.Println(err.Error())
	}
}
