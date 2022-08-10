package tools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func genPrivateKey() error {
	private, _ := rsa.GenerateKey(rand.Reader, 1024)
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
