package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/flowcontrol"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/MTRNord/matrix_protobuf_fed/rpcserver"

	protocol "github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1"
)

// This implements a dummy server for testing purposes of the rough api design especially around signatures

// Serve serves a Cap'n Proto RPC to incoming connections.
//
// Serve will take ownership of bootstrapClient and release it after the listener closes.
//
// Serve exits with the listener error if the listener is closed by the owner.
func Serve(lis net.Listener, boot capnp.Client) error {
	if !boot.IsValid() {
		err := errors.New("bootstrap client is not valid")
		return err
	}
	// Since we took ownership of the bootstrap client, release it after we're done.
	defer boot.Release()
	for {
		// Accept incoming connections
		conn, err := lis.Accept()
		if err != nil {
			conn.Close()
			return err
		}
		defer conn.Close()

		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits, so use AddRef to avoid releasing the provided bootstrap client capability.
		opts := rpc.Options{
			BootstrapClient: boot.AddRef(),
			Logger:          slog.Default(),
		}
		// For each new incoming connection, create a new RPC transport connection that will serve incoming RPC requests
		transport := rpc.NewStreamTransport(conn)
		defer transport.Close()
		rpc_conn := rpc.NewConn(transport, &opts)
		defer rpc_conn.Close()
	}
}

// ListenAndServe opens a listener on the given address and serves a Cap'n Proto RPC to incoming connections
//
// network and address are passed to net.Listen. Use network "unix" for Unix Domain Sockets
// and "tcp" for regular TCP IP4 or IP6 connections.
//
// ListenAndServe will take ownership of bootstrapClient and release it on exit.
func ListenAndServe(ctx context.Context, network, addr string, bootstrapClient capnp.Client) error {
	listener, err := net.Listen(network, addr)

	if err == nil {
		// to close this listener, close the context
		go func() {
			<-ctx.Done()
			_ = listener.Close()
		}()
		err = Serve(listener, bootstrapClient)
	}
	return err
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")
var f *os.File

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // subscribe to system signals
	onKill := func(c chan os.Signal) {
		select {
		case <-c:
			fmt.Println("Got killed")
			pprof.StopCPUProfile()
			f.Close()
			if *memprofile != "" {
				f, err := os.Create(*memprofile)
				if err != nil {
					log.Fatal(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
			}
			os.Exit(0)
		}
	}

	// try to handle os interrupt(signal terminated)
	go onKill(c)

	log.Println("Starting server on port localhost:8449")
	server := rpcserver.NewServer()

	client := protocol.MatrixFederation_ServerToClient(server)
	client.SetFlowLimiter(flowcontrol.NewFixedLimiter(1 << 17))

	ListenAndServe(context.Background(), "tcp", "localhost:8449", capnp.Client(client))
}
