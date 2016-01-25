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

func main() {

	var wg sync.WaitGroup
	n_cnt = 12

	incN := func() {
		g_Lock.Lock()
		g_data = fmt.Sprintf("Hello World %d", n_cnt)
		n_cnt += 2
		g_Lock.Unlock()
	}

	var gs []*graceful.Server

	m1 := func() {
		wg.Add(1)
		mux1 := http.NewServeMux()
		fmt.Println("Launching server on :3000")
		mux1.HandleFunc("/", func(www http.ResponseWriter, req *http.Request) {
			g_Lock.Lock()
			fmt.Fprintf(www, "From 3000, %s\n", g_data)
			g_Lock.Unlock()
		})
		gs = append(gs, graceful.Run2(":3000", 0, mux1))
		fmt.Println("Terminated server on :3000")
		wg.Done()
	}
	m2 := func() {
		wg.Add(1)
		mux2 := http.NewServeMux()
		fmt.Println("Launching server on :3001")
		mux2.HandleFunc("/", func(www http.ResponseWriter, req *http.Request) {
			g_Lock.Lock()
			fmt.Fprintf(www, "From 3001, %s\n", g_data)
			g_Lock.Unlock()
		})
		gs = append(gs, graceful.Run2(":3001", 0, mux2))
		fmt.Println("Terminated server on :3001")
		wg.Done()
	}

	go m1()
	go m2()

	go func() {
		wg.Add(1)
		mux3 := http.NewServeMux()
		fmt.Println("Launching server on :9940 - will exit now")
		mux3.HandleFunc("/api/t1_setg", func(www http.ResponseWriter, req *http.Request) {
			incN()
			fmt.Fprintf(www, `{"status":"success"}`)
		})
		mux3.HandleFunc("/api/restart", func(www http.ResponseWriter, req *http.Request) {
			incN()
			// graceful.performShutdown() // quitting chan struct{}, listener net.Listener)
			for _, v := range gs {
				v.Stop(2 * time.Second)
			}
			// go client()
			gs = make([]*graceful.Server, 0, 10)
			go m1()
			go m2()
			fmt.Fprintf(www, `{"status":"success"}`)
		})
		http.ListenAndServe(":9940", mux3)
		fmt.Println("Terminated server on :9940 - will exit now")
		wg.Done()
	}()

	fmt.Println("Press ctrl+c to kill, use /api/restart to restart client servers\n")
	wg.Wait()

}

/*
func main() {

	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		n := negroni.New()
		fmt.Println("Launching server on :3000")
		graceful.Run(":3000", 0, n)
		fmt.Println("Terminated server on :3000")
		wg.Done()
	}()
	go func() {
		n := negroni.New()
		fmt.Println("Launching server on :3001")
		graceful.Run(":3001", 0, n)
		fmt.Println("Terminated server on :3001")
		wg.Done()
	}()
	go func() {
		n := negroni.New()
		fmt.Println("Launching server on :3002")
		graceful.Run(":3002", 0, n)
		fmt.Println("Terminated server on :3002")
		wg.Done()
	}()
	fmt.Println("Press ctrl+c. All servers should terminate.")
	wg.Wait()

}
*/

/*
func main() {
	var wg sync.WaitGroup

	var srv *graceful.Server

	client := func() {
		wg.Add(1)
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(www http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(www, g_data)
		})

		// srv = graceful.ListenAndServe(":3001", 10*time.Second, mux)
		srv := &graceful.Server{
			Timeout: 10 * time.Second,
			Server:  &http.Server{Addr: ":3001", Handler: mux},
		}
		srv.ListenAndServe()
	}

	go func() {
		wg.Add(1)

		n_cnt = 10

		go client()

		mux := http.NewServeMux()

		mux.HandleFunc("/api/test2_setg", func(www http.ResponseWriter, req *http.Request) {
			g_data = fmt.Sprintf("Hello World %d", n_cnt)
			n_cnt += 2
			fmt.Fprintf(www, `{"status":"success"}`)
		})
		mux.HandleFunc("/api/restart", func(www http.ResponseWriter, req *http.Request) {
			g_Lock.Lock()
			g_data = fmt.Sprintf("Hello World %d", n_cnt)
			g_Lock.Unlock()
			n_cnt += 2

			graceful.performShutdown() // quitting chan struct{}, listener net.Listener)

			go client()
			fmt.Fprintf(www, `{"status":"success"}`)
		})

		http.ListenAndServe(":9940", 10*time.Second, mux)
	}()

	fmt.Printf("Listeneing on 3001, client and 9940 control\n")
	wg.Wait()
}
*/
