package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

const (
	host  = ""
	port  = "1234"
	proto = "tcp4"
)

func main() {
	listener, err := net.Listen(proto, net.JoinHostPort(host, port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	chSayings := make(chan string, 10)

	signal.Notify(c, os.Interrupt)

	var wg sync.WaitGroup

	go func() {

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Accept failed: %v ", err)
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				HangleSayng(ctx, conn, chSayings)

			}()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		GenerateSayngs(ctx, chSayings)
	}()

	select {
	case <-c:
		cancel()
		log.Println("Service stoped..........")
		wg.Wait()
		os.Exit(1)
	}

}

func HangleSayng(ctx context.Context, conn net.Conn, sayngs <-chan string) {
	defer conn.Close()
	tiker := time.NewTicker(1 * time.Second)
	thirdAccount := 2
	for {
		select {
		case <-ctx.Done():
			return
		case <-tiker.C:
			if thirdAccount == 0 {
				say := <-sayngs
				_, err := conn.Write([]byte(say))
				if err != nil {
					log.Println("Connection close, host: ", conn.LocalAddr().String())
					return
				}
				thirdAccount = 2
			} else {
				conn.Write([]byte(fmt.Sprintf("%d sec to sayngs...\n", thirdAccount)))
				thirdAccount--
			}
		}
	}
}

func GenerateSayngs(ctx context.Context, chSayngs chan<- string) {
	bsayngs, err := ioutil.ReadFile("sayngs.txt")
	if err != nil {
		log.Fatal(err)
	}
	sayngs := strings.Split(string(bsayngs), "\n")
	rand.Seed(time.Now().UnixNano())
	for {
		select {
		case <-ctx.Done():
			return
		case chSayngs <- sayngs[0+rand.Intn(len(sayngs)-1)] + "\n":
		}
	}

}
