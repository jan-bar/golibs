package golibs

import (
  "syscall"
  "time"
  
  "unsafe"
  
  "github.com/lxn/win"
)

const (
  StdOutputHandle     = 0xFFFFFFF5
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
)

type (
  Word uint16
  DWord uint32
  TChar rune
)

type Coord struct {
  X, Y int
}

type SmallRect struct {
  Left, Top, Right, Bottom int16
}

type ConsoleScreenBufferInfo struct {
  DwSize              Coord
  DwCursorPosition    Coord
  WAttributes         Word
  SrWindow            SmallRect
  DwMaximumWindowSize Coord
}

type ConsoleCursorInfo struct {
  dwSize   DWord
  bVisible DWord
}

var (
  fillConsoleOutputAttribute  uintptr
  fillConsoleOutputCharacterW uintptr
  getStdHandle                uintptr
  getConsoleScreenBufferInfo  uintptr
  setConsoleCursorPosition    uintptr
  setConsoleTextAttribute     uintptr
  setConsoleCursorInfo        uintptr
  getConsoleWindow            uintptr
  getKeyState                 uintptr /* 处理win32api,获取键盘事件 */
  setWindowText               uintptr
  hConsole                    win.HWND
)

/* 将 Coord 转换为 Dword */
func CoordToDword(c Coord) DWord {
  return DWord(int32(c.Y)<<16 + int32(c.X))
}

/**
* 初始化
* 主要加载win32api的方法
**/
func init() {
  kernel32, err := syscall.LoadLibrary("kernel32.dll")
  if err != nil {
    panic(err)
  }
  
  user32, err := syscall.LoadLibrary("user32.dll")
  if err != nil {
    panic(err)
  }
  
  fillConsoleOutputAttribute, err = syscall.GetProcAddress(kernel32, "FillConsoleOutputAttribute")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  fillConsoleOutputCharacterW, err = syscall.GetProcAddress(kernel32, "FillConsoleOutputCharacterW")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  getStdHandle, err = syscall.GetProcAddress(kernel32, "GetStdHandle")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  getConsoleScreenBufferInfo, err = syscall.GetProcAddress(kernel32, "GetConsoleScreenBufferInfo")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  setConsoleCursorPosition, err = syscall.GetProcAddress(kernel32, "SetConsoleCursorPosition")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  setConsoleTextAttribute, err = syscall.GetProcAddress(kernel32, "SetConsoleTextAttribute")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  setConsoleCursorInfo, err = syscall.GetProcAddress(kernel32, "SetConsoleCursorInfo")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  getConsoleWindow, err = syscall.GetProcAddress(kernel32, "GetConsoleWindow")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  getKeyState, err = syscall.GetProcAddress(user32, "GetKeyState")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  setWindowText, err = syscall.GetProcAddress(user32, "SetWindowTextW")
  if err != nil { /* 获取句柄失败 */
    panic(err)
  }
  
  hConsole = GetStdHandle(StdOutputHandle)
}

