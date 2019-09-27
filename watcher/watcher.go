package watcher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	Create  = "Create"
	Remove  = "Remove"
	Modify  = "Modify"
	Unknown = "Unknown"
)

var watcherStruct = struct {
	sync.Mutex
	Rec map[string]*watcher
}{
	Rec: make(map[string]*watcher, 32),
}

/*----------------------------------------------------------------------------*/

type Notifier interface {
	Notify(event string, path string) error
}

type watcher struct {
	path string
	fw   *fsnotify.Watcher
	n    Notifier
}

func (w *watcher) watch() {
	var err error
	for {
		select {
		case event := <-w.fw.Events:
			err = w.handleEvent(event)
			if err != nil {
				fmt.Println(err)
			}
		case err = <-w.fw.Errors:
			if err == nil {
				return
			}
			fmt.Println(err)
		}
	}
}

func (w *watcher) handleEvent(fe fsnotify.Event) error {
	e := w.acquireEventStr(fe.Op)
	if e == Unknown {
		return nil
	}
	if e == Remove && fe.Name == w.path {
		// NOTE 如果监听文件被移除，则监听器被移除，后续关于该文件的任何信息，均不会收到通知
		return w.handleReInit()
	}
	return w.n.Notify(e, fe.Name)
}

func (w *watcher) acquireEventStr(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return Create
	case op&fsnotify.Remove == fsnotify.Remove ||
		op&fsnotify.Rename == fsnotify.Rename:
		return Remove
	case op&fsnotify.Write == fsnotify.Write:
		return Modify
	default:
		return Unknown
	}
}

func (w *watcher) handleReInit() error {
	err := w.fw.Close()
	if err != nil {
		return err
	}
	w.fw = nil

	for {
		fw, err := fsnotify.NewWatcher()
		if err != nil {
			goto handleError
		}

		err = fw.Add(w.path)
		if err != nil {
			goto handleError
		}
		w.fw = fw
		break
	handleError:
		time.Sleep(3 * time.Second)
	}

	return w.n.Notify(Create, w.path)
}

func Add(path string, n Notifier) error {
	err := checkWatchParams(path, n)
	if err != nil {
		return err
	}

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = fw.Add(path)
	if err != nil {
		return err
	}

	w := &watcher{
		path: path,
		fw:   fw,
		n:    n,
	}

	watcherStruct.Lock()
	watcherStruct.Rec[path] = w
	watcherStruct.Unlock()
	go w.watch()
	return nil
}

func checkWatchParams(path string, n Notifier) error {
	switch {
	case path == "":
		return errors.New("path can't be nil")
	case n == nil:
		return errors.New("notifier can't be nil")
	}

	watcherStruct.Lock()
	_, ok := watcherStruct.Rec[path]
	watcherStruct.Unlock()
	if ok {
		return fmt.Errorf("duplicated notified for path %s", path)
	}
	_, err := os.Stat(path)
	return err
}

/*----------------------------------------------------------------------------*/

// UserProfile 为用户配置结构的接口，当以函数Name的返回值为名称的配置文件被创建或者修改时，
// Modify函数会被调用，触发配置结构内容的更改
// Modify函数必须有事务支持，即当Modify函数返回error时，配置结构的内容不应被改变
// TODO 是否应该支持Name会动态变化的？
type UserProfile interface {
	Name() string
	Modify(data []byte) error
}

var watcherUserProfile struct {
	sync.RWMutex
	called   int32
	watchMap map[string]UserProfile
}

func initUserProfile(dir string) error {
	watcherUserProfile.Lock()
	defer watcherUserProfile.Unlock()
	watcherUserProfile.watchMap = make(map[string]UserProfile, 8)
	return Add(dir, handler{}) // 监控整个目录
}

/*----------------------------------------------------------------------------*/

func Load(dir, name string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(dir, name))
}

// Register 函数注册配置文件到观察列表中，每次配置文件发送变动时，通知profile进行更新
func Register(dir string, profile UserProfile) error {
	err := checkParams(profile)
	if err != nil {
		return err
	}

	if atomic.CompareAndSwapInt32(&watcherUserProfile.called, 0, 1) {
		err = initUserProfile(dir)
		if err != nil {
			atomic.StoreInt32(&watcherUserProfile.called, 0)
			return err
		}
	}

	watcherUserProfile.Lock()
	watcherUserProfile.watchMap[profile.Name()] = profile
	watcherUserProfile.Unlock()

	data, err := Load(dir, profile.Name())
	if err != nil {
		return err
	}
	return profile.Modify(data)
}

func checkParams(profile UserProfile) error {
	if profile == nil {
		return errors.New("profile can't be nil")
	} else if profile.Name() == "" {
		return errors.New("profile name can't be nil")
	}
	return nil
}

/*----------------------------------------------------------------------------*/

type handler struct{}

func (_ handler) Notify(event string, path string) error {
	if event != Create && event != Modify {
		return nil
	}

	watcherUserProfile.RLock()
	p, ok := watcherUserProfile.watchMap[filepath.Base(path)]
	watcherUserProfile.RUnlock()
	if !ok {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return p.Modify(data)
}
