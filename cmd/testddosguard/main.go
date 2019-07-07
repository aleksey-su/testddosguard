package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"storage"
	"strconv"
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
				// log.Println(scanr.Text())

				var item storage.Package

				err := msgpack.Unmarshal(scanr.Bytes(), &item)
				if err != nil {
					log.Println(err)
					w.WriteString(strings.ToUpper("BAD") + "\n")
					w.Flush()
					continue
				}

				ttl, err := strconv.Atoi(item.TTL)

				if err != nil {
					log.Println(err)
					w.WriteString(strings.ToUpper("BAD") + "\n")
					w.Flush()
					continue
				}

				if ttl == 10 {
					st.Add(item.Domain, item.IP)
				}

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

	log.Println("Server exiting")

}