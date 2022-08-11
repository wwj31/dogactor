package tools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func genRSAKey() error {
	private, _ := rsa.GenerateKey(rand.Reader, 512)
	bytes := x509.MarshalPKCS1PrivateKey(private)
	block := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   bytes,
	}

	privFile, err := os.Create("private.pem")
	if err != nil {
		return err
	}

	if err = pem.Encode(privFile, &block); err != nil {
		return err
	}

	_ = privFile.Close()

	bytes, err = x509.MarshalPKIXPublicKey(&private.PublicKey)
	if err != nil {
		return err
	}

	block = pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   bytes,
	}

	pubFile, err := os.Create("public.pem")
	if err != nil {
		return err
	}

	if err = pem.Encode(pubFile, &block); err != nil {
		return err
	}

	_ = pubFile.Close()
	return nil
}

// RSAPublicKeyEncrypt 公钥加密
func RSAPublicKeyEncrypt(key, data []byte) ([]byte, error) {
	pubKey, err := publicKey(key)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
}

// RSAPrivateDecrypt 私钥解密
func RSAPrivateDecrypt(key, ciphertext []byte) ([]byte, error) {
	pubKey, err := privateKey(key)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, pubKey, ciphertext)
}

func publicKey(key []byte) (*rsa.PublicKey, error) {
	block, rest := pem.Decode(key)
	fmt.Println(string(rest))
	if block == nil {
		return nil, fmt.Errorf("pem decode failed key:%v", string(key))
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pubKey := pub.(*rsa.PublicKey)

	return pubKey, nil
}

func privateKey(key []byte) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode(key)
	fmt.Println(string(rest))
	if block == nil {
		return nil, fmt.Errorf("pem decode failed key:%v", string(key))
	}
	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priKey, nil
}
