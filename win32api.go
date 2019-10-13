package golibs

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

const (
	StdOutputHandle     = 0xFFFFFFF5
	StdInputHandle      = 0xFFFFFFF6
	ForegroundBlue      = 0x01
	ForegroundGreen     = 0x02
	ForegroundRed       = 0x04
	ForegroundIntensity = 0x08
	BackgroundBlue      = 0x10
	BackgroundGreen     = 0x20
	BackgroundRed       = 0x40
	BackgroundIntensity = 0x80
	KeyUp               = 38 /* 向上按键的键值 */
	KeyDown             = 40 /* 向下按键的键值 */
	KeyLeft             = 37 /* 向左按键的键值 */
	KeyRight            = 39 /* 向右按键的键值 */
	SB_HORZ             = 0  /* 显示或隐藏窗体的标准的水平滚动条 */
	SB_VERT             = 1  /* 显示或隐藏窗体的标准的垂直滚动条 */
	SB_CTL              = 2  /* 显示或隐藏滚动条控制。参数hWnd必须是指向滚动条控制的句柄 */
	SB_BOTH             = 3  /* 显示或隐藏窗体的标准的水平或垂直滚动条 */
)

type (
	DWord uint32
	Word  uint16
	TChar rune

	Coord struct {
		X, Y int
	}

	SmallRect struct {
		Left, Top, Right, Bottom int16
	}

	ConsoleScreenBufferInfo struct {
		DwSize              Coord
		DwCursorPosition    Coord
		WAttributes         uint16
		SrWindow            SmallRect
		DwMaximumWindowSize Coord
	}

	ConsoleCursorInfo struct {
		dwSize   DWord
		bVisible DWord
	}

	Win32Api struct {
		hStdOutPut win.HWND /* 标准输出句柄 */
		hStdInPut  win.HWND /* 标准输入句柄 */
		cWindow    win.HWND /* 控制台窗体句柄 */
	}
)

var (
	fillConsoleOutputAttribute  *windows.LazyProc
	fillConsoleOutputCharacterW *windows.LazyProc
	getStdHandle                *windows.LazyProc
	getConsoleScreenBufferInfo  *windows.LazyProc
	setConsoleCursorPosition    *windows.LazyProc
	setConsoleTextAttribute     *windows.LazyProc
	setConsoleCursorInfo        *windows.LazyProc
	getConsoleWindow            *windows.LazyProc
	getKeyState                 *windows.LazyProc /* 处理win32api,获取键盘事件 */
	setWindowText               *windows.LazyProc
	showScrollBar               *windows.LazyProc
	getTickCount                *windows.LazyProc
	setLayeredWindowAttributes  *windows.LazyProc
	getConsoleMode              *windows.LazyProc
	setConsoleMode              *windows.LazyProc
	readConsoleInput            *windows.LazyProc
	mouseEvent            *windows.LazyProc
)

/* 将 Coord 转换为 Dword */
func mCoordToDword(c Coord) DWord {
	return DWord(c.Y)<<16 | DWord(c.X)
}

/**
* 初始化
* 主要加载win32api的方法
**/
func init() {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")

	fillConsoleOutputAttribute = kernel32.NewProc("FillConsoleOutputAttribute")
	fillConsoleOutputCharacterW = kernel32.NewProc("FillConsoleOutputCharacterW")
	getStdHandle = kernel32.NewProc("GetStdHandle")
	getConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	setConsoleCursorPosition = kernel32.NewProc("SetConsoleCursorPosition")
	setConsoleTextAttribute = kernel32.NewProc("SetConsoleTextAttribute")
	setConsoleCursorInfo = kernel32.NewProc("SetConsoleCursorInfo")
	getConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	getTickCount = kernel32.NewProc("GetTickCount")
	getConsoleMode = kernel32.NewProc("GetConsoleMode")
	setConsoleMode = kernel32.NewProc("SetConsoleMode")
	readConsoleInput = kernel32.NewProc("ReadConsoleInputA")

	user32 := windows.NewLazySystemDLL("user32.dll")

	mouseEvent = user32.NewProc("mouse_event")
	getKeyState = user32.NewProc("GetKeyState")
	setWindowText = user32.NewProc("SetWindowTextW")
	showScrollBar = user32.NewProc("ShowScrollBar")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
}

