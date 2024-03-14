package main

import (
	"context"
	"log"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/flowcontrol"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/MTRNord/matrix_protobuf_fed/rpcserver"

	protocol "github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1"
)

// This implements a dummy server for testing purposes of the rough api design especially around signatures

func main() {
	log.Println("Starting server on port localhost:2000")
	server := rpcserver.NewServer()

	client := protocol.MatrixFederation_ServerToClient(server)
	client.SetFlowLimiter(flowcontrol.NewFixedLimiter(1 << 16))

	rpc.ListenAndServe(context.Background(), "tcp", "localhost:2000", capnp.Client(client))
}
