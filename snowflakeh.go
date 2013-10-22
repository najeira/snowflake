package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const Epoch int64 = 1288834974657

const ServerBits uint8 = 5
const DatacenterBits uint8 = 5
const SequenceBits uint8 = 12

const ServerShift uint8 = SequenceBits
const DatacenterShift uint8 = SequenceBits + ServerBits
const TimeShift uint8 = SequenceBits + ServerBits + DatacenterBits

const ServerMax int = -1 ^ (-1 << ServerBits)
const DatacenterMax int = -1 ^ (-1 << DatacenterBits)

const SequenceMask int32 = -1 ^ (-1 << SequenceBits)

type Server struct {
	port          int
	serverId      int
	datacenterId  int
	lastTimestamp int64
	sequence      int32
}

var (
	Port         int
	ServerId     int
	DatacenterId int
)

func NewServer(p, s, d int) *Server {
	if s < 0 || ServerMax < s {
		panic(fmt.Errorf("invalid server Id"))
	}
	if d < 0 || DatacenterMax < d {
		panic(fmt.Errorf("invalid datacenter Id"))
	}
	return &Server{
		port:          p,
		serverId:      s,
		datacenterId:  d,
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
	dp := int64(s.datacenterId << DatacenterShift)
	sp := int64(s.serverId << ServerShift)
	n := tp | dp | sp | int64(s.sequence)
	log.Print(n, t, s.datacenterId, s.serverId, s.sequence)
	return n, nil
}

func (s *Server) nextMillis() int64 {
	t := now()
	for t <= s.lastTimestamp {
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
	flag.Parse()
	flag.IntVar(&Port, "port", 8181, "port")
	flag.IntVar(&ServerId, "server", 0, "server")
	flag.IntVar(&DatacenterId, "datacenter", 0, "datacenter")
	server := NewServer(Port, ServerId, DatacenterId)
	log.Fatal(server.ListenAndServe())
}
