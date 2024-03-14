package helpers

import (
	"context"
	"log"

	protocol "github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1"
	"github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1/types"
)

func QuoteString(s string) string {
	return "\"" + s + "\""
}

type KeyStreamCallback struct {
}

func NewKeyStreamCallback() KeyStreamCallback {
	return KeyStreamCallback{}
}

func (k KeyStreamCallback) Write(ctx context.Context, call protocol.StreamCallback_write) error {
	log.Println("StreamCallback.Write")
	p := call.Args()

	response_ptr, err := p.Value()
	if err != nil {
		return err
	}
	response := types.ServerKeysResponse(response_ptr.Struct())

	if response.HasMetadata() {
		metadata, err := response.Metadata()
		if err != nil {
			return err
		}

		servername, err := metadata.ServerName()
		if err != nil {
			return err
		}
		valid_until_ts := metadata.ValidUntilTS()

		log.Println("Server keys metadata:\n", "servername:", QuoteString(servername), "valid_until_ts:", valid_until_ts)
	} else if response.HasVerifyKeys() {
		verify_keys, err := response.VerifyKeys()
		if err != nil {
			return err
		}

		entries, err := verify_keys.Entries()
		if err != nil {
			return err
		}
		len_of_entries := entries.Len()

		// Print the verify keys
		for i := 0; i < len_of_entries; i++ {
			entry := entries.At(i)

			key_id_ptr, err := entry.Key()
			if err != nil {
				return err
			}

			key_id := key_id_ptr.Text()

			key_ptr, err := entry.Value()
			if err != nil {
				return err
			}

			key := key_ptr.Data()

			log.Println("Server keys verify_keys entry:", "key_id:", QuoteString(key_id), "key:", key)
		}
	} else if response.HasOldVerifyKeys() {
		old_verify_keys, err := response.OldVerifyKeys()
		if err != nil {
			return err
		}

		entries, err := old_verify_keys.Entries()
		if err != nil {
			return err
		}
		len_of_entries := entries.Len()

		// Print the verify keys
		for i := 0; i < len_of_entries; i++ {
			entry := entries.At(i)

			key_id_ptr, err := entry.Key()
			if err != nil {
				return err
			}

			key_id := key_id_ptr.Text()

			key_ptr, err := entry.Value()
			if err != nil {
				return err
			}

			key_wrapper := types.ServerKeysResponse_OldVerifyKey(key_ptr.Struct())
			key, err := key_wrapper.Key()
			if err != nil {
				return err
			}

			expired_ts := key_wrapper.ExpiredTS()

			log.Println("Server keys old_verify_keys entry:", "key_id:", QuoteString(key_id), "key:", key, "expired_ts:", expired_ts)
		}
	} else {
		log.Println("Server keys response has no metadata, verify_keys, or old_verify_keys")
	}

	// Print the signatures
	if response.HasSignatures() {
		signatures, err := response.Signatures()
		if err != nil {
			return err
		}
		len_of_signatures := signatures.Len()

		for i := 0; i < len_of_signatures; i++ {
			signature := signatures.At(i)

			server, err := signature.Server()
			if err != nil {
				return err
			}

			signatures_map, err := signature.Signatures()
			if err != nil {
				return err
			}
			signatures_map_entries, err := signatures_map.Entries()
			if err != nil {
				return err
			}
			len_of_signatures_map_entries := signatures_map_entries.Len()

			for j := 0; j < len_of_signatures_map_entries; j++ {
				signature_entry := signatures_map_entries.At(j)

				key_id_ptr, err := signature_entry.Key()
				if err != nil {
					return err
				}
				key_id := key_id_ptr.Text()

				signature_ptr, err := signature_entry.Value()
				if err != nil {
					return err
				}

				signature := signature_ptr.Data()

				log.Println("Server keys signature:", "server:", QuoteString(server), "key_id:", QuoteString(key_id), "signature:", signature)
			}
		}

	} else {
		log.Println("Server keys response has no signatures")
	}
	return nil
}

func (k KeyStreamCallback) Done(ctx context.Context, call protocol.StreamCallback_done) error {
	log.Println("StreamCallback.Done")
	return nil
}
