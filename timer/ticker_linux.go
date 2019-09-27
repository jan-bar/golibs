package timer

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var sysUpTime = struct {
	time int64 // 系统启动时间戳
	sync.RWMutex
}{
	time: time.Now().Unix() - getUpTime(),
}

func upSysUpTime() int64 {
	sysUpTime.Lock()
	defer sysUpTime.Unlock()
	sysUpTime.time = time.Now().Unix() - getUpTime()
	return sysUpTime.time
}

func getUpTime() int64 {
	info := new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(info)
	if err != nil {
		return 0
	}
	return info.Uptime
}

func getSysUpTime() int64 {
	sysUpTime.RLock()
	defer sysUpTime.RUnlock()
	return sysUpTime.time
}

// 获取时区偏移小时值
func getZoneHour() (int, error) {
	cmd := exec.Command("date", "+%:::z")
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out.String()))
}
