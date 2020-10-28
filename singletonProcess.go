package golibs

import (
    "errors"
    "net"
    "os"
    "strconv"
    "strings"
    "sync"

    "github.com/jan-bar/golibs/filelock"
)

var (
    singleton struct {
        f   *os.File
        sync.RWMutex
    }
    ErrSingleton = errors.New("singleton process")
)

/*
通过加锁文件实现进程单例
返回ErrSingleton表示文件被锁,已有进程在运行
*/
func SingletonFile(path string) error {
    singleton.RLock()
    if singleton.f != nil {
        singleton.RUnlock()
        return nil
    }
    singleton.RUnlock()

    singleton.Lock()
    defer singleton.Unlock()
    _, err := os.Stat(path)
    if err == nil { // 文件存在则打开
        singleton.f, err = os.Open(path)
        if err != nil {
            return err
        }
    } else { // 文件不存在则创建,并写入pid
        singleton.f, err = os.Create(path)
        if err != nil {
            return err
        }
        singleton.f.Write([]byte(strconv.Itoa(os.Getpid())))
    }
    if err = filelock.Lock(singleton.f); err == filelock.ErrFileLock {
        return ErrSingleton
    }
    return err
}

/*
通过监听TCP端口实现进程单例
*/
func SingletonTcp(port int) error {
    addr := "127.0.0.1:" + strconv.Itoa(port)
    l, err := net.Listen("tcp", addr)
    if err != nil {
        if strings.Contains(err.Error(), addr) {
            return ErrSingleton // 已有端口在监听,表示已有进程在运行
        }
        return err
    }
    go func() {
        for { // 启动协程监听端口
            l.Accept()
        }
    }()
    return nil
}