/**
* 设置固定区域内的文本属性，从指定的控制台屏幕缓冲区字符坐标开始。
**/
func FillConsoleOutputAttribute(hConsoleOutput win.HWND, wAttribute Word, nLength DWord, dwWriteCoord Coord) *DWord {
  var lpNumberOfAttrsWritten DWord
  ret, _, _ := syscall.Syscall6(fillConsoleOutputAttribute, 5,
    uintptr(hConsoleOutput),
    uintptr(wAttribute),
    uintptr(nLength),
    uintptr(CoordToDword(dwWriteCoord)),
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
func FillConsoleOutputCharacter(hConsoleOutput win.HWND, cCharacter TChar, nLength DWord, dwWriteCoord Coord) *DWord {
  var lpNumberOfAttrsWritten DWord
  ret, _, _ := syscall.Syscall6(fillConsoleOutputCharacterW, 5,
    uintptr(hConsoleOutput),
    uintptr(cCharacter),
    uintptr(nLength),
    uintptr(CoordToDword(dwWriteCoord)),
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
func GetStdHandle(nStdHandle DWord) win.HWND {
  ret, _, _ := syscall.Syscall(getStdHandle, 1,
    uintptr(nStdHandle),
    0,
    0)
  return win.HWND(ret)
}

/**
* 用于检索指定的控制台屏幕缓冲区的信息
**/
func GetConsoleScreenBufferInfo(hConsoleOutput win.HWND) *ConsoleScreenBufferInfo {
  var CsBi ConsoleScreenBufferInfo
  ret, _, _ := syscall.Syscall(getConsoleScreenBufferInfo, 2,
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
func SetConsoleCursorPosition(hConsoleOutput win.HWND, dwCursorPosition Coord) bool {
  ret, _, _ := syscall.Syscall(setConsoleCursorPosition, 2,
    uintptr(hConsoleOutput),
    uintptr(CoordToDword(dwCursorPosition)),
    0)
  return ret != 0
}

/**
* 设置控制台窗口字体颜色和背景色的计算机函数
**/
func SetConsoleTextAttribute(hConsoleOutput win.HWND, wAttributes Word) bool {
  ret, _, _ := syscall.Syscall(setConsoleTextAttribute, 2,
    uintptr(hConsoleOutput),
    uintptr(wAttributes),
    0)
  return ret != 0
}

/**
* 设置光标属性
**/
func SetConsoleCursorInfo(hConsoleOutput win.HWND, lpConsoleCursorInfo ConsoleCursorInfo) bool {
  ret, _, _ := syscall.Syscall(setConsoleCursorInfo, 2,
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
func ShowHideCursor(show bool) {
  var bVisible DWord = 0
  if show {
    bVisible = 1
  }
  SetConsoleCursorInfo(hConsole, ConsoleCursorInfo{1, bVisible})
}

/**
* 清屏函数
**/
func Clear() {
  hConsole = GetStdHandle(StdOutputHandle)
  coordScreen := Coord{0, 0}
  csbi := GetConsoleScreenBufferInfo(hConsole)
  dwConSize := DWord(csbi.DwSize.X * csbi.DwSize.Y)
  FillConsoleOutputCharacter(hConsole, TChar(' '), dwConSize, coordScreen)
  csbi = GetConsoleScreenBufferInfo(hConsole)
  FillConsoleOutputAttribute(hConsole, csbi.WAttributes, dwConSize, coordScreen)
  SetConsoleCursorPosition(hConsole, coordScreen)
}

/**
* 光标定位到某个位置
**/
func GotoXY(x, y int) {
  hConsole = GetStdHandle(StdOutputHandle)
  SetConsoleCursorPosition(hConsole, Coord{x, y})
}

/**
* 设置打印颜色
**/
func TextBackground(color int) {
  hConsole = GetStdHandle(StdOutputHandle)
  SetConsoleTextAttribute(hConsole, Word(color))
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
  ret, _, _ := syscall.Syscall(getKeyState, 1, uintptr(nVirtKey), 0, 0)
  return int16(ret) < 0
}

/**
* 获取当前控制台句柄
**/
func GetConsoleWindow() win.HWND {
  ret, _, _ := syscall.Syscall(getConsoleWindow, 0, 0, 0, 0)
  return win.HWND(ret)
}

/**
* 设置窗体左上角文字
* SetWindowTextW (Unicode) and SetWindowTextA (ANSI)
**/
func SetWindowText(text string) bool {
  str := win.SysAllocString(text) /* 申请字符串 */
  ret, _, _ := syscall.Syscall(setWindowText, 2,
    uintptr(GetConsoleWindow()),
    uintptr(unsafe.Pointer(str)),
    0)
  win.SysFreeString(str) /* api说明中,用完就释放 */
  return ret != 0
}

/**
* 居中显示窗体
* 并设置宽高
**/
func CenterWindowOnScreen(w, h int32) {
  xLeft := (win.GetSystemMetrics(win.SM_CXFULLSCREEN) - w) / 2
  yTop := (win.GetSystemMetrics(win.SM_CYFULLSCREEN) - h) / 2
  hWnd := GetConsoleWindow()
  win.SetWindowPos(hWnd, win.HWND_TOPMOST, xLeft, yTop, w, h, win.SWP_NOZORDER)
  win.SetWindowPos(hWnd, win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
}
