package golibs

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"reflect"
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

	KeyUp      int32 = win.VK_UP      /* 向上按键的键值 */
	KeyDown    int32 = win.VK_DOWN    /* 向下按键的键值 */
	KeyLeft    int32 = win.VK_LEFT    /* 向左按键的键值 */
	KeyRight   int32 = win.VK_RIGHT   /* 向右按键的键值 */
	MouseLeft  int32 = win.VK_LBUTTON /* 鼠标左键 */
	MouseRight int32 = win.VK_RBUTTON /* 鼠标右键 */
	MouseMid   int32 = win.VK_MBUTTON /* 鼠标中键 */

	SB_HORZ = 0 /* 显示或隐藏窗体的标准的水平滚动条 */
	SB_VERT = 1 /* 显示或隐藏窗体的标准的垂直滚动条 */
	SB_CTL  = 2 /* 显示或隐藏滚动条控制。参数hWnd必须是指向滚动条控制的句柄 */
	SB_BOTH = 3 /* 显示或隐藏窗体的标准的水平或垂直滚动条 */
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
	setWindowText               *windows.LazyProc
	showScrollBar               *windows.LazyProc
	getTickCount                *windows.LazyProc
	setLayeredWindowAttributes  *windows.LazyProc
	getConsoleMode              *windows.LazyProc
	setConsoleMode              *windows.LazyProc
	readConsoleInput            *windows.LazyProc
	mouseEvent                  *windows.LazyProc
	getCursorPos                *windows.LazyProc
	createToolHelp32Snapshot    *windows.LazyProc
	process32First              *windows.LazyProc
	process32Next               *windows.LazyProc
	readProcessMemory           *windows.LazyProc
	openProcess                 *windows.LazyProc
	writeProcessMemory          *windows.LazyProc
	virtualQueryEx              *windows.LazyProc
	getSystemInfo               *windows.LazyProc
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
	createToolHelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	process32First = kernel32.NewProc("Process32First")
	process32Next = kernel32.NewProc("Process32Next")
	openProcess = kernel32.NewProc("OpenProcess")
	readProcessMemory = kernel32.NewProc("ReadProcessMemory")
	writeProcessMemory = kernel32.NewProc("WriteProcessMemory")
	virtualQueryEx = kernel32.NewProc("VirtualQueryEx")
	getSystemInfo = kernel32.NewProc("GetSystemInfo")

	user32 := windows.NewLazySystemDLL("user32.dll")

	mouseEvent = user32.NewProc("mouse_event")
	setWindowText = user32.NewProc("SetWindowTextW")
	showScrollBar = user32.NewProc("ShowScrollBar")
	getCursorPos = user32.NewProc("GetCursorPos")
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
	return win32Api
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
	if ret == 0 {
		return nil
	}
	return &mode
}

/**
* 设置标准输入方式
* 具体参数看微软文档,这里懒得写成预定义了
**/
const (
	EnableQuickEditMode DWord = 0x0040 // 快速编辑模式
	EnableInsertMode    DWord = 0x0020 // 插入模式
	EnableMouseInput    DWord = 0x0010 // 鼠标输入
	EnableExtendedFlags DWord = 0x0080
	EnableWindowInput   DWord = 0x0008
)

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
		sTime    = time.Millisecond * 100
		trap     = uintptr(api.hStdInPut)
		a1       = uintptr(unsafe.Pointer(&lpBuffer[0]))
		a4       = uintptr(unsafe.Pointer(&lpNumber))
		read     = readConsoleInput.Addr()
	)
	for {
		ret, _, _ = syscall.Syscall6(read, 2, trap, a1, 1, a4, 0, 0)
		if ret != 0 && lpBuffer[0] == 1 { // 按键事件
			if lpBuffer[4] == 1 && keyVal == 255 {
				keyVal = lpBuffer[10] // 按下,且为首次按键
			} else if lpBuffer[4] == 0 && keyVal == lpBuffer[10] {
				break // 该按键松开
			}
			time.Sleep(sTime)
		}
	}
	return keyVal
}

