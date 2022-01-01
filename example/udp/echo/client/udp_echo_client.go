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
	client.SetCompleteDataCb(onCompleteData) // set user callbacks
	client.SetInitCompletedCb(onInitCompleted)
	if client.InitUdpClient("udp", "127.0.0.1", 9990, 10240) != nil {
		log.Println("error! ", client.GetLastErrMsg())
		return
	}
	consoleReader := bufio.NewReader(os.Stdin)
	for {
		input, _ := consoleReader.ReadString('\n')
		input = strings.Replace(input, "\n", "", 1)
		err := client.SendToUdpServer([]byte(input))
		if err != nil {
			fmt.Println(err)
			return
		}
	} // for
}

//------------------------------------------------------------------------------
// user specific callbacks
//------------------------------------------------------------------------------
func onInitCompleted() {
	log.Println("initialized") // successfully initialized.
}
func onCompleteData(ctx *gosof.Context, data []byte, packetLen int) {
	// your whole data has arrived.
	log.Println("received  [", string(data), "] len =", packetLen)
}
