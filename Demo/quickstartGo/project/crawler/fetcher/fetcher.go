package fetcher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36 Edg/88.0.705.81"
const cookie = "sid=69222b28-da04-4e29-abb7-8f070f0730a4; FSSBBIl1UgzbN7NO=5WGo3XSgAK7XdGgEcQ5KE7X9BWBx4f6unpiDu5sNe6xw796jH_T3DrDyPI3eGD0V_Ga3sKupQRHRFFjnKXU3NBG; ec=ldNJNEZd-1620024091439-de4bbf31a9dc3860282082; Hm_lvt_2c8ad67df9e787ad29dbd54ee608f5d2=1620057995,1620109355,1620117333,1620126862; _exid=RbB7vkgnaWjKNxi32qyj87xG8t2hfsr/SOiRNEp1vXVm8L9sToLykUPoS54hrwan+jCst56D9U9f0bfLdN4tpw==; _efmdata=mVQlFPEoJJ0pKtw00+r76CtqzHhOfkZK8FLgCOcz4c8vQbzshPV3lEx0niVPOIGkWQnmHZ9EeeWogoeqG2J09WOFYKJ7FNLq9zHGM27HooY=; Hm_lpvt_2c8ad67df9e787ad29dbd54ee608f5d2=1620127829; FSSBBIl1UgzbN7NP=53xlaYCcpWmWqqqmyJnX0rA5F6e266LWZ1ZK7RBV1szEl1IlYjVAIfR4wKDFXJlfbyVxKOG7Z8WC52v4.VSES1wnPZArJZvgVR77QeH0e9ZOCVbBE7669dp18KJrxVXGwgzRrCXxRSeZrFI7Xiul847KwFK2xpwcokdhqD5vHj7byJvwT.qW5evXdl90e5o4UGKTDskGihGaSeOTstz4SOibK71GCiHUI0U3zZ6SHsI70ZQcSWR3I0NpjzZs7qlRaV0LNZdw9VixvpywvMASNDmrmGkuzGtYchyzXwKI_jGJjsLpIUlGEj2HOTrnUGQTog"

// 封装爬取url的方法 并将body 内容返回
// 控制访问频率  让多个worker 都拥有一个全局变量 rateLimiter 争抢channel返回的信号
var rateLimiter = time.Tick(time.Second)

func Fetch(url string) ([]byte, error) {
	<-rateLimiter
	url = strings.Replace(url, "http://", "https://", 1)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)

	}
	req.Header.Set("User-Agent", UA)
	req.Header.Set("cookie", cookie)
	client := &http.Client{}
	// 没网导致resp为空 err 不为空 invalid memory address or nil pointer dereference
	// 无法解决cookie问题
	resp, err := client.Do(req)

	//为空
	//cookie := resp.Header.Get("cookie")
	//fmt.Printf("fetcher cookie: %s\n",cookie)

	//f := NewFetcher("golang.org")
	//resp, _, err := f.Get(url)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	//Error : status code 202
	//fetch error : http response code is not okget person profile
	//https://zhuanlan.zhihu.com/p/81008327
	//
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error : status code", resp.StatusCode)
		return nil, errors.New("http response code is not ok\n")
	}
	result, err := ioutil.ReadAll(resp.Body)
	return result, err
}
