package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

const Epoch int64 = 1288834974657

const ServerBits uint8 = 10
const SequenceBits uint8 = 12

const ServerShift uint8 = SequenceBits
const TimeShift uint8 = SequenceBits + ServerBits

const ServerMax int = -1 ^ (-1 << ServerBits)

const SequenceMask int32 = -1 ^ (-1 << SequenceBits)

type Worker struct {
	serverId      int
	lastTimestamp int64
	sequence      int32
}

type Server struct {
	port    int
	workers chan *Worker
}

func NewWorker(serverId int) *Worker {
	if serverId < 0 || ServerMax < serverId {
		panic(fmt.Errorf("invalid server Id"))
	}
	return &Worker{
		serverId:      serverId,
		lastTimestamp: 0,
		sequence:      0,
	}
}

func NewServer(port, serverId, serverNum int) *Server {
	workers := make(chan *Worker, serverNum)
	for n := 0; n < serverNum; n++ {
		workers <- NewWorker(serverId + n)
	}
	return &Server{
		port:    port,
		workers: workers,
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
	worker := <-s.workers
	id, err := worker.Next()
	s.workers <- worker
	return id, err
}

func (s *Worker) Next() (int64, error) {
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

func (s *Worker) nextMillis() int64 {
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

func now2() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	var portNumber int
	var serverId int
	var serverNum int
	flag.IntVar(&portNumber, "port", 8181, "port")
	flag.IntVar(&serverId, "server", 0, "server")
	flag.IntVar(&serverNum, "num", 1, "num")
	flag.Parse()
	server := NewServer(portNumber, serverId, serverNum)
	log.Fatal(server.ListenAndServe())
}
