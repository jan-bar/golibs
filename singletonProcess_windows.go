package golibs

import (
    "sync"
    "syscall"
    "unsafe"
)

var (
    modKernel32      = syscall.NewLazyDLL("kernel32.dll")
    procCreateMutexW = modKernel32.NewProc("CreateMutexW")

    singletonWin struct {
        h   syscall.Handle
        sync.RWMutex
    }
)

/*
通过windows信号量互斥原理实现单例运行
*/
func SingletonWin(name string) error {
    singletonWin.RLock()
    if singletonWin.h != 0 {
        singletonWin.RUnlock()
        return nil
    }
    singletonWin.RUnlock()

    singletonWin.Lock()
    defer singletonWin.Unlock()
    lpName, err := syscall.UTF16PtrFromString(name)
    if err != nil {
        return err
    }
    r0, _, e1 := syscall.Syscall(procCreateMutexW.Addr(), 3,
        uintptr(unsafe.Pointer(nil)), uintptr(0),
        uintptr(unsafe.Pointer(lpName)))
    singletonWin.h = syscall.Handle(r0)
    if e1 != 0 {
        if e1 == 0xb7 { // 已有互斥体错误码
            return ErrSingleton
        }
        return e1
    }
    return nil
}
