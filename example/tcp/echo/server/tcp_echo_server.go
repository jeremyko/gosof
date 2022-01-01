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
	"github.com/jeremyko/gosof"
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

// calculate your complete packet length here using buffer data.
func onCalculateDataLen(data []byte, receivedLen int) (gosof.SocketOpFlag, int) {
	return gosof.AnalyzedCompleted, receivedLen // this is a simple echo server
}

// your whole data has arrived.
func onCompleteData(ctx *gosof.Context, data []byte, packetLen int) {
	log.Println("client data [", string(data), "] len =", packetLen)
	err := svr.SendTcp(ctx, packetLen, data) // this is a simple echo server
	if err != nil {
		log.Println(err.Error())
	}
}
