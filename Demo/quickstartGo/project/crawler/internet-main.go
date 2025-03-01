package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

var gCurCookies []*http.Cookie
var gCurCookieJar *cookiejar.Jar

func initAll() {
	gCurCookies = nil
	//var err error;
	gCurCookieJar, _ = cookiejar.New(nil)

}

// 1 get url response html
func getUrlRespHtml(url string) string {
	fmt.Printf("\ngetUrlRespHtml, url=%s", url)

	var respHtml string = ""

	httpClient := &http.Client{
		CheckRedirect: nil,
		Jar:           gCurCookieJar,
	}

	httpReq, err := http.NewRequest("GET", url, nil)
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		fmt.Printf("\nhttp get url=%s response error=%s\n", url, err.Error())
	}
	fmt.Printf("\nhttpResp.Header=%s", httpResp.Header)
	fmt.Printf("\nhttpResp.Status=%s", httpResp.Status)

	defer httpResp.Body.Close()

	body, errReadAll := ioutil.ReadAll(httpResp.Body)
	if errReadAll != nil {
		fmt.Printf("\nget response for url=%s got error=%s\n", url, errReadAll.Error())
	}
	//全局保存
	gCurCookies = gCurCookieJar.Cookies(httpReq.URL)

	respHtml = string(body)
	return respHtml
}

// 2
func getUrlRespHtmlWithHeader(url, headers string) string {
	fmt.Printf("\ngetUrlRespHtml, url=%s", url)

	var respHtml string = ""

	httpClient := &http.Client{
		CheckRedirect: nil,
		Jar:           gCurCookieJar,
	}

	httpReq, err := http.NewRequest("GET", url, nil)
	AddHeaders(httpReq, headers)
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		fmt.Printf("\nhttp get url=%s response error=%s\n", url, err.Error())
	}
	fmt.Printf("\nhttpResp.Header=%s", httpResp.Header)
	fmt.Printf("\nhttpResp.Status=%s", httpResp.Status)
	fmt.Printf("\nhttpResp.cookies=%s", httpResp.Cookies())

	defer httpResp.Body.Close()

	body, errReadAll := ioutil.ReadAll(httpResp.Body)
	if errReadAll != nil {
		fmt.Printf("\nget response for url=%s got error=%s\n", url, errReadAll.Error())
	}
	//全局保存
	gCurCookies = gCurCookieJar.Cookies(httpReq.URL)

	respHtml = string(body)
	return respHtml
}

// 3
func PostUrlRespHtmlWithHeader(url, headers, formdata string) string {
	fmt.Printf("\ngetUrlRespHtml, url=%s", url)

	var respHtml string = ""

	httpClient := &http.Client{
		CheckRedirect: nil,
		Jar:           gCurCookieJar,
	}

	httpReq, err := http.NewRequest("POST", url, ioutil.NopCloser(strings.NewReader(formdata)))
	AddHeaders(httpReq, headers)
	httpReq.Header.Set("ContentType", "application/x-www-form-urlencoded")
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		fmt.Printf("\nhttp get url=%s response error=%s\n", url, err.Error())
	}
	fmt.Printf("\nhttpResp.Header=%s", httpResp.Header)
	fmt.Printf("\nhttpResp.Status=%s", httpResp.Status)

	defer httpResp.Body.Close()

	body, errReadAll := ioutil.ReadAll(httpResp.Body)
	if errReadAll != nil {
		fmt.Printf("\nget response for url=%s got error=%s\n", url, errReadAll.Error())
	}
	//全局保存
	gCurCookies = gCurCookieJar.Cookies(httpReq.URL)

	respHtml = string(body)
	return respHtml
}

