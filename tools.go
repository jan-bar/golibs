package golibs

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"unsafe"
)

func Md5sum(s string, isFile bool) (string, error) {
	h := md5.New()
	if isFile {
		fr, err := os.Open(s)
		if err != nil {
			return "", err
		}
		defer fr.Close()
		if _, err = io.Copy(h, fr); err != nil {
			return "", err
		}
	} else {
		if _, err := h.Write([]byte(s)); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// 由于共用内存,转换后的[]byte不可写
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// 由于共用内存,[]byte改变时string也会变
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
