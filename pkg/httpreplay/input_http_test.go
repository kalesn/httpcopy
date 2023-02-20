package httpreplay

import (
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHTTPInput(t *testing.T) {
	wg := new(sync.WaitGroup)

	input := NewHTTPInput("127.0.0.1:0")
	time.Sleep(time.Millisecond)
	output := NewTestOutput(func(*Message) {
		wg.Done()
	})

	plugins := &InOutPlugins{
		Inputs:  []PluginReader{input},
		Outputs: []PluginWriter{output},
	}
	plugins.All = append(plugins.All, input, output)

	emitter := NewEmitter()
	go emitter.Start(plugins)

	address := strings.Replace(input.address, "[::]", "127.0.0.1", -1)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		http.Get("http://" + address + "/")
	}

	wg.Wait()
	emitter.Close()
}
