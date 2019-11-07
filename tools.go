package golibs

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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

// 更方便易用的exec.Command
func Command(name, args string) (*exec.Cmd, error) {
	if filepath.Base(name) == name {
		lp, err := exec.LookPath(name)
		if err != nil {
			return nil, err
		}
		name = lp
	}
	return &exec.Cmd{
		Path:        name,
		SysProcAttr: &syscall.SysProcAttr{CmdLine: name + " " + args},
	}, nil
}

// 判断文件或目录存在且类型正确
func IsFilePathExists(path string, isFile bool) error {
	if path == "" {
		return errors.New("path is nil")
	}
	f, err := os.Stat(path)
	if err != nil {
		return err
	}
	if isFile != f.IsDir() {
		return nil
	}
	if isFile {
		return errors.New(path + " not file")
	}
	return errors.New(path + " not dir")
}
