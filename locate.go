package qn_webui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

//本地Chrome安装路径
var ChromeExecutable = LocateChrome

/*
加载一个Chrome安装
*/
func LocateChrome() string {

	// 获取环境变量
	if path, ok := os.LookupEnv("LORCACHROME"); ok {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
	case "windows":
		_path, _ := os.Getwd()
		paths = []string{
			_path + "/client/chrome.exe",
			_path + "/client/chrome.exe",
			_path + "/client/chrome.exe",
			_path + "/client/chrome.exe",
			_path + "/client/chrome.exe",
			_path + "/client/chrome.exe",
		}
	default:
		paths = []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		return path
	}
	return ""
}

/*
加载Chrome浏览器
*/
func PromptDownload() {
	title := "Chrome没有安装"
	text := "未发现Chrome/Chromium。是否立即下载并安装？"
	// 弹窗提示
	if !messageBox(title, text) {
		return
	}
	//打开下载页面
	url := "https://www.google.com/chrome/"
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Run()
	case "darwin":
		exec.Command("open", url).Run()
	case "windows":
		r := strings.NewReplacer("&", "^&")
		exec.Command("cmd", "/c", "start", r.Replace(url)).Run()
	}
}
