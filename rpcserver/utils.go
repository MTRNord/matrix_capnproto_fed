package rpcserver

import (
	"crypto/ed25519"

	capnp "capnproto.org/go/capnp/v3"
	"github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1/types"
)

// The go parser currently doesnt support custom annotations so we keep track of it manually
const (
	GetVersion_MethodID       = 0xab1eb3e81f3344d1
	GetKeys_MethodID          = 0x932dd11596dad50e
	SendTransactions_MethodID = 0xec6b6ce167005c84
	Backfill_MethodID         = 0x970e8e8dfe8f0ced
)

// A KeyID is the ID of a ed25519 key used to sign Protobuf.
// The key IDs have a format of "ed25519:[0-9A-Za-z]+"
// If we switch to using a different signing algorithm then we will change the
// prefix used.
type KeyID string

// This signs a capnproto byte sequeence in canonical form (has to be canonicalized by the caller) and adds the signature to the signatures list.
// This is the capnproto equivalent to https://spec.matrix.org/v1.9/appendices/#signing-json
// This is not expected to yield the same signature as the json counterpart as the capnproto serialization is different.
func SignCapnproto(signingName string, keyID KeyID, privateKey ed25519.PrivateKey, message []byte, signatures *types.Signature_List) error {
	signature := ed25519.Sign(privateKey, message)

	// Go doesnt support generic types from go-capnp yet https://github.com/capnproto/go-capnp/issues/27
	list_size := signatures.Len()
	map_entry := signatures.At(list_size - 1)
	map_entry.SetServer(signingName)
	signatures_map, err := map_entry.NewSignatures()
	if err != nil {
		return err
	}
	signatures_map_wrapper := FromMap(&signatures_map, 1)
	key, err := capnp.NewText(signatures_map_wrapper.Segment(), string(keyID))
	if err != nil {
		return err
	}

	data, err := capnp.NewData(signatures_map_wrapper.Segment(), signature)
	if err != nil {
		return err
	}
	signatures_map_wrapper.AddEntry(key.ToPtr(), data.ToPtr())

	return nil
}
