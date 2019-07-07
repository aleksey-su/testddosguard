package main

import (
	"bufio"
	"flag"
	"log"
	"net"

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

	packages := []*Package{
		&Package{Domain: "google.ru", IP: "74.125.131.94", TTL: "10"},
		&Package{Domain: "yandex.ru", IP: "77.88.55.55", TTL: "10"},
		&Package{Domain: "habrahabr.ru", IP: "178.248.237.68", TTL: "1"},
	}

	conn, _ := net.Dial("tcp", tcpaddr+":"+tcpport)
	defer conn.Close()
	for _, p := range packages {

		b, err := msgpack.Marshal(p)
		if err != nil {
			panic(err)
		}

		log.Printf("Send %s %s %s", p.Domain, p.IP, p.TTL)
		// log.Println(b)

		// Send to socket
		w := bufio.NewWriter(conn)
		w.Write(append(b, 10))
		w.Flush()

		// Resive message
		message, _ := bufio.NewReader(conn).ReadString('\n')
		log.Println("Message from server: " + message)
		// time.Sleep(2 * time.Second)
	}
}

type Package struct {
	Domain string
	IP     string
	TTL    string
}