func dbgPrintCurCookies() {
	var cookieNum int = len(gCurCookies)
	fmt.Printf("cookieNum=%d", cookieNum)
	for i := 0; i < cookieNum; i++ {
		var curCk *http.Cookie = gCurCookies[i]
		fmt.Printf("\n\n\n\n------ Cookie [%d]------", i)
		fmt.Printf("\n\tName=%s", curCk.Name)
		fmt.Printf("\n\tValue=%s", curCk.Value)
		fmt.Printf("\n\tPath=%s", curCk.Path)
		fmt.Printf("\n\tDomain=%s", curCk.Domain)
		fmt.Printf("\n\tExpires=%s", curCk.Expires)
		fmt.Printf("\n\tRawExpires=%s", curCk.RawExpires)
		fmt.Printf("\n\tMaxAge=%d", curCk.MaxAge)
		fmt.Printf("\n\tSecure=%t", curCk.Secure)
		fmt.Printf("\n\tHttpOnly=%t", curCk.HttpOnly)
		fmt.Printf("\n\tRaw=%s", curCk.Raw)
		fmt.Printf("\n\tUnparsed=%s", curCk.Unparsed)
	}
}

func AddHeaders(req *http.Request, headers string) *http.Request {
	//将传入的Header分割成[]ak和[]av
	a := strings.Split(headers, "\n")
	ak := make([]string, len(a[:]))
	av := make([]string, len(a[:]))
	//要用copy复制值；若用等号仅表示指针，会造成修改ak也就是修改了av
	copy(ak, a[:])
	copy(av, a[:])
	//fmt.Println(ak[0], av[0])
	for k, v := range ak {
		i := strings.Index(v, ":")
		j := i + 1
		ak[k] = v[:i]
		av[k] = v[j:]
		//设置Header
		req.Header.Set(ak[k], av[k])
	}
	return req
}

func main() {
	initAll()
	/*
	   fmt.Printf("====== step 1：get Cookie ======")
	   var MainUrl string = "http://192.168.132.80/login/login.jsp"
	   fmt.Printf("\nMainUrl=%s", MainUrl)
	   getUrlRespHtmlWithHeader(MainUrl, headers2)
	   dbgPrintCurCookies()
	*/

	fmt.Printf("\n\n\n====== step 2：get Cookie ======")
	var headers2 = `Accept: text/html, application/xhtml+xml, */*
Referer: http://192.168.132.80/login/login.jsp
Accept-Language: zh-CN
User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko
Content-Type: application/x-www-form-urlencoded
Accept-Encoding: gzip, deflate
Host: 192.168.132.80
Content-Length: 258
Connection: Keep-Alive
Pragma: no-cache
Cookie: logincookiecheck=1581819550262+C1D3FCB434C8223BE9C4CE5AD9497183; testBanCookie=test; JSESSIONID=abcrJrk4lxqzZccwgDUax; loginfileweaver=%2Fwui%2Ftheme%2Fecology7%2Fpage%2Flogin.jsp%3FtemplateId%3D6%26logintype%3D1%26gopage%3D; loginidweaver=114; languageidweaver=7`
	var formdata = `loginfile=%2Fwui%2Ftheme%2Fecology7%2Fpage%2Flogin.jsp%3FtemplateId%3D6%26logintype%3D1%26gopage%3D&logintype=1&fontName=%CE%A2%C8%ED%D1%C5%BA%DA&message=&gopage=&formmethod=post&rnd=&serial=&username=&isie=true&loginid=admin&userpassword=1234&submit=`
	var getapiUrl string = "http://192.168.132.80/login/VerifyLogin.jsp "
	PostUrlRespHtmlWithHeader(getapiUrl, headers2, formdata)
	dbgPrintCurCookies()

	fmt.Printf("\n\n\n====== step 3：use the Cookie ======")
	var headers3 = `Host: 192.168.132.80
Connection: keep-alive
Pragma: no-cache
Cache-Control: no-cache
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.79 Safari/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
Accept-Encoding: gzip, deflate
Accept-Language: zh-CN,zh;q=0.9`
	var getapiUrl3 string = "http://192.168.132.80/docs/docs/DocMoreForHp.jsp?eid=660&date2during=0&tabid=2"
	getUrlRespHtmlWithHeader(getapiUrl3, headers3)
	dbgPrintCurCookies()
}
