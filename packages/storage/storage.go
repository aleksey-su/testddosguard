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

type StoredPackage struct {
	Domain string
	IP     uint32
	TTL    uint8
}

type Storage struct {
	data map[*StoredPackage]struct{}
	mu   sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[*StoredPackage]struct{}),
	}
}

func (s *Storage) Add(domain, IPv4String string) {
	defer s.mu.Unlock()
	s.mu.Lock()

	log.Printf("add %s %s", domain, IPv4String)
	IPv4Int := uint32(IP4toInt(IPv4String))
	s.data[&StoredPackage{
		Domain: domain,
		IP:     IPv4Int,
		TTL:    10,
	}] = struct{}{}
}

func (s *Storage) Print() []*StoredPackage {

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			for k, _ := range s.data {
				log.Printf("domain: %s, ip: %d - %s, ttl: %d\n", k.Domain, k.IP, InttoIPv4(k.IP).String(), k.TTL)
				k.TTL--
				if k.TTL <= 0 {
					delete(s.data, k)
				}
			}
			s.mu.Unlock()
		default:
			continue
		}
	}
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
