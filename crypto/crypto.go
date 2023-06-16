package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/esoptra/v8go"
	"github.com/nzhenev/v8go-polyfills-extended/uuid"
)

type Crypto struct {
	KeyMap sync.Map
}

type KeyAlgorithm string

const (
	RSA1_5       = KeyAlgorithm("RSASSA-PKCS1-v1_5") // RSA-PKCS1v1.5
	RSA_OAEP     = KeyAlgorithm("RSA-OAEP")          // RSA-OAEP-SHA1
	RSA_OAEP_256 = KeyAlgorithm("RSA-OAEP-256")      // RSA-OAEP-SHA256
)

func NewCrypto(opt ...Option) *Crypto {
	c := &Crypto{}

	for _, o := range opt {
		o.apply(c)
	}

	return c
}

// cryptoVerifyFunctionCallback implements https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/verify
// const result = crypto.subtle.verify(algorithm, key, signature, data);
// result is a Promise with a Boolean: true if the signature is valid, false otherwise.
func (c *Crypto) cryptoVerifyFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		iso := ctx.Isolate()
		resolver, err := v8go.NewPromiseResolver(ctx)
		if err != nil {
			strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error creating newPromiseResolver with ctx: %#v", err))
			return iso.ThrowException(strErr)
		}
		go func() {
			passed := false
			defer func() {
				v, err := v8go.NewValue(iso, passed)
				if err != nil {
					resolver.Reject(newErrorValue(iso, "Error creating newvalue: %#v\n", err))
					return
				}
				resolver.Resolve(v)
			}()
			args := info.Args()
			if len(args) != 4 {
				resolver.Reject(newErrorValue(iso, "Expected algorithm, key, signature, data (4) arguments\n"))
				return
			}

			algorithm, _, err := getAlgorithm(args[0])
			if err != nil {
				resolver.Reject(newErrorValue(iso, "Error parsing Algorithm arg: %#v\n", err))
				return
			}
			_ = algorithm //Need this in future
			keyBytes, err := args[1].MarshalJSON()
			if err != nil {
				resolver.Reject(newErrorValue(iso, "Error marshalling key arg: %#v\n", err))
				return
			}

			if !args[2].IsUint8Array() {
				resolver.Reject(newErrorValue(iso, "Expecting signature in []byte(ArrayBuffer) format\n"))
				return
			}

			key := &CryptoKey{}
			err = json.Unmarshal(keyBytes, key)
			if err != nil {
				resolver.Reject(newErrorValue(iso, "Error unmarshalling keybytes arg: %#v\n", err))
				return
			}

			if key.Type == "public" {
				//this expecting public rsa key
				pubKey, ok := c.KeyMap.Load(key.Kid)
				if !ok {
					resolver.Reject(newErrorValue(iso, "Invalid Key : %#v\n", key))
					return
				}
				sign := args[2]
				//fmt.Println("sign", string(sign))
				payload := args[3]
				//fmt.Println("payload", string(payload))

				var rsaKey *rsa.PublicKey

				if rsaKey, ok = pubKey.(*rsa.PublicKey); !ok {
					resolver.Reject(newErrorValue(iso, "Invalid Key : %#v\n", pubKey))
					return
				}
				var hash crypto.Hash

				switch algorithm.(*RSAAlgoOut).Hash.Name {
				case "SHA-256":
					hash = crypto.SHA256
				case "SHA-384":
					hash = crypto.SHA384
				case "SHA-512":
					hash = crypto.SHA512
				default:
					resolver.Reject(newErrorValue(iso, "unknown/unsupported algorithm"))
					return
				}

				hasher := hash.New()

				// According to documentation, Write() on hash never fails
				_, _ = hasher.Write([]byte(payload.String()))
				hashed := hasher.Sum(nil)
				err = rsa.VerifyPKCS1v15(rsaKey, hash, hashed, []byte(sign.String()))
				passed = (err == nil)
			}

		}()

		return resolver.GetPromise().Value
	}
}

