package watcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"janbar.top/config"
)

// go test . -v
// 检测文件或文件夹被修改,可用于重新加载配置文件

type watch struct {
	f string
}

func (w *watch) Notify(event string, path string) error {
	fmt.Println(event, path)
	return nil
}

func TestWatcher(t *testing.T) {
	Convey("test watcher", t, func() {
		name := config.JoinConfigPath("b.json")
		err := ioutil.WriteFile(name, []byte("123"), os.ModePerm)
		So(err, ShouldBeNil)
		err = Add(name, &watch{f: name})
		So(err, ShouldBeNil)
		ioutil.WriteFile(name, []byte("456"), os.ModePerm)
		time.Sleep(time.Second)
	})
}

type test struct {
	name string
}

func (t *test) Name() string {
	return t.name
}

func (t *test) Modify(data []byte) error {
	fmt.Println(string(data))
	return nil
}

func TestWatcherUserProfile(t *testing.T) {
	Convey("test watcher UserProfile", t, func() {
		name := config.JoinConfigPath("a.json")
		err := ioutil.WriteFile(name, []byte("123"), os.ModePerm)
		So(err, ShouldBeNil)
		err = Register(&test{name: "a.json"})
		So(err, ShouldBeNil)
		c := make(chan struct{})
		go func() {
			for i := 0; i < 3; i++ {
				ioutil.WriteFile(name, []byte(strconv.Itoa(i)), os.ModePerm)
				time.Sleep(time.Second)
			}
			c <- struct{}{}
		}()
		<-c
	})
}