/**
* 新建win32api对象
**/
func NewWin32Api() *Win32Api {
	win32Api := new(Win32Api)                            /* 初始化对象 */
	win32Api.hStdOutPut = mGetStdHandle(StdOutputHandle) /* 得到标准输出句柄 */
	win32Api.hStdInPut = mGetStdHandle(StdInputHandle)   /* 得到标准输出句柄 */
	win32Api.cWindow = mGetConsoleWindow()               /* 得到控制台句柄 */
	return win32Api                                      /* 这2个变量全局有效 */
}

/**
* 设置固定区域内的文本属性，从指定的控制台屏幕缓冲区字符坐标开始。
**/
func mFillConsoleOutputAttribute(hConsoleOutput win.HWND, wAttribute uint16, nLength DWord, dwWriteCoord Coord) *DWord {
	var lpNumberOfAttrsWritten DWord
	ret, _, _ := syscall.Syscall6(fillConsoleOutputAttribute.Addr(), 5,
		uintptr(hConsoleOutput),
		uintptr(wAttribute),
		uintptr(nLength),
		uintptr(mCoordToDword(dwWriteCoord)),
		uintptr(unsafe.Pointer(&lpNumberOfAttrsWritten)),
		0)
	if ret == 0 {
		return nil
	} /* 返回上色成功的长度,上色失败返回-1 */
	return &lpNumberOfAttrsWritten
}

/**
* 在指定的坐标开始写入指定次数的字符到指定控制台屏幕缓冲区
* FillConsoleOutputCharacterW (Unicode)
* FillConsoleOutputCharacterA (ANSI)
* FillConsoleOutputCharacter  (Default)
**/
func mFillConsoleOutputCharacter(hConsoleOutput win.HWND, cCharacter TChar, nLength DWord, dwWriteCoord Coord) *DWord {
	var lpNumberOfAttrsWritten DWord
	ret, _, _ := syscall.Syscall6(fillConsoleOutputCharacterW.Addr(), 5,
		uintptr(hConsoleOutput),
		uintptr(cCharacter),
		uintptr(nLength),
		uintptr(mCoordToDword(dwWriteCoord)),
		uintptr(unsafe.Pointer(&lpNumberOfAttrsWritten)),
		0)
	if ret == 0 {
		return nil
	}
	return &lpNumberOfAttrsWritten
}

/**
* 它用于从一个特定的标准设备（标准输入、标准输出或标准错误）
* 中取得一个句柄（用来标识不同设备的数值）
**/
func mGetStdHandle(nStdHandle DWord) win.HWND {
	ret, _, _ := syscall.Syscall(getStdHandle.Addr(), 1,
		uintptr(nStdHandle),
		0,
		0)
	return win.HWND(ret)
}

/**
* 用于检索指定的控制台屏幕缓冲区的信息
**/
func mGetConsoleScreenBufferInfo(hConsoleOutput win.HWND) *ConsoleScreenBufferInfo {
	var CsBi ConsoleScreenBufferInfo
	ret, _, _ := syscall.Syscall(getConsoleScreenBufferInfo.Addr(), 2,
		uintptr(hConsoleOutput),
		uintptr(unsafe.Pointer(&CsBi)),
		0)
	if ret == 0 {
		return nil
	}
	return &CsBi
}

/**
* 是API中定位光标位置的函数
**/
func mSetConsoleCursorPosition(hConsoleOutput win.HWND, dwCursorPosition Coord) bool {
	ret, _, _ := syscall.Syscall(setConsoleCursorPosition.Addr(), 2,
		uintptr(hConsoleOutput),
		uintptr(mCoordToDword(dwCursorPosition)),
		0)
	return ret != 0
}

/**
* 设置控制台窗口字体颜色和背景色的计算机函数
* 私有,不对外开发
**/
func mSetConsoleTextAttribute(hConsoleOutput win.HWND, wAttributes int) bool {
	ret, _, _ := syscall.Syscall(setConsoleTextAttribute.Addr(), 2,
		uintptr(hConsoleOutput),
		uintptr(uint16(wAttributes)),
		0)
	return ret != 0
}

/**
* 设置光标属性
**/
func SetConsoleCursorInfo(hConsoleOutput win.HWND, lpConsoleCursorInfo ConsoleCursorInfo) bool {
	ret, _, _ := syscall.Syscall(setConsoleCursorInfo.Addr(), 2,
		uintptr(hConsoleOutput),
		uintptr(unsafe.Pointer(&lpConsoleCursorInfo)),
		0)
	return ret != 0
}

