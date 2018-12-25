package golibs

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

// https://github.com/polaris1119/myblog_article_code/blob/master/aes/aes.go
func AesEncrypt(origData, key, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	cryptEd := make([]byte, len(origData))
	blockMode.CryptBlocks(cryptEd, origData)
	return base64.StdEncoding.EncodeToString(cryptEd), nil
}

func AesDecrypt(cryptEd string, key, iv []byte) ([]byte, error) {
	byt, err := base64.StdEncoding.DecodeString(cryptEd)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(byt))
	blockMode.CryptBlocks(origData, byt)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}
