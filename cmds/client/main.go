package main

import (
	"context"
	"log"
	"net"
	"time"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/MTRNord/matrix_protobuf_fed/cmds/client/helpers"
	protocol "github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1"
)

// This implements a dummy server for testing purposes of the rough api design especially around signatures

func main() {
	log.Println("Connecting to server on port localhost:2000")

	conn, err := net.Dial("tcp", "localhost:2000")
	if err != nil {
		log.Fatalln("Failed to connect to server:", err)
	}
	defer conn.Close()

	log.Println("Connected to server on port localhost:2000")
	log.Println("Creating rpc client...")

	rpc_conn := rpc.NewConn(rpc.NewStreamTransport(conn), nil)
	defer rpc_conn.Close()

	log.Println("Created rpc connection...")

	client := protocol.MatrixFederation(rpc_conn.Bootstrap(context.TODO()))

	log.Println("Created client...")

	timeoutCtx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	err = printServerVersion(timeoutCtx, client)
	if err != nil {
		log.Println("Failed to print server version:", err)
	}
	cancel()

	timeoutCtx, cancel = context.WithTimeout(context.TODO(), time.Second*10)
	err = printServerKeys(timeoutCtx, client)
	if err != nil {
		log.Println("Failed to print server version:", err)
	}
	cancel()
}

func printServerKeys(ctx context.Context, client protocol.MatrixFederation) error {
	callback := helpers.NewKeyStreamCallback()
	callback_client := protocol.StreamCallback_ServerToClient(callback)
	keys_future, release := client.GetKeys(ctx, func(p protocol.MatrixFederation_getKeys_Params) error {
		log.Println("Sending getKeys request...")

		return p.SetCallback(callback_client)
	})
	defer release()

	log.Println("Sent getKeys request...")
	_, err := keys_future.Struct()
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

func printServerVersion(ctx context.Context, client protocol.MatrixFederation) error {
	version_future, release := client.GetVersion(ctx, func(p protocol.MatrixFederation_getVersion_Params) error {
		log.Println("Sending getVersion request...")
		return nil
	})
	defer release()

	log.Println("Sent getVersion request...")

	version_struct, err := version_future.Struct()
	if err != nil {
		return err
	}
	if !version_struct.HasServerVersion() {
		log.Fatalln("No server version")
	}

	server_version, err := version_struct.ServerVersion()
	if err != nil {
		return err
	}

	name, err := server_version.Name()
	if err != nil {
		return err
	}

	version, err := server_version.Version()
	if err != nil {
		return err
	}

	log.Println("Server name:", helpers.QuoteString(name), "version:", helpers.QuoteString(version))
	return nil
}
