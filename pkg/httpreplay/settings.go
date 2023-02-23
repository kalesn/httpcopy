package httpreplay

import (
	"flag"
	"fmt"
	"httpcopy/pkg/output"
	"httpcopy/pkg/size"
	"os"
	"strconv"
	"sync"
	"time"
)

// DEMO indicates that goreplay is running in demo mode
var DEMO string

// MultiOption allows to specify multiple flags with same name and collects all values into array
type MultiOption struct {
	a *[]string
}

func (h *MultiOption) String() string {
	if h.a == nil {
		return ""
	}
	return fmt.Sprint(*h.a)
}

// Set gets called multiple times for each flag with same name
func (h *MultiOption) Set(value string) error {
	if h.a == nil {
		return nil
	}

	*h.a = append(*h.a, value)
	return nil
}

// MultiOption allows to specify multiple flags with same name and collects all values into array
type MultiIntOption struct {
	a *[]int
}

func (h *MultiIntOption) String() string {
	if h.a == nil {
		return ""
	}

	return fmt.Sprint(*h.a)
}

// Set gets called multiple times for each flag with same name
func (h *MultiIntOption) Set(value string) error {
	if h.a == nil {
		return nil
	}

	val, _ := strconv.Atoi(value)
	*h.a = append(*h.a, val)
	return nil
}

// AppSettings is the struct of main configuration
type AppSettings struct {
	Verbose   int           `json:"verbose"`
	Stats     bool          `json:"stats"`
	ExitAfter time.Duration `json:"exit-after"`

	SplitOutput          bool   `json:"split-output"`
	RecognizeTCPSessions bool   `json:"recognize-tcp-sessions"`
	Pprof                string `json:"http-pprof"`

	CopyBufferSize size.Size `json:"copy-buffer-size"`

	OutputStdout bool `json:"output-stdout"`
	OutputNull   bool `json:"output-null"`

	InputFile          []string      `json:"input-file"`
	InputFileLoop      bool          `json:"input-file-loop"`
	InputFileReadDepth int           `json:"input-file-read-depth"`
	InputFileDryRun    bool          `json:"input-file-dry-run"`
	InputFileMaxWait   time.Duration `json:"input-file-max-wait"`
	OutputFile         []string      `json:"output-file"`
	OutputFileConfig   output.FileOutputConfig

	InputHTTP    []string
	OutputHTTP   []string `json:"output-http"`
	PrettifyHTTP bool     `json:"prettify-http"`

	OutputHTTPConfig output.HTTPOutputConfig
}

// Settings holds Gor configuration
var Settings AppSettings

func usage() {
	fmt.Printf("Gor is a simple http traffic replication tool written in Go. Its main goal is to replay traffic from production servers to staging and dev environments.\nProject page: https://github.com/buger/gor\nAuthor: <Leonid Bugaev> leonsbox@gmail.com\nCurrent Version: v%s\n\n", "1.0")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.StringVar(&Settings.Pprof, "http-pprof", "", "Enable profiling. Starts  http server on specified port, exposing special /debug/pprof endpoint. Example: `:8181`")
	flag.IntVar(&Settings.Verbose, "verbose", 0, "set the level of verbosity, if greater than zero then it will turn on debug output")
	flag.BoolVar(&Settings.Stats, "stats", false, "Turn on queue stats output")

	if DEMO == "" {
		flag.DurationVar(&Settings.ExitAfter, "exit-after", 0, "exit after specified duration")
	} else {
		Settings.ExitAfter = 5 * time.Minute
	}

	flag.BoolVar(&Settings.SplitOutput, "split-output", false, "By default each output gets same traffic. If set to `true` it splits traffic equally among all outputs.")
	flag.BoolVar(&Settings.RecognizeTCPSessions, "recognize-tcp-sessions", false, "[PRO] If turned on http output will create separate worker for each TCP session. Splitting output will session based as well.")

	flag.BoolVar(&Settings.OutputStdout, "output-stdout", false, "Used for testing inputs. Just prints to console data coming from inputs.")
	flag.BoolVar(&Settings.OutputNull, "output-null", false, "Used for testing inputs. Drops all requests.")

	flag.Var(&MultiOption{&Settings.InputFile}, "input-file", "Read requests from file: \n\tgor --input-file ./requests.gor --output-http staging.com")
	flag.BoolVar(&Settings.InputFileLoop, "input-file-loop", false, "Loop input files, useful for performance testing.")
	flag.IntVar(&Settings.InputFileReadDepth, "input-file-read-depth", 100, "GoReplay tries to read and cache multiple records, in advance. In parallel it also perform sorting of requests, if they came out of order. Since it needs hold this buffer in memory, bigger values can cause worse performance")
	flag.BoolVar(&Settings.InputFileDryRun, "input-file-dry-run", false, "Simulate reading from the data source without replaying it. You will get information about expected replay time, number of found records etc.")
	flag.DurationVar(&Settings.InputFileMaxWait, "input-file-max-wait", 0, "Set the maximum time between requests. Can help in situations when you have too long periods between request, and you want to skip them. Example: --input-raw-max-wait 1s")

	flag.Var(&MultiOption{&Settings.OutputFile}, "output-file", "Write incoming requests to file: \n\tgor --input-raw :80 --output-file ./requests.gor")

	flag.BoolVar(&Settings.PrettifyHTTP, "prettify-http", false, "If enabled, will automatically decode requests and responses with: Content-Encoding: gzip and Transfer-Encoding: chunked. Useful for debugging, in conjunction with --output-stdout")

	flag.Var(&Settings.CopyBufferSize, "copy-buffer-size", "Set the buffer size for an individual request (default 5MB)")

	flag.Var(&MultiOption{&Settings.OutputHTTP}, "output-http", "Forwards incoming requests to given http address.\n\t# Redirect all incoming requests to staging.com address \n\tgor --input-raw :80 --output-http http://staging.com")

	// default values, using for tests
	Settings.OutputFileConfig.SizeLimit = 33554432
	Settings.OutputFileConfig.OutputFileMaxSize = 1099511627776
	Settings.CopyBufferSize = 5242880

}

func CheckSettings() {
	if Settings.OutputFileConfig.SizeLimit < 1 {
		Settings.OutputFileConfig.SizeLimit.Set("32mb")
	}
	if Settings.OutputFileConfig.OutputFileMaxSize < 1 {
		Settings.OutputFileConfig.OutputFileMaxSize.Set("1tb")
	}
	if Settings.CopyBufferSize < 1 {
		Settings.CopyBufferSize.Set("5mb")
	}
}

var previousDebugTime = time.Now()
var debugMutex sync.Mutex

// Debug take an effect only if --verbose greater than 0 is specified
func Debug(level int, args ...interface{}) {
	if Settings.Verbose >= level {
		debugMutex.Lock()
		defer debugMutex.Unlock()
		now := time.Now()
		diff := now.Sub(previousDebugTime)
		previousDebugTime = now
		fmt.Fprintf(os.Stderr, "[DEBUG][elapsed %s]: ", diff)
		fmt.Fprintln(os.Stderr, args...)
	}
}