/**
* 等待对应按键按下并松开,返回对应键值
**/
func WaitKeyBoard(key ...int32) int32 {
	keyVal, sTime := int32(0), time.Millisecond*100
	for {
		for _, v := range key {
			if win.GetKeyState(v) < 0 {
				keyVal = v
				goto waitUp
			}
		}
		time.Sleep(sTime)
	}
waitUp:
	for win.GetKeyState(keyVal) < 0 {
		time.Sleep(sTime)
	} /* 松开才返回,避免判断按键重复按下 */
	return keyVal
}

/**
* 鼠标事件
**/
//const (
//	MOUSEEVENTF_ABSOLUTE   DWord = 0x8000
//	MOUSEEVENTF_LEFTDOWN   DWord = 0x0002
//	MOUSEEVENTF_LEFTUP     DWord = 0x0004
//	MOUSEEVENTF_MIDDLEDOWN DWord = 0x0020
//	MOUSEEVENTF_MIDDLEUP   DWord = 0x0040
//	MOUSEEVENTF_MOVE       DWord = 0x0001
//	MOUSEEVENTF_RIGHTDOWN  DWord = 0x0008
//	MOUSEEVENTF_RIGHTUP    DWord = 0x0010
//	MOUSEEVENTF_WHEEL      DWord = 0x0800
//	MOUSEEVENTF_XDOWN      DWord = 0x0080
//	MOUSEEVENTF_XUP        DWord = 0x0100
//	MOUSEEVENTF_HWHEEL     DWord = 0x01000
//)

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
* 获取当前鼠标坐标
* mouse: 当按下鼠标左键、中键、右键时返回键值和鼠标坐标
**/
func GetCursorPos(pt *win.POINT, mouse *int32) bool {
	if mouse != nil {
		*mouse = WaitKeyBoard(MouseLeft, MouseRight, MouseMid)
	}
	ret, _, _ := syscall.Syscall(getCursorPos.Addr(), 1,
		uintptr(unsafe.Pointer(pt)), 0, 0)
	return ret != 0
}

/**
* pt1鼠标相对控制台坐标
* pt2鼠标全局坐标
**/
func (api *Win32Api) GetCursorPos(pt1 *win.POINT, pt2 *win.POINT, mouse *int32) bool {
	if !GetCursorPos(pt1, mouse) {
		return false
	}
	pt2.X, pt2.Y = pt1.X, pt1.Y
	return win.ScreenToClient(api.cWindow, pt1)
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

/*----------------------------------------------------------------------------*/
// 截屏, name!="" 则会保存文件,args代表x,y,w,h,如果参数不全则截取全屏
func CaptureRect(name string, args ...int) (image.Image, error) {
	var x, y, w, h int32
	if len(args) != 4 {
		w = win.GetSystemMetrics(win.SM_CXFULLSCREEN)
		h = win.GetSystemMetrics(win.SM_CYFULLSCREEN)
	} else {
		x, y, w, h = int32(args[0]), int32(args[1]), int32(args[2]), int32(args[3])
	}
	hDC := win.GetDC(0)
	if hDC == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", win.GetLastError())
	}
	defer win.ReleaseDC(0, hDC)

	mhDC := win.CreateCompatibleDC(hDC)
	if mhDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", win.GetLastError())
	}
	defer win.DeleteDC(mhDC)

	bt := new(win.BITMAPINFO)
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = w
	bt.BmiHeader.BiHeight = -h
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = win.BI_RGB

	ptr := unsafe.Pointer(uintptr(0))
	mhBmp := win.CreateDIBSection(mhDC, (*win.BITMAPINFOHEADER)(unsafe.Pointer(bt)), win.DIB_RGB_COLORS, &ptr, 0, 0)
	if mhBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", win.GetLastError())
	}
	if win.GpStatus(mhBmp) == win.InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer win.DeleteObject(win.HGDIOBJ(mhBmp))

	obj := win.SelectObject(mhDC, win.HGDIOBJ(mhBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", win.GetLastError())
	}
	if obj == 0xffffffff { // GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", win.GetLastError())
	}
	defer win.DeleteObject(obj)

	//Note:BitBlt contains bad error handling, we will just assume it works and if it doesn't it will panic :x
	win.BitBlt(mhDC, 0, 0, w, h, hDC, x, y, win.SRCCOPY)

	var (
		slice  []uint8
		ww, hh = int(w), int(h)
		lSlice = ww * hh * 4
	)
	hDrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hDrp.Data = uintptr(ptr)
	hDrp.Len, hDrp.Cap = lSlice, lSlice

	imageBytes := make([]uint8, lSlice)
	for i := 0; i < lSlice; i += 4 {
		imageBytes[i], imageBytes[i+2], imageBytes[i+1], imageBytes[i+3] = slice[i+2], slice[i], slice[i+1], slice[i+3]
	}

	img := &image.RGBA{Pix: imageBytes, Stride: 4 * ww, Rect: image.Rect(0, 0, ww, hh)}
	if name != "" { // 保存图片
		fw, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		defer fw.Close()
		if err = png.Encode(fw, img); err != nil {
			return nil, err
		}
	}
	return img, nil
}

