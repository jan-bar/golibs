package timer

// 获取系统运行时间
func GetUpTime() int64 {
	return getUpTime()
}

// 获取系统启动时间
func GetSysUpTime() int64 {
	return getSysUpTime()
}

// 更新系统启动时间
// 如果启动时间有变动则说明系统修改过时间
func UpSysUpTime() int64 {
	return upSysUpTime()
}
