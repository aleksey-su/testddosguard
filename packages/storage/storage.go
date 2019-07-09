package storage

import (
	"log"
	"math/big"
	"net"
	"sync"
	"time"
)

type Package struct {
	Domain string `msgpack:"domain"`
	IP     string `msgpack:"ip"`
}

type Storage struct {
	data   map[string]uint32
	mu     sync.Mutex
	ticker *time.Ticker
}

func NewStorage() *Storage {
	return &Storage{
		data:   make(map[string]uint32),
		ticker: time.NewTicker(10 * time.Second),
	}
}

func (s *Storage) Add(domain, IPv4String string) {
	defer s.mu.Unlock()
	s.mu.Lock()

	select {
	case <-s.ticker.C:
		log.Printf("add %s %s", domain, IPv4String)
		IPv4Int := uint32(IP4toInt(IPv4String))
		s.data[domain] = IPv4Int
	default:
		return
	}
}

func (s *Storage) GetAll() map[string]uint32 {
	return s.data
}

func (s *Storage) StopTicker() {
	s.ticker.Stop()
}

func IP4toInt(IPv4String string) uint64 {
	IPv4Address := net.ParseIP(IPv4String)
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address)
	return IPv4Int.Uint64()
}

func InttoIPv4(ipnr uint32) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}
