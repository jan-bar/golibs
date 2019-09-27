package timer

import (
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

var (
	// Library
	libKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	// Functions
	getTickCount = libKernel32.NewProc("GetTickCount")

	sysUpTime = struct {
		time int64 // 系统启动时间戳
		sync.RWMutex
	}{
		time: time.Now().Unix() - getUpTime(), // 初始时更新系统启动时间
	}
)

func upSysUpTime() int64 {
	sysUpTime.Lock()
	defer sysUpTime.Unlock()
	sysUpTime.time = time.Now().Unix() - getUpTime()
	return sysUpTime.time
}

func getUpTime() int64 {
	ret, _, _ := syscall.Syscall(getTickCount.Addr(), 0, 0, 0, 0)
	return int64(ret) / 1000
}

func getSysUpTime() int64 {
	sysUpTime.RLock()
	defer sysUpTime.RUnlock()
	return sysUpTime.time
}

// 获取时区偏移小时值
func getZoneHour() (int, error) {
	zone := new(windows.Timezoneinformation)
	_, err := windows.GetTimeZoneInformation(zone)
	if err != nil {
		return 0, err
	}
	return int(zone.Bias / -60), nil
}