/*----------------------------------------------------------------------------*/
const (
	TH32CsINHERIT      = 0x80000000
	TH32CsSNAPHEAPLIST = 0x00000001
	TH32CsSNAPMODULE   = 0x00000008
	TH32CsSNAPMODULE32 = 0x00000010
	TH32CsSNAPPROCESS  = 0x00000002
	TH32CsSNAPTHREAD   = 0x00000004
	InvalidHandleValue = 0xFFFFFFFF

	MaxPath = 260 // 可以适当减小
)

type ProcessEntry32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   *uint32
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      uint32
	DwFlags             uint32
	SzExeFile           [MaxPath]byte
}

func (p *ProcessEntry32) GetFileName() string {
	i := 0 // 因为C语言'\0'结束,所以go里面得截成string
	for ; i < MaxPath; i++ {
		if p.SzExeFile[i] == 0 {
			break
		}
	}
	return BytesToString(p.SzExeFile[:i])
}

func CreateToolHelp32Snapshot(dwFlags, th32ProcessID DWord) win.HANDLE {
	ret, _, _ := syscall.Syscall(createToolHelp32Snapshot.Addr(), 2,
		uintptr(dwFlags), uintptr(th32ProcessID), 0)
	return win.HANDLE(ret)
}

func Process32First(hSnapshot win.HANDLE, lppe *ProcessEntry32) bool {
	ret, _, _ := syscall.Syscall(process32First.Addr(), 2,
		uintptr(hSnapshot), uintptr(unsafe.Pointer(lppe)), 0)
	return ret != 0
}

func Process32Next(hSnapshot win.HANDLE, lppe *ProcessEntry32) bool {
	ret, _, _ := syscall.Syscall(process32Next.Addr(), 2,
		uintptr(hSnapshot), uintptr(unsafe.Pointer(lppe)), 0)
	return ret != 0
}

// 遍历所有进程,出现异常则退出
func RangeProcess(f func(*ProcessEntry32) error) (err error) {
	h := CreateToolHelp32Snapshot(TH32CsSNAPPROCESS, 0)
	if h == InvalidHandleValue {
		return fmt.Errorf("CreateToolHelp32Snapshot:%d", win.GetLastError())
	}
	defer win.CloseHandle(h)

	pe := new(ProcessEntry32)
	pe.DwSize = uint32(reflect.TypeOf(*pe).Size())
	for ok := Process32First(h, pe); ok; ok = Process32Next(h, pe) {
		if err = f(pe); err != nil {
			return
		}
	}
	return
}

// 根据进程名得到进程pid,返回nil表示空或错误
func FindPidFromName(name string) []uint32 {
	pid := make([]uint32, 0, 1)
	if RangeProcess(func(entry32 *ProcessEntry32) error {
		if name == BytesToString(entry32.SzExeFile[:len(name)]) {
			pid = append(pid, entry32.Th32ProcessID)
		}
		return nil
	}) != nil {
		return nil
	}
	return pid
}

