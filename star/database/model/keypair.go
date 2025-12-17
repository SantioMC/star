package model

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/tubblelive/star/star/crypto"
)

type KeypairKind string

const (
	KeypairKindSigning KeypairKind = "signing"
)

type Keypair struct {
	BaseModel
	Name       string
	Kind       KeypairKind
	PublicKey  string `json:"-" encrypt:"true"`
	PrivateKey string `json:"-" encrypt:"true"`
}

func (k *Keypair) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(k.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key.(*ecdsa.PrivateKey), nil
}

func (k *Keypair) GetPublicKey() (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(k.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key.(*ecdsa.PublicKey), nil
}

func NewKeypair(name string, kind KeypairKind) (*Keypair, error) {
	keypair, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.MarshalPrivateKey(keypair)
	if err != nil {
		return nil, err
	}

	publicKey, err := crypto.MarshalPublicKey(&keypair.PublicKey)
	if err != nil {
		return nil, err
	}

	return &Keypair{
		BaseModel: BaseModel{
			ID: CreateTypedId("key", 12),
		},
		Name: name,
		Kind: kind,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}