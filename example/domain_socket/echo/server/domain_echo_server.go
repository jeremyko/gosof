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
	"log"
	"os"

	"github.com/jeremyko/gosof"
)

var svr gosof.Server

func main() {
	svr.SetInitCompletedCb(onInitCompleted)
	svr.SetNewClientCb(onNewClient)
	svr.SetCompleteDataCb(onCompleteData)
	svr.SetDisConnectedCB(onClientDisconnected)
	if svr.InitUnixServer("unix", "/tmp/sosof_test_server.sock", 10240) != nil {
		log.Println("error! ", svr.GetLastErrMsg())
		return
	}
	input := bufio.NewScanner(os.Stdin) // wait user input to terminate
	input.Scan()
}

//------------------------------------------------------------------------------
//user specific callbacks
//------------------------------------------------------------------------------
func onInitCompleted() {
	log.Println("initialized") //successfully initialized.
	log.Println("press enter to exit")
}

func onNewClient(ctx *gosof.Context) {
	log.Println("new connection : ", ctx.UnixConn.RemoteAddr().String())
}

func onClientDisconnected(ctx *gosof.Context, err error) {
	log.Println("client disconnected : ", ctx.UnixConn.RemoteAddr().String(), " - ", err.Error())
}

// your whole data has arrived.
func onCompleteData(ctx *gosof.Context, data []byte, packetLen int) {
	log.Println(ctx.UnixConn.RemoteAddr().String(), "  - client data : ", string(data), ", len =", packetLen)
	err := svr.SendUnix(ctx, data) // this is a simple echo server
	if err != nil {
		log.Println(err.Error())
	}
}
