package jwt

import (
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/pkg/errors"

	hivecrypto "github.com/iotaledger/hive.go/crypto"
	"github.com/iotaledger/hive.go/ioutils"
)

var (
	ErrPrivKeyInvalid = errors.New("invalid private key")
	ErrNoPrivKeyFound = errors.New("no private key found")
)

// ParseEd25519PrivateKeyFromString parses an Ed25519 private key from a hex encoded string.
func ParseEd25519PrivateKeyFromString(identityPrivKey string) (ed25519.PrivateKey, error) {
	if identityPrivKey == "" {
		return nil, ErrNoPrivKeyFound
	}

	hivePrivKey, err := hivecrypto.ParseEd25519PrivateKeyFromString(identityPrivKey)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", ErrPrivKeyInvalid)
	}

	return ed25519.PrivateKey(hivePrivKey), nil
}

// ReadEd25519PrivateKeyFromPEMFile reads an Ed25519 private key from a file with PEM format.
func ReadEd25519PrivateKeyFromPEMFile(filepath string) (ed25519.PrivateKey, error) {

	pemPrivateBlockBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	pemPrivateBlock, _ := pem.Decode(pemPrivateBlockBytes)
	if pemPrivateBlock == nil {
		return nil, fmt.Errorf("unable to decode private key: %w", err)
	}

	stdCryptoPrvKey, err := x509.ParsePKCS8PrivateKey(pemPrivateBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	stdPrvKey, ok := stdCryptoPrvKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unable to type assert private key: %w", err)
	}

	return stdPrvKey, nil
}

// WriteEd25519PrivateKeyToPEMFile stores an Ed25519 private key to a file with PEM format.
func WriteEd25519PrivateKeyToPEMFile(filepath string, privateKey ed25519.PrivateKey) error {

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("unable to mashal private key: %w", err)
	}

	pemPrivateBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	var pemBuffer bytes.Buffer
	if err := pem.Encode(&pemBuffer, pemPrivateBlock); err != nil {
		return fmt.Errorf("unable to encode private key: %w", err)
	}

	if err := ioutils.WriteToFile(filepath, pemBuffer.Bytes(), 0660); err != nil {
		return fmt.Errorf("unable to write private key: %w", err)
	}

	return nil
}

// LoadOrCreateIdentityPrivateKey loads an existing Ed25519 based identity private key
// or creates a new one and stores it as a PEM file in the identityFilePath.
func LoadOrCreateIdentityPrivateKey(identityFilePath string, identityPrivKey string) (ed25519.PrivateKey, bool, error) {

	privKeyFromConfig, err := ParseEd25519PrivateKeyFromString(identityPrivKey)
	if err != nil {
		if errors.Is(err, ErrPrivKeyInvalid) {
			return nil, false, errors.New("configuration contains an invalid private key")
		}

		if !errors.Is(err, ErrNoPrivKeyFound) {
			return nil, false, fmt.Errorf("unable to parse private key from config: %w", err)
		}
	}

	_, err = os.Stat(identityFilePath)
	switch {
	case err == nil || os.IsExist(err):
		// private key already exists, load and return it
		privKey, err := ReadEd25519PrivateKeyFromPEMFile(identityFilePath)
		if err != nil {
			return nil, false, fmt.Errorf("unable to load Ed25519 private key for identity: %w", err)
		}

		if privKeyFromConfig != nil && !privKeyFromConfig.Equal(privKey) {
			return nil, false, fmt.Errorf("stored Ed25519 private key (%s) for identity doesn't match private key in config (%s)", hex.EncodeToString(privKey[:]), hex.EncodeToString(privKeyFromConfig[:]))
		}

		return privKey, false, nil

	case os.IsNotExist(err):
		var privKey ed25519.PrivateKey

		if privKeyFromConfig != nil {
			privKey = privKeyFromConfig
		} else {
			// private key does not exist, create a new one
			_, privKey, err = ed25519.GenerateKey(nil)
			if err != nil {
				return nil, false, fmt.Errorf("unable to generate Ed25519 private key for identity: %w", err)
			}
		}
		if err := WriteEd25519PrivateKeyToPEMFile(identityFilePath, privKey); err != nil {
			return nil, false, fmt.Errorf("unable to store private key file for identity: %w", err)
		}
		return privKey, true, nil

	default:
		return nil, false, fmt.Errorf("unable to check private key file for identity (%s): %w", identityFilePath, err)
	}
}
