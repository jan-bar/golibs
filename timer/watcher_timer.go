package timer

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// 检测系统时间或时区变化
const (
	accuracy  = time.Second     // 检测精度
	threshold = 5 * time.Second // 时间触发精度
)

type Notifier interface {
	Notify(time time.Time) error
}

var watcherStruct = struct {
	sync.Mutex
	Rec map[string]*watcher
}{
	Rec: make(map[string]*watcher, 32),
}

type watcher struct {
	t chan time.Time
	n Notifier
}

func (w *watcher) watch() {
	for {
		t, ok := <-w.t
		if ok {
			w.n.Notify(t)
		}
	}
}

func Add(name string, n Notifier) error {
	if name == "" {
		return errors.New("name can't be nil")
	}
	if n == nil {
		return errors.New("notifier can't be nil")
	}
	watcherStruct.Lock()
	_, ok := watcherStruct.Rec[name]
	watcherStruct.Unlock()
	if ok {
		return fmt.Errorf("duplicated notified for path %s", name)
	}
	watch.Do(func() {
		watch.rec = make([]chan time.Time, 0, 8)
		go watch.watch()
	})

	t := make(chan time.Time, 1)
	watch.register(t)
	w := &watcher{t: t, n: n}

	watcherStruct.Lock()
	watcherStruct.Rec[name] = w
	watcherStruct.Unlock()
	go w.watch()
	return nil
}

/*----------------------------------------------------------------------------*/

type timeWatch struct {
	sync.Once
	sync.Mutex
	rec []chan time.Time
}

var watch timeWatch

func (t *timeWatch) register(time chan time.Time) {
	t.Lock()
	t.rec = append(t.rec, time)
	t.Unlock()
}

func (t *timeWatch) watch() {
	lastTime, nowTime := time.Now(), time.Now()
	ticker := time.NewTicker(accuracy)
	defer ticker.Stop()
	for nowTime = range ticker.C {
		nowTime = nowTime.Round(0) // strip monotonic clock reading
		if nowTime.Before(lastTime) || lastTime.Add(threshold).Before(nowTime) {
			t.Lock()
			tNow := time.Now()
			for _, cTime := range t.rec {
				cTime <- tNow
			}
			t.Unlock()
		}
		lastTime = nowTime
	}
}
