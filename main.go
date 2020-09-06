package main

import (
	"github.com/elgs/cron"
	winapi "github.com/kbinani/win"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Windows struct {
	user32 *syscall.LazyDLL
}

func NewWindows() *Windows {
	// 加载类库
	user32 := syscall.NewLazyDLL("user32.dll")
	return &Windows{user32}
}

// 是否已经锁屏
func (w *Windows) isScreenLock() (bool, error) {
	const successCallMessage = "The operation completed successfully."

	// 创建新的调用进程
	getForegroundWindow := w.user32.NewProc("GetForegroundWindow")
	// 调用相应的函数
	activeWindowId, _, err := getForegroundWindow.Call()

	if err != nil && err.Error() != successCallMessage {
		return false, err
	}
	if activeWindowId == 0 {
		return true, nil
	}
	return false, nil
}

func main() {
	w := NewWindows()
	c := cron.New()
	c.Start()
	defer c.Stop()
	cronStr := "0 0 * * * *"
	_, err := c.AddFunc(cronStr, func() {
		f, err2 := w.isScreenLock()
		if err2 != nil {
			logrus.Errorf("isScreenLock:%s", err2)
			return
		}
		if !f {
			var a winapi.HWND
			winapi.MessageBox(a, "10秒后要休息了", "提示", 100)
			time.Sleep(10 * time.Second)
			_, _, err4 := Execute("rundll32.exe user32.dll LockWorkStation")
			if err4 != nil {
				logrus.Errorf("Execute Cmd:%s", err4)
			}
			logrus.Errorf("LockScreen Success")
		}
	})
	if err != nil {
		logrus.Errorf("AddFunc:%s %s", cronStr, err)
		os.Exit(-1)
	}

	for _, entry := range c.Entries() {
		logrus.WithFields(logrus.Fields{
			"func":   "TurnGroup",
			"Id":     entry.Id,
			"Next":   entry.Next,
			"Prev":   entry.Prev,
			"Status": entry.Status,
		}).Info("job entry")
	}

	logrus.Infof("Waiting Stop Signal ...pid:%d",os.Getpid())
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutdown Server Done")
}

func Execute(command string) (bool, string, error) {
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(command)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		return false, "", err
	}
	return true, string(out), nil
}

func strToUint16(s string) (*uint16, error) {
	return syscall.UTF16PtrFromString(s)
}
