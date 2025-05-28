package httpx

import (
	"fmt"
	"net/http"
	"strings"
)

// Ping 检查一个url是否有效，有效则返回true，无效则返回false
func Ping(url string) (bool, error) {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}
	cli := http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true, // 只访问一次，禁用长连接
		},
	}
	resp, err := cli.Get(url)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status code want %d, but got %d", http.StatusOK, resp.StatusCode)
	}
	return true, nil
}