/*----------------------------------------------------------------------------
Win32 C/C++ golang 字符对照表

	WIN32类型		C/C++ 类型			GO 类型
	HANDLE			void *				uintptr
	BYTE			unsigned char		uint8, byte
	SHORT			short				int16
	WORD			unsigned short		uint16
	INT				int					int32, int
	UINT			unsigned int		uint32
	LONG			long				int32
	BOOL			int					int
	DWORD			unsigned long		uint32
	ULONG			unsigned long		uint32
	CHAR			char				byte
	WCHAR			wchar_t				uint16
	LPSTR			utf8/char *			*byte
	LPCSTR			const utf8/char *	*byte, syscall.StringBytePtr(), xc.UTF8PtrToSting()
	LPWSTR			wchar_t *			*uint16
	LPCWSTR			const wchar_t *		*uint16, syscall.StringToUTF16Ptr()
	FLOAT			float				float32
	DOUBLE			double				float64
	LONGLONG		__int64				int64
	DWORD64			unsigned __int64	uint64
*/

const ProcessAllAccess = 0x1F0FFF // 所有权限

func OpenProcess(dwDesiredAccess, bInheritHandle, dwProcessId uint32) win.HANDLE {
	ret, _, _ := syscall.Syscall(openProcess.Addr(), 3,
		uintptr(dwDesiredAccess), /* 访问权限 */
		uintptr(bInheritHandle),  /* 是否继承句柄 */
		uintptr(dwProcessId))     /* 进程pid */
	return win.HANDLE(ret)
}

// 将hProcess进程的lpBaseAddress内存地址读出内容到lpBuffer中,lpNumberOfBytesRead为真实读出个数
func ReadProcessMemory(hProcess win.HANDLE, lpBaseAddress int32, lpBuffer []byte) bool {
	var lpNumberOfBytesRead int32
	readSize := int32(len(lpBuffer))
	ret, _, _ := syscall.Syscall6(readProcessMemory.Addr(), 5,
		uintptr(hProcess),
		uintptr(lpBaseAddress),
		uintptr(unsafe.Pointer(&lpBuffer[0])),
		uintptr(readSize),
		uintptr(unsafe.Pointer(&lpNumberOfBytesRead)), 0)
	if ret == 0 {
		return false
	}
	return readSize == lpNumberOfBytesRead
}

// 往hProcess进程的lpBaseAddress内存地址写入lpBuffer,lpNumberOfBytesWritten为真实写入个数
func WriteProcessMemory(hProcess win.HANDLE, lpBaseAddress int32, lpBuffer []byte) bool {
	var lpNumberOfBytesWritten int32
	writeSize := int32(len(lpBuffer))
	ret, _, _ := syscall.Syscall6(writeProcessMemory.Addr(), 5,
		uintptr(hProcess),
		uintptr(lpBaseAddress),
		uintptr(unsafe.Pointer(&lpBuffer[0])),
		uintptr(writeSize),
		uintptr(unsafe.Pointer(&lpNumberOfBytesWritten)), 0)
	if ret == 0 {
		return false
	}
	return writeSize == lpNumberOfBytesWritten
}

type MemoryBasicInformation struct {
	BaseAddress       int64
	AllocationBase    int64
	AllocationProtect int32
	RegionSize        int64
	State             int32
	Protect           int32
	Type              int32
}

const (
	PageReadWrite = 4
	MemCommit     = 4096
)

// 成功返回true
func VirtualQueryEx(hProcess win.HANDLE, lpAddress int64, lpBuffer *MemoryBasicInformation) bool {
	dwLength := unsafe.Sizeof(*lpBuffer)
	ret, _, _ := syscall.Syscall6(virtualQueryEx.Addr(), 4,
		uintptr(hProcess),
		uintptr(lpAddress),
		uintptr(unsafe.Pointer(lpBuffer)),
		dwLength, 0, 0)
	return ret == dwLength
}

type SystemInfo struct {
	WProcessorArchitecture      int16 // dwOemId
	WReserved                   int16 // dwOemId
	DwPageSize                  int32
	LpMinimumApplicationAddress int64
	LpMaximumApplicationAddress int32
	DwActiveProcessorMask       int64
	DwNumberOfProcessors        int32
	DwProcessorType             int32
	DwAllocationGranularity     int32
	WProcessorLevel             int16
	WProcessorRevision          int16
}

func GetSystemInfo(lpSystemInfo *SystemInfo) bool {
	ret, _, _ := syscall.Syscall(getSystemInfo.Addr(), 1,
		uintptr(unsafe.Pointer(lpSystemInfo)), 0, 0)
	return ret != 0
}

/*----------------------------------------------------------------------------*/
