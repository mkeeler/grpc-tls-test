package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/mkeeler/grpc-tls-test/internal/grpc/greeting"
)

var (
	plainAddr = &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8502}
	tlsAddr   = &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 8503}
)

func main() {
	log.Printf("Staring signal notifier")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log.Printf("Creating listener for %s", plainAddr.String())
	plainLn, err := net.ListenTCP("tcp", plainAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s - %v", plainAddr.String(), err)
	}
	defer plainLn.Close()

	log.Printf("Loading TLS certificates")
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("Failed to load certificates: %v", err)
	}

	log.Printf("Creating TLS listener for %s", tlsAddr.String())
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	tlsLn, err := tls.Listen("tcp", tlsAddr.String(), config)
	if err != nil {
		log.Fatalf("Failed to listen on %s - %v", tlsAddr.String(), err)
	}
	defer tlsLn.Close()

	opts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			// This must be less than the keealive.ClientParameters Time setting, otherwise
			// the server will disconnect the client for sending too many keepalive pings.
			// Currently the client param is set to 30s.
			MinTime: 15 * time.Second,
		}),
	}

	log.Printf("Creating the gRPC server")
	srv := grpc.NewServer(opts...)
	log.Printf("Registering the GreetingService with the gRPC server")
	greeting.NewServer().Register(srv)

	var wg sync.WaitGroup
	wg.Add(2)

	serveLn := func(ln net.Listener, name string) {
		log.Printf("Serving %s on %s", name, ln.Addr().String())
		srv.Serve(ln)
		log.Printf("Finished serving %s on %s", name, ln.Addr().String())
		wg.Done()
	}

	go serveLn(plainLn, "gRPC")
	go serveLn(tlsLn, "gRPC + TLS")

	<-ctx.Done()
	log.Printf("Caught signal - exiting")
	srv.Stop()
	log.Printf("Initiated gRPC server shutdown")
	wg.Wait()
}
