package main

import (
	"expvar"
	"flag"
	"fmt"
	"net/http"
	httppptof "net/http/pprof"
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
