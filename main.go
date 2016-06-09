package main

import (
	"bufio"
	"flag"
	"net"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/gwilym/go-listenerreader"
)

type Flags struct {
	Addr       string
	BufMaxSize int
	Delim      string
	Net        string
}

func main() {
	flags := Flags{}

	flag.StringVar(&flags.Addr, "addr", ":3333", "address to listen on")
	flag.IntVar(&flags.BufMaxSize, "bufmaxsize", 64*1024, "max token buffer size in bytes")
	flag.StringVar(&flags.Delim, "delim", "\n", "delimiter byte")
	flag.StringVar(&flags.Net, "net", "tcp", "network to listen on")

	flag.Parse()

	os.Exit(run(flags))
}

func run(flags Flags) int {
	logger := log.NewLogfmtLogger(os.Stderr)

	l, err := net.Listen(flags.Net, flags.Addr)
	if err != nil {
		logger.Log("net", flags.Net, "addr", flags.Addr, "error", err)
		return 1
	}

	logger.Log("net", flags.Net, "addr", flags.Addr, "msg", "listening")

	delim := flags.Delim[0]

	// these should be command-line adjustable, once I figure out how best to describe them
	bufStartSize := 1
	chanLen := 1

	lr := listenerreader.NewListenerReader(l, delim, bufStartSize, flags.BufMaxSize, chanLen)

	lr.AcceptFunc = func(conn net.Conn) {
		go logger.Log("net", flags.Net, "localaddr", conn.LocalAddr(), "remoteaddr", conn.RemoteAddr(), "msg", "connected")
	}

	lr.CloseFunc = func(conn net.Conn) {
		go logger.Log("net", flags.Net, "localaddr", conn.LocalAddr(), "remoteaddr", conn.RemoteAddr(), "msg", "closing")
	}

	s := bufio.NewScanner(lr)

	for s.Scan() {
		b := s.Bytes()
		b = append(b, '\n')
		os.Stdout.Write(b)
	}

	if s.Err() != nil {
		logger.Log("net", flags.Net, "addr", flags.Addr, "error", s.Err())
		return 1
	}

	return 0
}
