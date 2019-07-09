package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"storage"
	"strings"
	"syscall"
	"tcpserver"
	"time"

	"github.com/vmihailenco/msgpack"
)

var tcpaddr string
var tcpport string

func init() {
	flag.StringVar(&tcpaddr, "tcpaddr", "0.0.0.0", "TCP address")
	flag.StringVar(&tcpport, "tcpport", "6000", "TCP port")
}

func main() {
	flag.Parse()

	st := storage.NewStorage()

	srv := tcpserver.NewTCPServer(tcpaddr, tcpport, func(conn *tcpserver.SafeConn) error {

		r := bufio.NewReader(conn)
		w := bufio.NewWriter(conn)
		scanr := bufio.NewScanner(r)
		// scanr.Split(bufio.ScanBytes)

		sc := make(chan bool)
		deadline := time.After(conn.IdleTimeout)
		for {
			go func(s chan bool) {
				s <- scanr.Scan()
			}(sc)
			select {
			case <-deadline:
				return nil
			case scanned := <-sc:
				if !scanned {
					if err := scanr.Err(); err != nil {
						return err
					}
					return nil
				}

				t := scanr.Text()
				b := scanr.Bytes()
				// log.Println(t)
				// log.Println(b)

				var item storage.Package

				err := msgpack.Unmarshal(b, &item)
				if err != nil {
					log.Printf("Bad request %s", t)
					w.WriteString(strings.ToUpper("BAD") + "\n")
					w.Flush()
					continue
				}

				// log.Println(item)

				st.Add(item.Domain, item.IP)

				w.WriteString(strings.ToUpper("OK") + "\n")
				w.Flush()
				deadline = time.After(conn.IdleTimeout)
			}
		}

		return nil
	})

	go func() {
		// Start tcp server
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	go func() {
		for {
			for k, v := range st.GetAll() {
				fmt.Printf("%s %d %s\n", k, v, storage.InttoIPv4(v).String())
			}
			time.Sleep(time.Second)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	srv.Shutdown()
	st.StopTicker()

	log.Println("Server exiting")

}