// cryptoImportKeyFunctionCallback implements https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/importKey
// const result = crypto.subtle.importKey(format, keyData, algorithm, extractable, keyUsages);
// result is a Promise that fulfills with the imported key as a CryptoKey object.
func (c *Crypto) cryptoImportKeyFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		iso := ctx.Isolate()
		resolver, err := v8go.NewPromiseResolver(ctx)
		if err != nil {
			strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error creating newPromiseResolver with ctx: %#v", err))
			return iso.ThrowException(strErr)
		}
		go func() {
			args := info.Args()

			format := args[0].String()
			var result interface{}
			switch format {
			case "jwk":
				keyData, err := args[1].AsObject() //object type
				if err != nil {
					resolver.Reject(newErrorValue(iso, "error getting keyData %#v", err))
					return
				}
				algorithm, _, err := getAlgorithm(args[2])
				if err != nil {
					resolver.Reject(newErrorValue(iso, "Error parsing Algorithm arg: %#v\n", err))
					return
				}

				extractable := args[3] //boolean
				if !extractable.IsBoolean() {
					resolver.Reject(newErrorValue(iso, "Expected extractable argument as boolean type\n"))
					return
				}

				keyUsages := args[4] //array
				if !keyUsages.IsArray() {
					resolver.Reject(newErrorValue(iso, "Expected keyUsages argument as array type\n"))
					return
				}

				keyDataBytes, err := keyData.MarshalJSON()
				if err != nil {
					resolver.Reject(newErrorValue(iso, "error marshalling keyData %#v", err))
					return
				}
				//fmt.Println(string(keyDataBytes))

				isKeySet := keyData.Has("keys")

				var key interface{}
				//this expecting public rsa key
				if isKeySet {
					keys, err := parseKeySet(keyDataBytes)
					if err != nil {
						resolver.Reject(newErrorValue(iso, "Could not parse DER encoded key (encryption key): %#v", err))
						return
					}
					//select the first key from the set
					key = keys[0]
				} else {
					key, err = parseKey(keyDataBytes)
					if err != nil {
						resolver.Reject(newErrorValue(iso, "Could not parse DER encoded key (encryption key): %#v", err))
						return
					}
				}
				//fmt.Println(key)
				miniPub := uuid.NewUuid()
				c.KeyMap.Store(miniPub, key)

				result = &CryptoKey{
					Type:        "public",
					Kid:         miniPub,
					Extractable: extractable.Boolean(),
					Algorithm:   algorithm,
					Usages:      keyUsages,
				}

			default:
				resolver.Reject(newErrorValue(iso, "format %q not supported", format))
				return
			}

			resultBytes, err := json.Marshal(result)
			if err != nil {
				resolver.Reject(newErrorValue(iso, "error marshalling jsonKey: %#v", err))
				return
			}
			v, err := v8go.JSONParse(info.Context(), string(resultBytes))
			if err != nil {
				resolver.Reject(newErrorValue(iso, "error jsonParse on result: %#v", err))
				return
			}

			resolver.Resolve(v)
		}()

		return resolver.GetPromise().Value
	}
}

type RSAAlgoIn struct {
	Name           string           `json:"name"`                         //"RSA-OAEP",
	ModulusLength  int              `json:"modulusLength" default:"4096"` // 4096,
	PublicExponent map[string]uint8 `json:"publicExponent"`               // new Uint8Array([1, 0, 1]),
	Hash           string           `json:"hash"`                         // "SHA-256"
}

type RSAAlgoOut struct {
	Name           string                  `json:"name"`                         //"RSA-OAEP",
	ModulusLength  int                     `json:"modulusLength" default:"4096"` // 4096,
	PublicExponent map[string]uint8        `json:"publicExponent"`               // new Uint8Array([1, 0, 1]),
	Hash           HashAlgorithmIdentifier `json:"hash"`                         // "SHA-256"
}

type HashAlgorithmIdentifier struct {
	Name string `json:"name"`
}

// for symmetric algo https://developer.mozilla.org/en-US/docs/Web/API/CryptoKey
type CryptoKey struct {
	Type        string      `json:"type"`
	Kid         string      `json:"kid"` //additional property to refer the actual key
	Extractable bool        `json:"extractable"`
	Algorithm   interface{} `json:"algorithm"`
	Usages      interface{} `json:"usages"`
}

// for public-key algorithms
type CryptoKeyPair struct {
	PrivateKey CryptoKey `json:"privateKey"`
	PublicKey  CryptoKey `json:"publicKey"`
}

