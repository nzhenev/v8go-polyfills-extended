package crypto

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"
)

//NOTE: see https://github.com/MicahParks/keyfunc for future modification
func parseKey(keyDataBytes []byte) (*rsa.PublicKey, error) {
	key, err := jwk.ParseKey(keyDataBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing keyDataBytes: %#v", err)
	}
	var rawkey interface{} // This is the raw key, like *rsa.PrivateKey or *ecdsa.PrivateKey
	if err := key.Raw(&rawkey); err != nil {
		return nil, fmt.Errorf("failed to create public key: %#v", err)
	}

	// We know this is an RSA Key so...
	rsa, ok := rawkey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected ras key, got %T", rawkey)
	}

	return rsa, nil
}

func parseKeySet(keySetDataBytes []byte) ([]*rsa.PublicKey, error) {
	set, err := jwk.Parse(keySetDataBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing keySetDataBytes: %#v", err)
	}
	result := make([]*rsa.PublicKey, 0)
	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		var rawkey interface{} // This is the raw key, like *rsa.PrivateKey or *ecdsa.PrivateKey
		if err := key.Raw(&rawkey); err != nil {
			return nil, fmt.Errorf("failed to create public key: %#v", err)
		}
		//fmt.Println("=>", key.KeyID())
		// We know this is an RSA Key so...
		rsa, ok := rawkey.(*rsa.PublicKey)
		if !ok {
			fmt.Printf("expected ras key, got %T", rawkey)
		} else {
			// append the key to the result list
			result = append(result, rsa)
			return result, nil
		}
	}
	if len(result) < 1 {
		return nil, fmt.Errorf("failed to parse any public key from ketset: %s", keySetDataBytes)
	}
	return result, nil
}
