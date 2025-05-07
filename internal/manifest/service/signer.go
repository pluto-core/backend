package service

import (
	"crypto/ed25519"
	"encoding/base64"
)

type Signer interface {
	Sign(data []byte) (string, error)
	Verify(data []byte, sig string) (bool, error)
	GetPublicKey() (string, error)
}

type Ed25519Signer struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func NewEd25519Signer(privKey ed25519.PrivateKey, pubKey ed25519.PublicKey) *Ed25519Signer {
	return &Ed25519Signer{priv: privKey, pub: pubKey}
}

func (s *Ed25519Signer) GetPublicKey() (string, error) {
	pubKeyB64 := base64.StdEncoding.EncodeToString(s.pub)
	return pubKeyB64, nil
}

func (s *Ed25519Signer) Sign(data []byte) (string, error) {
	sig := ed25519.Sign(s.priv, data)
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (s *Ed25519Signer) Verify(data []byte, sigB64 string) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false, err
	}
	ok := ed25519.Verify(s.pub, data, sig)
	return ok, nil
}
