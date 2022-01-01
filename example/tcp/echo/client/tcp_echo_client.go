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
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jeremyko/gosof"
)

func main() {
	var client gosof.Client
	client.SetInitCompletedCb(onInitCompleted)
	client.SetServerConnectedCb(onServerConnected)
	client.SetCalculateDataLenCb(onCalculateDataLen)
	client.SetCompleteDataCb(onCompleteData)
	client.SetDisConnectedCB(onServerDisConnected)

	if client.InitTcpClient("tcp", "127.0.0.1", 9990, 10) != nil {
		log.Println("error! ", client.GetLastErrMsg())
		return
	}
	consoleReader := bufio.NewReader(os.Stdin)
	for {
		input, _ := consoleReader.ReadString('\n')
		input = strings.Replace(input, "\n", "", 1)
		err := client.SendToServer(len(input), []byte(input))
		if err != nil {
			fmt.Println(err)
			return
		}
	} // for
}

//------------------------------------------------------------------------------
//user specific callbacks
//------------------------------------------------------------------------------

//successfully initialized.
func onInitCompleted() {
	log.Println("initialized")
}
func onServerConnected(ctx *gosof.Context) {
	log.Println("server connected : ", ctx.Conn.RemoteAddr().String())
}

func onServerDisConnected(ctx *gosof.Context, err error) {
	log.Println("server disconnected : ", ctx.Conn.RemoteAddr().String(), " - ", err.Error())
}

// calculate your complete packet length here using buffer data.
func onCalculateDataLen(data []byte, receivedLen int) (gosof.SocketOpFlag, int) {
	return gosof.AnalyzedCompleted, receivedLen // this is a simple echo example
}

// your whole data has arrived.
func onCompleteData(ctx *gosof.Context, data []byte, packetLen int) {
	log.Println("received  [", string(data), "] len =", packetLen)
}
