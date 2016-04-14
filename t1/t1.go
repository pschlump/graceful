package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pschlump/graceful"
)

/*
1. How to call "stop"
2. NoSignalHandling bool
*/

var g_data = "Hello World 1"
var g_Lock sync.Mutex
var n_cnt int
var p1 string
var p2 string
var pc string

func main() {
	p1 := ":3005"
	p2 := ":3004"
	pc := ":9944"
	var wg sync.WaitGroup
	n_cnt = 12
	var gs []*graceful.Server

	incN := func() {
		g_Lock.Lock()
		g_data = fmt.Sprintf("Hello World %d", n_cnt)
		n_cnt += 2
		g_Lock.Unlock()
	}

	m1 := func() {
		// wg.Add(1)
		mux1 := http.NewServeMux()
		fmt.Printf("Launching server on %s\n", p1)
		mux1.HandleFunc("/", func(www http.ResponseWriter, req *http.Request) {
			g_Lock.Lock()
			fmt.Fprintf(www, "From %s, %s\n", p1, g_data)
			g_Lock.Unlock()
		})
		gs = append(gs, graceful.Run2(p1, 0, mux1)) // xyzzy - should lock
		fmt.Printf("Terminated server on %s\n", p1)
		// wg.Done()
	}

	m2 := func() {
		// wg.Add(1)
		mux2 := http.NewServeMux()
		fmt.Printf("Launching server on %s\n", p2)
		mux2.HandleFunc("/", func(www http.ResponseWriter, req *http.Request) {
			g_Lock.Lock()
			fmt.Fprintf(www, "From %s, %s\n", p2, g_data)
			g_Lock.Unlock()
		})
		gs = append(gs, graceful.Run2(p2, 0, mux2)) // xyzzy - should lock
		fmt.Printf("Terminated server on %s\n", p2)
		// wg.Done()
	}

	fmt.Printf("About to run go-routines\n")
	go m1()
	go m2()

	wg.Add(1)
	go func() {
		mux3 := http.NewServeMux()
		fmt.Printf("Launching server on %s - will start now\n", pc)
		mux3.HandleFunc("/api/status", func(www http.ResponseWriter, req *http.Request) {
			incN()
			fmt.Fprintf(www, `{"status":"success","msg":"global data incremented","data":%q}`, g_data)
		})
		mux3.HandleFunc("/api/restart", func(www http.ResponseWriter, req *http.Request) {
			fmt.Printf("Increment of global data\n")
			incN()
			// Graceful shutdown of each of the servers
			for i, v := range gs {
				fmt.Printf("Shutdown of server %d\n", i)
				v.Stop(2 * time.Second)
			}
			// restart of the servers
			fmt.Printf("Statup of servers\n")
			gs = make([]*graceful.Server, 0, 10)
			go m1()
			go m2()
			fmt.Fprintf(www, `{"status":"restart-complete"}`)
		})
		http.ListenAndServe(pc, mux3)
		fmt.Printf("Terminated server on %s - will exit now\n", pc)
		wg.Done()
	}()

	fmt.Println("Press ctrl+c to kill, use /api/restart to restart client servers\n")
	wg.Wait()

}
