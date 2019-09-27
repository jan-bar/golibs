package golibs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

func GenRsaKey(bits int, private, public string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits) // 生成私钥文件
	if err != nil {
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	privFile, err := os.Create(private)
	if err != nil {
		return err
	}
	defer privFile.Close()

	err = pem.Encode(privFile, block)
	if err != nil {
		return err
	}

	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey) // 生成公钥文件
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubFile, err := os.Create(public)
	if err != nil {
		return err
	}
	defer pubFile.Close()
	err = pem.Encode(pubFile, block)
	if err != nil {
		return err
	}
	return nil
}

// 加密
func RsaEncrypt(origData, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(cipherText, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, cipherText)
}
