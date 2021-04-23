package qn_webui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

// WEBUI接口
type WEBUI interface {
	Load(url string) error
	Bounds() (Bounds, error)
	SetBounds(Bounds) error
	Bind(name string, f interface{}) error
	Eval(js string) Value
	Done() <-chan struct{}
	Close() error
}

//webui结构体
type webui struct {
	chrome *chrome
	done   chan struct{}
	tmpDir string
}

var defaultChromeArgs = []string{
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-breakpad",
	"--disable-client-side-phishing-detection",
	"--disable-default-apps",
	"--disable-dev-shm-usage",
	"--disable-infobars",
	"--disable-extensions",
	"--disable-features=site-per-process",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	"--disable-sync",
	"--disable-translate",
	"--disable-windows10-custom-titlebar",
	"--metrics-recording-only",
	"--no-first-run",
	"--no-default-browser-check",
	"--safebrowsing-disable-auto-update",
	"--enable-automation",
	"--password-store=basic",
	"--use-mock-keychain",
}

/*
新建窗体
@url 网址
@dir 目录
@width 窗体宽度
@height 窗体高度
@customArgs 其它参数
*/
func New(url, dir string, width, height int, customArgs ...string) (WEBUI, error) {
	if url == "" {
		url = "data:text/html,<html></html>"
	}
	tmpDir := ""
	if dir == "" {
		name, err := ioutil.TempDir("", "lorca")
		if err != nil {
			return nil, err
		}
		dir, tmpDir = name, name
	}
	args := append(defaultChromeArgs, fmt.Sprintf("--app=%s", url))
	args = append(args, fmt.Sprintf("--user-data-dir=%s", dir))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", width, height))
	args = append(args, customArgs...)
	args = append(args, "--remote-debugging-port=0")

	chrome, err := newChromeWithArgs(ChromeExecutable(), args...)
	done := make(chan struct{})
	if err != nil {
		return nil, err
	}

	go func() {
		chrome.cmd.Wait()
		close(done)
	}()
	return &webui{chrome: chrome, done: done, tmpDir: tmpDir}, nil
}

func (u *webui) Done() <-chan struct{} {
	return u.done
}

/*
关闭窗体
*/
func (u *webui) Close() error {
	//忽略错误，因为当用户关闭窗口时，已经杀死进程。
	u.chrome.kill()
	<-u.done
	if u.tmpDir != "" {
		if err := os.RemoveAll(u.tmpDir); err != nil {
			return err
		}
	}
	return nil
}

/*
加载窗体
@url 网址
*/
func (u *webui) Load(url string) error { return u.chrome.load(url) }

/*
窗体绑定函数
@name 函数名
@f go执行函数
*/
func (u *webui) Bind(name string, f interface{}) error {
	v := reflect.ValueOf(f)
	// f参数必须是函数
	if v.Kind() != reflect.Func {
		return errors.New("only functions can be bound")
	}
	// f 必须有返回值和错误信息
	if n := v.Type().NumOut(); n > 2 {
		return errors.New("函数只能返回值或值+错误信息")
	}
	return u.chrome.bind(name, func(raw []json.RawMessage) (interface{}, error) {
		if len(raw) != v.Type().NumIn() {
			return nil, errors.New("function arguments mismatch")
		}
		args := []reflect.Value{}
		for i := range raw {
			arg := reflect.New(v.Type().In(i))
			if err := json.Unmarshal(raw[i], arg.Interface()); err != nil {
				return nil, err
			}
			args = append(args, arg.Elem())
		}
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		res := v.Call(args)
		switch len(res) {
		case 0: //如果是0个结果函数没有结果返回空
			return nil, nil
		case 1: //如果是1个结果可能是返回值也可能是返回的错误信息
			if res[0].Type().Implements(errorType) {
				if res[0].Interface() != nil {
					return nil, res[0].Interface().(error)
				}
				return nil, nil
			}
			return res[0].Interface(), nil
		case 2: //如果是2个结果第一个为值第二个为错误信息
			if !res[1].Type().Implements(errorType) {
				return nil, errors.New("second return value must be an error")
			}
			if res[1].Interface() == nil {
				return res[0].Interface(), nil
			}
			return res[0].Interface(), res[1].Interface().(error)
		default: //其它返回值
			return nil, errors.New("unexpected number of return values")
		}
	})
}

func (u *webui) Eval(js string) Value {
	v, err := u.chrome.eval(js)
	return value{err: err, raw: v}
}

func (u *webui) SetBounds(b Bounds) error {
	return u.chrome.setBounds(b)
}

func (u *webui) Bounds() (Bounds, error) {
	return u.chrome.bounds()
}
