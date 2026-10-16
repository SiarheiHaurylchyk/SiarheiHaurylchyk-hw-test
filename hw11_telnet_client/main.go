package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var timeout time.Duration

	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}

	address := net.JoinHostPort(flag.Arg(0), flag.Arg(1))
	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...%v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sendErr := make(chan error, 1)
	recvErr := make(chan error, 1)

	go func() { sendErr <- client.Send() }()
	go func() { recvErr <- client.Receive() }()

	select {
	case <-ctx.Done():
		return
	case err := <-sendErr:
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "...%v\n", err)
			return
		}
		fmt.Fprintln(os.Stderr, "...EOF")
	case err := <-recvErr:
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "...%v\n", err)
			return
		}
		fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
	}
}
