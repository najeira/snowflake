package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const Epoch int64 = 1288834974657

const ServerBits uint8 = 10
const SequenceBits uint8 = 12

const ServerShift uint8 = SequenceBits
const TimeShift uint8 = SequenceBits + ServerBits

const ServerMax int = -1 ^ (-1 << ServerBits)

const SequenceMask int32 = -1 ^ (-1 << SequenceBits)

type Server struct {
	port          int
	serverId      int
	lastTimestamp int64
	sequence      int32
	mutex         sync.Mutex
}

func NewServer(p, s int) *Server {
	if s < 0 || ServerMax < s {
		panic(fmt.Errorf("invalid server Id"))
	}
	return &Server{
		port:          p,
		serverId:      s,
		lastTimestamp: 0,
		sequence:      0,
	}
}

func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id, err := s.Next()
	var status int
	var message string
	if err != nil {
		status = http.StatusInternalServerError
		message = err.Error()
	} else {
		status = http.StatusOK
		message = strconv.FormatInt(id, 10)
	}
	w.WriteHeader(status)
	if message != "" {
		bstr := []byte(message)
		w.Write(bstr)
	}
}

func (s *Server) Next() (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	t := now()
	if t < s.lastTimestamp {
		return -1, fmt.Errorf("invalid system clock")
	}
	if t == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & SequenceMask
		if s.sequence == 0 {
			t = s.nextMillis()
		}
	} else {
		s.sequence = 0
	}
	s.lastTimestamp = t
	tp := (t - Epoch) << TimeShift
	sp := int64(s.serverId << ServerShift)
	n := tp | sp | int64(s.sequence)
	//log.Print(n, t, s.serverId, s.sequence)
	return n, nil
}

func (s *Server) nextMillis() int64 {
	t := now()
	for t <= s.lastTimestamp {
		time.Sleep(100 * time.Microsecond)
		t = now()
	}
	return t
}

func now() int64 {
	now := time.Now()
	nows := int64(now.Unix() * 1000)
	nowm := int64(now.Nanosecond()) / time.Millisecond.Nanoseconds()
	nowi := nows + nowm
	return nowi
}

func main() {
	var portNumber int
	var serverId int
	flag.Parse()
	flag.IntVar(&portNumber, "port", 8181, "port")
	flag.IntVar(&serverId, "server", 0, "server")
	server := NewServer(portNumber, serverId)
	log.Fatal(server.ListenAndServe())
}
