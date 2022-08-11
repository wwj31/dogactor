package tools

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestRSAGenKey(t *testing.T) {
	err := GenRSAKey()
	assert.Nil(t, err)
	TestRSACrypt(t)
}

func TestRSACrypt(t *testing.T) {
	file, err := os.Open("private.pem")
	assert.Nil(t, err)

	private, err := io.ReadAll(file)
	assert.Nil(t, err)
	_ = file.Close()

	file, err = os.Open("public.pem")
	assert.Nil(t, err)

	public, err := io.ReadAll(file)
	assert.Nil(t, err)
	_ = file.Close()

	ciphertext, err := RSAPublicKeyEncrypt(public, []byte("hello 世界!"))
	assert.Nil(t, err)

	fmt.Println("ciphertext:", string(ciphertext))
	fmt.Println("ciphertext base64:", base64.StdEncoding.EncodeToString(ciphertext))

	data, err := RSAPrivateDecrypt(private, ciphertext)
	assert.Nil(t, err)

	fmt.Println("data:", string(data))
}