// cryptoGenerateKeyFunctionCallback implements https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/generateKey
// const result = crypto.subtle.generateKey(algorithm, extractable, keyUsages);
// result is a Promise that fulfills with a CryptoKey (for symmetric algorithms) or a CryptoKeyPair (for public-key algorithms).
func (c *Crypto) cryptoGenerateKeyFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		iso := ctx.Isolate()
		resolver, err := v8go.NewPromiseResolver(ctx)
		if err != nil {
			strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error creating newPromiseResolver with ctx: %#v", err))
			return iso.ThrowException(strErr)
		}
		go func() {
			args := info.Args()
			if len(args) != 3 {
				resolver.Reject(newErrorValue(iso, "Expected algorithm, extractable, keyUsages (3) arguments\n"))
				return
			}

			algorithm, _, err := getAlgorithm(args[0])
			if err != nil {
				resolver.Reject(newErrorValue(iso, "Error parsing Algorithm arg: %#v\n", err))
				return
			}

			extractable := args[1] //boolean
			if !extractable.IsBoolean() {
				resolver.Reject(newErrorValue(iso, "Expected extractable argument as boolean type\n"))
				return
			}

			keyUsages := args[2] //array
			if !keyUsages.IsArray() {
				resolver.Reject(newErrorValue(iso, "Expected keyUsages argument as array type\n"))
				return
			}

			var result interface{}

			primeBits := 2048

			rsaAlgo := algorithm.(*RSAAlgoOut)
			if rsaAlgo.ModulusLength != 0 {
				primeBits = rsaAlgo.ModulusLength
			}
			// The GenerateKey method takes in a reader that returns random bits, and
			// the number of bits
			privateKey, err := rsa.GenerateKey(rand.Reader, primeBits) //2048 by default
			if err != nil {
				resolver.Reject(newErrorValue(iso, "error generating RSA key: %#v", err))
				return
			}

			//store a pointer reference with the fetcher
			miniPriv := uuid.NewUuid()
			c.KeyMap.Store(miniPriv, privateKey)
			miniPub := uuid.NewUuid()
			c.KeyMap.Store(miniPub, &privateKey.PublicKey)

			result = &CryptoKeyPair{
				PrivateKey: CryptoKey{
					Type:        "private",
					Kid:         miniPriv,
					Extractable: extractable.Boolean(),
					Algorithm:   algorithm,
					Usages:      keyUsages.Object(),
				},
				PublicKey: CryptoKey{
					Type:        "public",
					Kid:         miniPub,
					Extractable: extractable.Boolean(),
					Algorithm:   algorithm,
					Usages:      keyUsages.Object(),
				},
			}

			resultBytes, err := json.Marshal(result)
			if err != nil {
				resolver.Reject(newErrorValue(iso, "error marshalling jsonKey: %#v", err))
				return
			}
			v, err := v8go.JSONParse(info.Context(), string(resultBytes))
			if err != nil {
				resolver.Reject(newErrorValue(iso, "error jsonParse on result: %#v", err))
				return
			}

			resolver.Resolve(v)
		}()

		return resolver.GetPromise().Value
	}
}

func newErrorValue(iso *v8go.Isolate, format string, a ...interface{}) *v8go.Value {
	e, _ := v8go.NewValue(iso, fmt.Sprintf(format, a...))
	return e
}

func getAlgorithm(v *v8go.Value) (interface{}, string, error) {
	algorithm, err := v.AsObject() //object type
	if err != nil {
		return nil, "", fmt.Errorf("expected algorithm argument as Object type: %#v", err)
	}
	if !algorithm.Has("name") {
		return nil, "", fmt.Errorf("missing algorithm's name property")
	}
	algoName, err := algorithm.Get("name")
	if err != nil {
		return nil, "", fmt.Errorf("missing algorithm's name property:%#v", err)
	}
	if algoName.String() == "" {
		return nil, "", fmt.Errorf("missing algorithm's name property value")
	}
	res, err := algorithm.MarshalJSON()
	if err != nil {
		return nil, "", fmt.Errorf("error Marshalling algorithm:%#v", err)
	}

	var result interface{}
	switch algoName.String() {
	case string(RSA1_5), string(RSA_OAEP), string(RSA_OAEP_256):
		rsa := &RSAAlgoIn{}
		err = json.Unmarshal(res, rsa)
		if err != nil {
			return nil, "", fmt.Errorf("error UnMarshalling algorithm:%#v", err)
		}
		rsaout := &RSAAlgoOut{
			Name:           rsa.Name,
			PublicExponent: rsa.PublicExponent,
			Hash: HashAlgorithmIdentifier{
				Name: rsa.Hash,
			},
		}
		if rsa.ModulusLength == 0 {
			rsaout.ModulusLength = 2048
		}
		result = rsaout
	default:
		return nil, "", fmt.Errorf("unsupported algorithm - %s is not yet supported", algoName.String())
	}

	return result, algoName.String(), nil
}