/**
* 显示或隐藏光标
* 传true则显示光标
* 传false则隐藏光标
**/
func (api *Win32Api) ShowHideCursor(show bool) {
	var bVisible DWord = 0
	if show {
		bVisible = 1
	}
	SetConsoleCursorInfo(api.hStdOutPut, ConsoleCursorInfo{1, bVisible})
}

/**
* 清屏函数
**/
func (api *Win32Api) Clear() {
	coordScreen := Coord{0, 0}
	csbi := mGetConsoleScreenBufferInfo(api.hStdOutPut)
	dwConSize := DWord(csbi.DwSize.X * csbi.DwSize.Y)
	mFillConsoleOutputCharacter(api.hStdOutPut, TChar(' '), dwConSize, coordScreen)
	csbi = mGetConsoleScreenBufferInfo(api.hStdOutPut)
	mFillConsoleOutputAttribute(api.hStdOutPut, csbi.WAttributes, dwConSize, coordScreen)
	mSetConsoleCursorPosition(api.hStdOutPut, coordScreen)
}

/**
* 光标定位到某个位置
**/
func (api *Win32Api) GotoXY(x, y int) {
	mSetConsoleCursorPosition(api.hStdOutPut, Coord{x, y})
}

/**
* 设置打印颜色
**/
func (api *Win32Api) TextBackground(color int) {
	mSetConsoleTextAttribute(api.hStdOutPut, color)
}

/**
* 获取标准输入方式
**/
func (api *Win32Api) GetConsoleMode() *DWord {
	var mode DWord
	ret, _, _ := syscall.Syscall(getConsoleMode.Addr(), 2,
		uintptr(api.hStdInPut),
		uintptr(unsafe.Pointer(&mode)), 0)
	if ret != 0 {
		return nil
	}
	return &mode
}

/**
* 设置标准输入方式
* 具体参数看微软文档,这里懒得写成预定义了
**/
func (api *Win32Api) SetConsoleMode(mode DWord) error {
	_, _, err := syscall.Syscall(setConsoleMode.Addr(), 2,
		uintptr(api.hStdInPut),
		uintptr(mode), 0)
	return err
}

/**
* 获取一个按键值,在该按键按下松开时才返回键值
**/
func (api *Win32Api) ReadOneKey() byte {
	var (
		lpBuffer = make([]byte, 20)
		lpNumber DWord
		keyVal   byte = 255
		ret      uintptr
		trap     = uintptr(api.hStdInPut)
		a1       = uintptr(unsafe.Pointer(&lpBuffer[0]))
		a4       = uintptr(unsafe.Pointer(&lpNumber))
	)
	for {
		ret, _, _ = syscall.Syscall6(readConsoleInput.Addr(), 2,
			trap, a1, 1, a4, 0, 0)
		if ret != 0 && lpBuffer[0] == 1 { // 按键事件
			if lpBuffer[4] == 1 && keyVal == 255 {
				keyVal = lpBuffer[10] // 按下,且为首次按键
			} else if lpBuffer[4] == 0 && keyVal == lpBuffer[10] {
				break // 该按键松开
			}
		}
	}
	return keyVal
}

