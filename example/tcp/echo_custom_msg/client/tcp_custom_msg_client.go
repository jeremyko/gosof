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
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/jeremyko/gosof"
	custommsg "github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def"
	"log"
	"time"
)

var recvedCnt uint = 0

func main() {
	var client gosof.Client
	client.SetInitCompletedCb(onInitCompleted)
	client.SetServerConnectedCb(onServerConnected)
	client.SetCalculateDataLenCb(onCalculateDataLen)
	client.SetCompleteDataCb(onCompleteData)
	client.SetDisConnectedCB(onServerDisConnected)
	log.SetFlags(log.Llongfile)

	if client.InitTcpClient("tcp", "127.0.0.1", 9990, 10) != nil {
		log.Println("error! ", client.GetLastErrMsg())
		return
	}

	for i := 0; i < 1000; i++ {
		go func(index int) {
			var binBufHeader bytes.Buffer
			var binBufBody bytes.Buffer

			// fixed length header
			myMsgHeader := custommsg.UserMsgHeader{}
			copy(myMsgHeader.MsgType[:], "type1")
			copy(myMsgHeader.EtcInfo[:], "some useful info")

			// body of dynamic length
			myMsg := custommsg.UserMsgBody{}
			myMsg.Field1 = index
			myMsg.Field2 = fmt.Sprintf("TEST %d", index)
			myMsg.Field3 = []byte("byte test")
			encBody := gob.NewEncoder(&binBufBody)
			err := encBody.Encode(myMsg)
			if err != nil {
				log.Fatal("encode error:", err)
			}
			bodyLen := binBufBody.Len()
			myMsgHeader.MsgTotalLen = uint32(custommsg.FixedHeaderSize + bodyLen)
			binary.Write(&binBufHeader, binary.LittleEndian, &myMsgHeader)
			//--------------------------------- send header and real data
			errSend := client.SendToServer(int(myMsgHeader.MsgTotalLen),
				binBufHeader.Bytes(), binBufBody.Bytes())
			if errSend != nil {
				log.Println(errSend.Error())
				return
			}
		}(i)
	} // for

	for {
		time.Sleep(1 * time.Second)
	}
}

//------------------------------------------------------------------------------
//user specific callbacks
//------------------------------------------------------------------------------
func onInitCompleted() {
	log.Println("initialized") //successfully initialized.
}

func onServerConnected(ctx *gosof.Context) {
	log.Println("server connected : ", ctx.Conn.RemoteAddr().String())
}

func onServerDisConnected(ctx *gosof.Context, err error) {
	log.Println("server disconnected : ", ctx.Conn.RemoteAddr().String(), " - ", err.Error())
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
	binBufHeader.Write(data[:custommsg.FixedHeaderSize])
	err := binary.Read(&binBufHeader, binary.LittleEndian, &myMsgHeader)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	return gosof.AnalyzedCompleted, int(myMsgHeader.MsgTotalLen)
}

// your whole data has arrived. The echo response sent by the server has arrived.
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
	myMsg := custommsg.UserMsgBody{}
	bufDataOnly := bytes.NewBuffer(data[custommsg.FixedHeaderSize:]) //except header data
	dec := gob.NewDecoder(bufDataOnly)                               // body of dynamic length
	err = dec.Decode(&myMsg)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	recvedCnt++
	log.Println(" received count :", recvedCnt, " : len [", packetLen, "] ==> ",
		myMsg.Field1, ", ", myMsg.Field2, ", ", string(myMsg.Field3))
}
