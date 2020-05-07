package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	Version = "1.0.0"
)

var (
	listenPtr = flag.String("l", "localhost:8080",
		"Listen on `host:port` for incoming traffic to be duplicated")

	primaryPtr = flag.String("p", "localhost:9090",
		"Relay traffic to primary `host:port` and establish a two way TCP connection")

	mirrorPtr = flag.String("m", "localhost:9091",
		"Mirror incoming traffic to `host:port[,host:port]...`. Can specify multiple addresses seperated by a comma. Eg. localhost:9091,localhost:9092")

	debugPtr = flag.Bool("d", false, "Print debug information on stdout")
)

func main() {

	flag.Usage = Usage
	flag.Parse()

	mirrorAddrs := strings.Split(*mirrorPtr, ",")

	fmt.Printf("Listening on                    %s\n", *listenPtr)
	fmt.Printf("Connecting in primary mode to   %s\n", *primaryPtr)
	fmt.Printf("Duplicating incoming traffic to %s\n", *mirrorPtr)

	l, err := net.Listen("tcp", *listenPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listening: ", err.Error())
		os.Exit(1)
	}

	for {
		in, _ := l.Accept()
		fmt.Printf("Incoming connection from %s\n", in)

		p, err := net.Dial("tcp", *primaryPtr)
		if err != nil {
			fmt.Println("Error connecting to primary: ", err.Error())
			os.Exit(1)
		}

		// create array of writers where writers are
		// - all mirrors
		// - primary addr
		// - stdout (if debug)
		ws := make([]io.Writer, len(mirrorAddrs))

		for i, mirrorAddr := range mirrorAddrs {
			m, err := net.Dial("tcp", mirrorAddr)
			if err != nil {
				fmt.Println("Error connecting to the mirror address: ", err.Error())
				os.Exit(1)
			}
			ws[i] = m
		}
		ws = append(ws, p) // add primary

		mw := io.MultiWriter(ws...)

		go io.Copy(mw, in)
		go io.Copy(in, p)

		// fmt.Println("After accept")
		// fmt.Printf("mw = %v\nin = %v\n", mw, in)
		// fmt.Printf("Num goroutines: %d \n", runtime.NumGoroutine())
	}
	// Close the listener when application closes
}

func Usage() {
	fmt.Fprintf(os.Stderr, "tcpmirror version %s\n", Version)
	fmt.Fprintf(os.Stderr, "Usage:   $ tcpmirror -l <listen_addr> -p <primary_addr> -m <mirror_addrs\n")
	fmt.Fprintf(os.Stderr, "Example: $ tcpmirror -l localhost:8080 -p localhost:9090 -m localhost:9091,localhost:9091 \n")
	fmt.Fprintf(os.Stderr, "-----------------------\nFlags:\n")
	flag.PrintDefaults()
}
