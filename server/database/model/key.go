package model

import (
	"crypto/rsa"

	"github.com/go-jose/go-jose/v4"
)

type SigningKey struct {
	Id               string
	SigningAlgorithm jose.SignatureAlgorithm
	PrivateKey       *rsa.PrivateKey
}

type PublicKey struct {
	SigningKey
}

func (s *SigningKey) SignatureAlgorithm() jose.SignatureAlgorithm {
	return s.SigningAlgorithm
}

func (s *SigningKey) Key() any {
	return s.PrivateKey
}

func (s *SigningKey) ID() string {
	return s.Id
}

func (s *PublicKey) ID() string {
	return s.Id
}

func (s *PublicKey) Algorithm() jose.SignatureAlgorithm {
	return s.SigningAlgorithm
}

func (s *PublicKey) Use() string {
	return "sig"
}

func (s *PublicKey) Key() any {
	return &s.PrivateKey.PublicKey
}