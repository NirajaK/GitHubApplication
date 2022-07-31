package main

import (
	"../github"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)


func main() {

	subrouter := deployment.AddHandlers()


	svcport := "10000"
	addr := fmt.Sprintf(":%s", svcport)

	srv := &http.Server{Addr: addr, Handler: subrouter}
	//log.Fatal(srv.ListenAndServe())

	errs := make(chan error)
	exit := make(chan bool, 1)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		err := <-c
		if err != nil {
			fmt.Println("Interrupt Signal.. Terminating application")
			errs <- fmt.Errorf("%s",err)
			exit <- true
		}
	}()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			errs <- fmt.Errorf("%s",err)
			exit <- true
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			select {
			case <-exit:
				fmt.Println("Exit Channel set. Closing server")
			case <-errs:
				fmt.Println("Err Channel set. Closing server")
			case <-time.After(30*time.Minute):
				fmt.Println("Server Timeout")
				exit <- true
			}
		}
	}()
	<- exit
	fmt.Println("Terminated from Application")
}

