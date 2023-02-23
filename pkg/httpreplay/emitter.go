package httpreplay

import (
	"fmt"
	"io"
	"sync"
)

// Emitter represents an abject to manage plugins communication
type Emitter struct {
	sync.WaitGroup
	plugins *InOutPlugins
}

// NewEmitter creates and initializes new Emitter object.
func NewEmitter() *Emitter {
	return &Emitter{}
}

// Start initialize loop for sending data from inputs to outputs
func (e *Emitter) Start(plugins *InOutPlugins) {
	e.plugins = plugins

	for _, in := range plugins.Inputs {
		e.Add(1)
		go func(in PluginReader) {
			defer e.Done()
			if err := CopyMulty(in, plugins.Outputs...); err != nil {
				fmt.Println(2, fmt.Sprintf("[EMITTER] error during copy: %q", err))
			}
		}(in)
	}
}

// Close closes all the goroutine and waits for it to finish.
func (e *Emitter) Close() {
	for _, p := range e.plugins.All {
		if cp, ok := p.(io.Closer); ok {
			cp.Close()
		}
	}
	if len(e.plugins.All) > 0 {
		// wait for everything to stop
		e.Wait()
	}
	e.plugins.All = nil // avoid Close to make changes again
}

// CopyMulty copies from 1 reader to multiple writers
func CopyMulty(src PluginReader, writers ...PluginWriter) error {
	//wIndex := 0
	for {
		msg, err := src.PluginRead()
		if err != nil {
			if err == ErrorStopped || err == io.EOF {
				return nil
			}
			return err
		}
		if msg != nil && len(msg.Data) > 0 {
			//if len(msg.Data) > int(Settings.CopyBufferSize) {
			//	msg.Data = msg.Data[:Settings.CopyBufferSize]
			//}
			meta := PayloadMeta(msg.Meta)
			if len(meta) < 3 {
				fmt.Println(2, fmt.Sprintf("[EMITTER] Found malformed record %q from %q", msg.Meta, src))
				continue
			}

			for _, dst := range writers {
				if _, err := dst.PluginWrite(msg); err != nil && err != io.ErrClosedPipe {
					return err
				}
			}
		}
	}
}
