package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var ErrCipherTextBlockSize = errors.New("ciphertext block size is too short")

const nonceSize = 12

func EncryptAES(data string, key [32]byte) (string, error) {
	encrypted, err := EncryptBytesAES([]byte(data), key)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(encrypted), nil
}

func EncryptBytesAES(plainText []byte, key [32]byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	stream, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := stream.Seal(nil, nonce, plainText, nil)
	return append(nonce, cipherText...), nil
}

func DecryptAES(data string, key [32]byte) (string, error) {
	text, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	decrypted, err := DecryptBytesAES(text, key)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func DecryptBytesAES(cipherText []byte, key [32]byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	if len(cipherText) < nonceSize {
		return nil, ErrCipherTextBlockSize
	}

	stream, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	decoded, err := stream.Open(nil, cipherText[:nonceSize], cipherText[nonceSize:], nil)
	return decoded, err
}