/**
* 等待<上,下,左,右>
* 这4个按键按下并松开
* 一旦满足则返回键值
**/
func WaitKeyBoard() (keyVal int32) {
	for keyVal == 0 {
		switch {
		case GetKeyState(KeyUp):
			keyVal = KeyUp
		case GetKeyState(KeyDown):
			keyVal = KeyDown
		case GetKeyState(KeyLeft):
			keyVal = KeyLeft
		case GetKeyState(KeyRight):
			keyVal = KeyRight
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}

	for GetKeyState(keyVal) {
		time.Sleep(time.Millisecond * 100)
	} /* 松开才返回,避免判断按键重复按下 */
	return
}

/**
* win32编程,获取键盘输入
* 传入键值
* 按下返回true,松开返回false
**/
func GetKeyState(nVirtKey int32) bool {
	ret, _, _ := syscall.Syscall(getKeyState.Addr(), 1, uintptr(nVirtKey), 0, 0)
	return int16(ret) < 0
}

/**
* 鼠标事件
**/
const (
	MOUSEEVENTF_ABSOLUTE   DWord = 0x8000
	MOUSEEVENTF_LEFTDOWN   DWord = 0x0002
	MOUSEEVENTF_LEFTUP     DWord = 0x0004
	MOUSEEVENTF_MIDDLEDOWN DWord = 0x0020
	MOUSEEVENTF_MIDDLEUP   DWord = 0x0040
	MOUSEEVENTF_MOVE       DWord = 0x0001
	MOUSEEVENTF_RIGHTDOWN  DWord = 0x0008
	MOUSEEVENTF_RIGHTUP    DWord = 0x0010
	MOUSEEVENTF_WHEEL      DWord = 0x0800
	MOUSEEVENTF_XDOWN      DWord = 0x0080
	MOUSEEVENTF_XUP        DWord = 0x0100
	MOUSEEVENTF_HWHEEL     DWord = 0x01000
)

func MouseEvent(dwFlags DWord, args ...DWord) {
	var dx, dy, dwData, dwExtraInfo DWord
	switch len(args) {
	case 4:
		dwExtraInfo = args[3]
		fallthrough
	case 3:
		dwData = args[2]
		fallthrough
	case 2:
		dy = args[1]
		fallthrough
	case 1:
		dx = args[0]
	}
	syscall.Syscall6(mouseEvent.Addr(), 5,
		uintptr(dwFlags),
		uintptr(dx),
		uintptr(dy),
		uintptr(dwData),
		uintptr(dwExtraInfo),
		0)
}

/**
* 获取当前控制台句柄
* 私有,不对外开发
**/
func mGetConsoleWindow() win.HWND {
	ret, _, _ := syscall.Syscall(getConsoleWindow.Addr(), 0, 0, 0, 0)
	return win.HWND(ret)
}

/**
* 设置窗体左上角文字
* SetWindowTextW (Unicode) and SetWindowTextA (ANSI)
**/
func (api *Win32Api) SetWindowText(text string) bool {
	str := win.SysAllocString(text) /* 申请字符串 */
	ret, _, _ := syscall.Syscall(setWindowText.Addr(), 2,
		uintptr(api.cWindow),
		uintptr(unsafe.Pointer(str)),
		0)
	win.SysFreeString(str) /* api说明中,用完就释放 */
	return ret != 0
}

/**
* 居中显示窗体
* 并设置宽高
**/
func (api *Win32Api) CenterWindowOnScreen(w, h int32) {
	xLeft := (win.GetSystemMetrics(win.SM_CXFULLSCREEN) - w) / 2
	yTop := (win.GetSystemMetrics(win.SM_CYFULLSCREEN) - h) / 2
	win.SetWindowPos(api.cWindow, win.HWND_TOPMOST, xLeft, yTop, w, h, win.SWP_NOZORDER)
	win.SetWindowPos(api.cWindow, win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
}

/**
* 显示或隐藏滚动条
* wBar看定义
* bShow=0隐藏,bShow=1显示
* 控制滚动条,还有EnableScrollBar
**/
func (api *Win32Api) ShowScrollBar(wBar, bShow DWord) bool {
	ret, _, _ := syscall.Syscall(showScrollBar.Addr(), 3,
		uintptr(api.cWindow),
		uintptr(wBar),
		uintptr(bShow))
	return ret != 0
}

func RGB(r, g, b byte) DWord {
	return DWord(r) | DWord(g)<<8 | DWord(b)<<16
}

/**
* crKey   : RGB(r,g,b)
* bAlpha  : 0~255
* dwFlags : LWA_ALPHA=0x2,LWA_COLORKEY=0x1,LWA_ALPHA | LWA_COLORKEY
**/
func (api *Win32Api) SetLayeredWindowAttributes(crKey, bAlpha, dwFlags DWord) bool {
	ret, _, _ := syscall.Syscall6(setLayeredWindowAttributes.Addr(), 4,
		uintptr(api.cWindow),
		uintptr(crKey),
		uintptr(bAlpha),
		uintptr(dwFlags), 0, 0)
	return ret != 0
}

/*
* 获取电脑开机到现在的秒数
 */
func GetTickCount() int64 {
	ret, _, _ := syscall.Syscall(getTickCount.Addr(), 0, 0, 0, 0)
	return int64(ret)
}
