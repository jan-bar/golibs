package filelock

import (
    "errors"
    "os"
)

var ErrFileLock = errors.New("file is lock")

// 排它锁锁住文件
func Lock(f *os.File) error {
    return lock(f, WriteLock)
}

// 共享锁锁住文件
func RLock(f *os.File) error {
    return lock(f, ReadLock)
}

// 释放文件锁
func Unlock(f *os.File) error {
    return unlock(f)
}

type file struct {
    File *os.File
}

// 打开文件并带上锁
func LockOpenFile(name string, flag int, perm os.FileMode, lt lockType) (*file, error) {
    fr, err := os.OpenFile(name, flag, perm)
    if err != nil {
        return nil, err
    }
    if err = lock(fr, lt); err != nil {
        fr.Close()
        return nil, err
    }
    return &file{File: fr}, nil
}

// 释放锁并关闭文件
func (f *file) Close() error {
    err := unlock(f.File)
    if closeErr := f.File.Close(); err == nil {
        err = closeErr
    }
    return err
}
