package timer

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// go test . -v
// 系统修改时间会通知注册者

type test struct {
}

func (t *test) Notify(time time.Time) error {
	fmt.Println("test", time.String())
	return nil
}

type test1 struct {
}

func (t *test1) Notify(time time.Time) error {
	fmt.Println("test1", time.String())
	return nil
}

func TestWatcher(t *testing.T) {
	Convey("test watcher", t, func() {
		err := Add("test", &test{})
		So(err, ShouldBeNil)
		err = Add("test1", &test1{})
		So(err, ShouldBeNil)

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
	})
}
