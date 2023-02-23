package httpreplay

import (
	"encoding/base64"
	"errors"
	"math/rand"
	"time"
)

// ErrorStopped is the error returned when the go routines reading the input is stopped.
var ErrorStopped = errors.New("reading stopped")

// TestInput used for testing purpose, it allows emitting requests on demand
type TestInput struct {
	data       chan []byte
	skipHeader bool
	stop       chan bool // Channel used only to indicate goroutine should shutdown
}

// NewTestInput constructor for TestInput
func NewTestInput() (i *TestInput) {
	i = new(TestInput)
	i.data = make(chan []byte, 100)
	i.stop = make(chan bool)
	return
}

// PluginRead reads message from this plugin
func (i *TestInput) PluginRead() (*Message, error) {
	var msg Message
	select {
	case buf := <-i.data:
		msg.Data = buf
		if !i.skipHeader {
			msg.Meta = PayloadHeader(RequestPayload, Uuid(), time.Now().UnixNano(), -1)
		} else {
			msg.Meta, msg.Data = PayloadMetaWithBody(msg.Data)
		}

		return &msg, nil
	case <-i.stop:
		return nil, ErrorStopped
	}
}

// Close closes this plugin
func (i *TestInput) Close() error {
	close(i.stop)
	return nil
}

// EmitBytes sends data
func (i *TestInput) EmitBytes(data []byte) {
	i.data <- data
}

// EmitGET emits GET request without headers
func (i *TestInput) EmitGET() {
	i.data <- []byte("GET / HTTP/1.1\r\n\r\n")
}

// EmitPOST emits POST request with Content-Length
func (i *TestInput) EmitPOST() {
	i.data <- []byte("POST /pub/WWW/ HTTP/1.1\r\nContent-Length: 7\r\nHost: www.w3.org\r\n\r\na=1&b=2")
}

// EmitChunkedPOST emits POST request with `Transfer-Encoding: chunked` and chunked body
func (i *TestInput) EmitChunkedPOST() {
	i.data <- []byte("POST /pub/WWW/ HTTP/1.1\r\nHost: www.w3.org\r\nTransfer-Encoding: chunked\r\n\r\n4\r\nWiki\r\n5\r\npedia\r\ne\r\n in\r\n\r\nchunks.\r\n0\r\n\r\n")
}

// EmitLargePOST emits POST request with large payload (5mb)
func (i *TestInput) EmitLargePOST() {
	size := 5 * 1024 * 1024 // 5 MB
	rb := make([]byte, size)
	rand.Read(rb)

	rs := base64.URLEncoding.EncodeToString(rb)

	i.data <- []byte("POST / HTTP/1.1\r\nHost: www.w3.org\nContent-Length:5242880\r\n\r\n" + rs)
}

// EmitSizedPOST emit a POST with a payload set to a supplied size
func (i *TestInput) EmitSizedPOST(payloadSize int) {
	rb := make([]byte, payloadSize)
	rand.Read(rb)

	rs := base64.URLEncoding.EncodeToString(rb)

	i.data <- []byte("POST / HTTP/1.1\r\nHost: www.w3.org\nContent-Length:5242880\r\n\r\n" + rs)
}

// EmitOPTIONS emits OPTIONS request, similar to GET
func (i *TestInput) EmitOPTIONS() {
	i.data <- []byte("OPTIONS / HTTP/1.1\r\nHost: www.w3.org\r\n\r\n")
}

func (i *TestInput) String() string {
	return "Test Input"
}
