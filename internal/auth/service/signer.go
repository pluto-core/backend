package service

import (
	_ "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Signer умеет подписывать данные и возвращать PEM-публичный ключ
type Signer interface {
	// Sign генерирует подпись для переданного payload
	Sign(payload []byte, exp time.Time, jti string) (string, error)
	// GetPublicKey возвращает PEM-публичный ключ
	GetPublicKey() (string, error)
}

type RSASigner struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSASigner поддерживает оба формата: PKCS#1 и PKCS#8.
func NewRSASigner(privateKeyB64, publicKeyB64 string) (*RSASigner, error) {
	// Декодируем base64 → PEM
	privPem, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, errors.New("не удалось декодировать PRIVATE_KEY_B64: " + err.Error())
	}
	block, _ := pem.Decode(privPem)
	if block == nil {
		return nil, errors.New("invalid PEM for private key")
	}

	var privKey *rsa.PrivateKey
	// Попробуем PKCS#1
	if parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		privKey = parsedKey
	} else {
		// Если PKCS#1 не прошёл, пробуем PKCS#8
		keyIfc, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, errors.New("не удалось распарсить RSA private key: " + err2.Error())
		}
		rsaKey, ok := keyIfc.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
		privKey = rsaKey
	}

	// Декодируем public key
	pubPem, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, errors.New("не удалось декодировать PUBLIC_KEY_B64: " + err.Error())
	}
	pubBlock, _ := pem.Decode(pubPem)
	if pubBlock == nil {
		return nil, errors.New("invalid PEM for public key")
	}
	pubIfc, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, errors.New("не удалось распарсить RSA public key: " + err.Error())
	}
	pubKey, ok := pubIfc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	return &RSASigner{
		privateKey: privKey,
		publicKey:  pubKey,
	}, nil
}

// Sign генерирует JWT с полным payload: issuer, sub=jti, iat, exp, fingerprint
func (s *RSASigner) Sign(fingerprint []byte, exp time.Time, jti string) (string, error) {
	claims := jwt.MapClaims{
		"iss": "pluto-auth",
		"jti": jti,
		"iat": time.Now().Unix(),
		"exp": exp.Unix(),
		"fp":  string(fingerprint), // положим JSON-фингерпринта в нативный claim "fp"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// GetPublicKey возвращает PEM-encoded public key
func (s *RSASigner) GetPublicKey() (string, error) {
	derBytes, _ := x509.MarshalPKIXPublicKey(s.publicKey)
	if derBytes == nil {
		return "", errors.New("не удалось маршалить public key")
	}
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}
	return string(pem.EncodeToMemory(pemBlock)), nil
}
