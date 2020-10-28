package filelock

import (
    "os"
    "syscall"
)

type lockType int

const (
    ReadLock  lockType = syscall.LOCK_SH
    WriteLock lockType = syscall.LOCK_EX
)

func lock(f *os.File, lt lockType) error {
    err := syscall.Flock(int(f.Fd()), int(lt)|syscall.LOCK_NB)
    if err != nil {
        if errNo, ok := err.(syscall.Errno); ok && errNo == 0xb {
            return ErrFileLock // 找到文件被锁错误码,返回自定义错误
        }
    }
    return err
}

func unlock(f *os.File) error {
    return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
