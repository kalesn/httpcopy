package main

import (
	"expvar"
	"flag"
	"fmt"
	"httpcopy/pkg/httpreplay"
	"log"
	"net/http"
	httppptof "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to this file")
)

func init() {
	var defaultServeMux http.ServeMux
	http.DefaultServeMux = &defaultServeMux

	http.HandleFunc("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "{\n")
		first := true
		expvar.Do(func(kv expvar.KeyValue) {
			if kv.Key == "memstats" || kv.Key == "cmdline" {
				return
			}

			if !first {
				fmt.Fprintf(w, ",\n")
			}
			first = false
			fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
		})
		fmt.Fprintf(w, "\n}\n")
	})

	http.HandleFunc("/debug/pprof/", httppptof.Index)
	http.HandleFunc("/debug/pprof/cmdline", httppptof.Cmdline)
	http.HandleFunc("/debug/pprof/profile", httppptof.Profile)
	http.HandleFunc("/debug/pprof/symbol", httppptof.Symbol)
	http.HandleFunc("/debug/pprof/trace", httppptof.Trace)
}

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	}

	args := os.Args[1:]
	var plugins *httpreplay.InOutPlugins
	if len(args) > 0 && args[0] == "file-server" {
		if len(args) != 2 {
			log.Fatal("You should specify port and IP (optional) for the file server. Example: `gor file-server :80`")

			fmt.Println(0, "Started example file server for current directory on address ", args[1])
		} else {
			flag.Parse()
			httpreplay.CheckSettings()
			plugins = httpreplay.NewPlugins()

		}
	}
	log.Printf("[PPID %d and PID %d] \n", os.Getppid(), os.Getpid())

	if len(plugins.Inputs) == 0 || len(plugins.Outputs) == 0 {
		log.Fatal("Required at least 1 input and 1 output")
	}

	closeCh := make(chan int)
	emitter := httpreplay.NewEmitter()
	go emitter.Start(plugins)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	exit := 0
	select {
	case <-c:
		exit = 1
	case <-closeCh:
		exit = 0
	}
	emitter.Close()
	os.Exit(exit)
}
