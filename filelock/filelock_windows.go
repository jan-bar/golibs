package filelock

import (
    "os"
    "syscall"
    "unsafe"
)

type lockType uint32

const (
    ReadLock  lockType = 0
    WriteLock lockType = 3 // LOCKFILE_FAIL_IMMEDIATELY | LOCKFILE_EXCLUSIVE_LOCK

    reserved = 0
    allBytes = ^uint32(0)
)

var (
    modKernel32      = syscall.NewLazyDLL("kernel32.dll")
    procLockFileEx   = modKernel32.NewProc("LockFileEx")
    procUnlockFileEx = modKernel32.NewProc("UnlockFileEx")
)

func lock(f *os.File, lt lockType) error {
    ol := new(syscall.Overlapped)
    r1, _, e1 := syscall.Syscall6(procLockFileEx.Addr(), 6, f.Fd(),
        uintptr(lt), uintptr(reserved), uintptr(allBytes),
        uintptr(allBytes), uintptr(unsafe.Pointer(ol)))
    if r1 == 0 {
        if e1 != 0 {
            if e1 == 0x21 { // 找到文件被锁错误码,返回自定义错误
                return ErrFileLock
            }
            return e1
        }
        return syscall.EINVAL
    }
    return nil
}

func unlock(f *os.File) error {
    ol := new(syscall.Overlapped)
    r1, _, e1 := syscall.Syscall6(procUnlockFileEx.Addr(), 5, f.Fd(),
        uintptr(reserved), uintptr(allBytes), uintptr(allBytes),
        uintptr(unsafe.Pointer(ol)), 0)
    if r1 == 0 {
        if e1 != 0 {
            return e1
        }
        return syscall.EINVAL
    }
    return nil
}
