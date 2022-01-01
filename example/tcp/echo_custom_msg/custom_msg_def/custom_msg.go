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

package custom_msg

// user specific custom data example

// UserMsgHeader : fixed length header
type UserMsgHeader struct {
	MsgTotalLen uint32 // total length of data -> header + body
	MsgType     [6]byte
	EtcInfo     [20]byte
}

// FixedHeaderSize :
// padding을 고려하여 길이를 계산하는 unsafe.Sizeof 메서드는 사용할 수 없다.
// 고정 길이 헤더를 위해서는 binary.Write를 사용해야하고, 이 방식에서는 데이터 패킹이 발생되기 때문이다.
// The unsafe.Sizeof method that calculates the length considering padding cannot be used.
// This is because binary.Write must be used for fixed-length headers, and data packing occurs in this method.
const FixedHeaderSize = 4 + 6 + 20

// UserMsgBody : body of dynamic length
type UserMsgBody struct {
	Field1 int
	Field2 string
	Field3 []byte
}
