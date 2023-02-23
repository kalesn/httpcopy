package input

import (
	"bytes"
	"httpcopy/pkg/httpreplay"
	output2 "httpcopy/pkg/output"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPInput(t *testing.T) {
	//wg := new(sync.WaitGroup)

	input := NewHTTPInput("127.0.0.1:8090")
	time.Sleep(time.Millisecond)
	//output := NewTestOutput(func(*Message) {
	//	wg.Done()
	//})
	output := output2.NewFileOutput("/tmp/test_requests_0.gor", &output2.FileOutputConfig{FlushInterval: time.Minute, Append: true})

	plugins := &httpreplay.InOutPlugins{
		Inputs:  []httpreplay.PluginReader{input},
		Outputs: []httpreplay.PluginWriter{output},
	}
	plugins.All = append(plugins.All, input, output)

	emitter := httpreplay.NewEmitter()
	go emitter.Start(plugins)

	address := strings.Replace(input.address, "[::]", "127.0.0.1", -1)

	for i := 0; i < 100; i++ {
		//wg.Add(1)
		http.Get("http://" + address + "/")
	}

	//wg.Wait()
	emitter.Close()
}

func TestInputHTTPLargePayload(t *testing.T) {
	//wg := new(sync.WaitGroup)
	const n = 10 << 20 // 10MB
	var large [n]byte
	large[n-1] = '0'

	input := NewHTTPInput("127.0.0.1:0")
	//output := NewTestOutput(func(msg *Message) {
	//	_len := len(msg.Data)
	//	if _len >= n { // considering http body CRLF
	//		t.Errorf("expected body to be >= %d", n)
	//	}
	//	wg.Done()
	//})
	output := output2.NewFileOutput("/tmp/test_requests_0.gor", &output2.FileOutputConfig{FlushInterval: time.Minute, Append: true})
	plugins := &httpreplay.InOutPlugins{
		Inputs:  []httpreplay.PluginReader{input},
		Outputs: []httpreplay.PluginWriter{output},
	}
	plugins.All = append(plugins.All, input, output)

	emitter := httpreplay.NewEmitter()
	defer emitter.Close()
	go emitter.Start(plugins)

	address := strings.Replace(input.address, "[::]", "127.0.0.1", -1)
	var req *http.Request
	var err error
	req, err = http.NewRequest("POST", "http://"+address+"/abc", bytes.NewBuffer(large[:]))
	if err != nil {
		t.Error(err)
		return
	}
	//wg.Add(1)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	//wg.Wait()
}
