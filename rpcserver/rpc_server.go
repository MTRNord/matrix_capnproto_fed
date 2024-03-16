package rpcserver

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"log"
	"time"

	capnp "capnproto.org/go/capnp/v3"

	protocol "github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1"
	"github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1/types"
)

type SigningKeyWrapper struct {
	*protocol.MatrixFederation
	entityName string
	keyID      KeyID
	privateKey ed25519.PrivateKey
}

type RPCMatrixServer struct {
	signing_key SigningKeyWrapper
}

func NewServer() RPCMatrixServer {
	// Check JSON verification using the test vectors from https://matrix.org/docs/spec/appendices.html
	seed, err := base64.RawStdEncoding.DecodeString("YJDBA9Xnr2sVqXD9Vj7XVUnmFZcZrlw8Md7kMW+3XA1")
	if err != nil {
		log.Fatal(err)
	}
	random := bytes.NewBuffer(seed)
	entityName := "domain"
	keyID := KeyID("ed25519:1")

	_, privateKey, err := ed25519.GenerateKey(random)
	if err != nil {
		log.Fatal(err)
	}

	return RPCMatrixServer{
		signing_key: SigningKeyWrapper{
			entityName: entityName,
			keyID:      keyID,
			privateKey: privateKey,
		},
	}
}

func (s RPCMatrixServer) GetVersion(ctx context.Context, call protocol.MatrixFederation_getVersion) error {
	call.Go()
	res, err := call.AllocResults() // Allocate the results struct
	if err != nil {
		return err
	}

	version, err := res.NewServerVersion()
	if err != nil {
		return err
	}
	version.SetName("Matrix Federation Cap'n'Proto RPC Proxy")
	version.SetVersion("0.1.0")

	return nil
}

func (s RPCMatrixServer) GetKeys(ctx context.Context, call protocol.MatrixFederation_getKeys) error {
	call.Go()

	client := call.Args().Callback()
	defer client.Release()

	err := client.Write(ctx, func(p protocol.StreamCallback_write_Params) error {
		log.Println("Sending server keys metadata response...")

		response, err := types.NewServerKeysResponse(p.Segment())
		if err != nil {
			return err
		}
		p.SetValue(response.ToPtr())

		metadata, err := response.NewMetadata()
		if err != nil {
			return err
		}
		metadata.SetServerName("placeholder")
		metadata.SetValidUntilTS(time.Now().UTC().Add(time.Hour * 24).Unix())

		metadata_bytes, err := capnp.Canonicalize(capnp.Struct(metadata))
		if err != nil {
			return err
		}
		signatures, err := response.NewSignatures(1)
		if err != nil {
			return err
		}
		err = SignCapnproto("placeholder", "placeholder", s.signing_key.privateKey, metadata_bytes, &signatures)
		if err != nil {
			return err
		}
		return response.SetSignatures(signatures)
	})
	if err != nil {
		return err
	}

	err = client.Write(ctx, func(p protocol.StreamCallback_write_Params) error {
		log.Println("Sending server keys verify_keys response...")

		response, err := types.NewServerKeysResponse(p.Segment())
		if err != nil {
			return err
		}
		p.SetValue(response.ToPtr())

		verify_keys_raw, err := response.NewVerifyKeys()
		if err != nil {
			return err
		}
		verify_keys := FromMap(&verify_keys_raw, 1)
		key, err := capnp.NewText(verify_keys.Segment(), "placeholder")
		if err != nil {
			return err
		}

		log.Println("Original Key bytes:", s.signing_key.privateKey.Public().(ed25519.PublicKey))
		data, err := capnp.NewData(verify_keys.Segment(), s.signing_key.privateKey.Public().(ed25519.PublicKey))
		if err != nil {
			return err
		}
		verify_keys.AddEntry(key.ToPtr(), data.ToPtr())

		verify_keys_bytes, err := capnp.Canonicalize(capnp.Struct(verify_keys_raw))
		if err != nil {
			return err
		}
		signatures, err := response.NewSignatures(1)
		if err != nil {
			return err
		}
		err = SignCapnproto("placeholder", "placeholder", s.signing_key.privateKey, verify_keys_bytes, &signatures)
		if err != nil {
			return err
		}
		return response.SetSignatures(signatures)
	})
	if err != nil {
		return err
	}

	_, release := client.Done(ctx, nil)
	defer release()

	if err := client.WaitStreaming(); err != nil {
		return err
	}

	return err
}

func (s RPCMatrixServer) SendTransactions(context.Context, protocol.MatrixFederation_sendTransactions) error {
	return nil
}

func (s RPCMatrixServer) Backfill(context.Context, protocol.MatrixFederation_backfill) error {
	return nil
}
