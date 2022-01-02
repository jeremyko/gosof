# Golang Socket Framework

### What
A simple and easy golang socket server/client framework especially convenient for handling TCP fixed-length header and variable-length body using pure net package. Of course, it also supports udp and domain socket.

- The framework calls the user-specified callback.
- For TCP, the total size of user data is passed to the framework via a callback, and the framework does TCP buffering automatically.
- Byte data is exchanged.
- Supports TCP, UDP and Domain Socket.

### Usage
```bash
go get github.com/jeremyko/gosof

```
### Example

See the example folder for all examples.
#### tcp echo server (Fixed-length header and variable-length body)
```go
package custom_msg

// user specific custom data example

// UserMsgHeader : fixed length header
type UserMsgHeader struct {
	MsgTotalLen uint32 // total length of data -> header + body
	MsgType     [6]byte
	EtcInfo     [20]byte
}

const FixedHeaderSize = 4 + 6 + 20

// UserMsgBody : body of dynamic length
type UserMsgBody struct {
	Field1 int
	Field2 string
	Field3 []byte
}
```

```go
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	custommsg "github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def"
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
```

#### tcp echo client (Fixed-length header and variable-length body)
```go
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	custommsg "github.com/jeremyko/gosof/example/tcp/echo_custom_msg/custom_msg_def"
	"github.com/jeremyko/gosof"
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

	for i := 0; i < 10; i++ {
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
```
