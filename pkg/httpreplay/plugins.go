package httpreplay

import (
	"reflect"
	"strings"
)

// Message represents data across plugins
type Message struct {
	Meta []byte // metadata
	Data []byte // actual data
}

// PluginReader is an interface for input plugins
type PluginReader interface {
	PluginRead() (msg *Message, err error)
}

// PluginWriter is an interface for output plugins
type PluginWriter interface {
	PluginWrite(msg *Message) (n int, err error)
}

// PluginReadWriter is an interface for plugins that support reading and writing
type PluginReadWriter interface {
	PluginReader
	PluginWriter
}

// InOutPlugins struct for holding references to plugins
type InOutPlugins struct {
	Inputs  []PluginReader
	Outputs []PluginWriter
	All     []interface{}
}

// extractLimitOptions detects if plugin get called with limiter support
// Returns address and limit
func extractLimitOptions(options string) (string, string) {
	split := strings.Split(options, "|")

	if len(split) > 1 {
		return split[0], split[1]
	}

	return split[0], ""
}

// NewPlugins specify and initialize all available plugins
func NewPlugins() *InOutPlugins {
	plugins := new(InOutPlugins)

	for _, options := range Settings.InputFile {
		plugins.registerPlugin(NewFileInput, options, Settings.InputFileLoop, Settings.InputFileReadDepth, Settings.InputFileMaxWait, Settings.InputFileDryRun)
	}

	for _, path := range Settings.OutputFile {
		plugins.registerPlugin(NewFileOutput, path, &Settings.OutputFileConfig)

	}

	for _, options := range Settings.InputHTTP {
		plugins.registerPlugin(NewHTTPInput, options)
	}

	for _, options := range Settings.OutputHTTP {
		plugins.registerPlugin(NewHTTPOutput, options, &Settings.OutputHTTPConfig)
	}

	return plugins
}

// Automatically detects type of plugin and initialize it
//
// See this article if curious about reflect stuff below: http://blog.burntsushi.net/type-parametric-functions-golang
func (plugins *InOutPlugins) registerPlugin(constructor interface{}, options ...interface{}) {
	var path, limit string
	vc := reflect.ValueOf(constructor)

	// Pre-processing options to make it work with reflect
	vo := []reflect.Value{}
	for _, oi := range options {
		vo = append(vo, reflect.ValueOf(oi))
	}

	if len(vo) > 0 {
		// Removing limit options from path
		path, limit = extractLimitOptions(vo[0].String())

		// Writing value back without limiter "|" options
		vo[0] = reflect.ValueOf(path)
	}

	// Calling our constructor with list of given options
	plugin := vc.Call(vo)[0].Interface()

	if limit != "" {
		plugin = NewLimiter(plugin, limit)
	}

	// Some of the output can be Readers as well because return responses
	if r, ok := plugin.(PluginReader); ok {
		plugins.Inputs = append(plugins.Inputs, r)
	}

	if w, ok := plugin.(PluginWriter); ok {
		plugins.Outputs = append(plugins.Outputs, w)
	}
	plugins.All = append(plugins.All, plugin)
}
