package tool

import (
	"fmt"
	"github.com/anaskhan96/soup"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"
)

func CCNULogin(studentID string, password string) bool {
	htmlBody, err := soup.Get("https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	if err != nil {
		return false
	}
	js, st, ok := parseCasLogin(htmlBody)
	if !ok {
		return false
	}
	jar, _ := cookiejar.New(&cookiejar.Options{})

	client := &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://account.ccnu.edu.cn/cas/login;jsessionid=%v?service=http", js) + "%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal"
	text := fmt.Sprintf("username=%v&password=%v&lt=%v&execution=e1s1&_eventId=submit&submit=", studentID, password, st) + "%E7%99%BB%E5%BD%95"
	body := strings.NewReader(text)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Cookie", "JSESSIONID="+js)
	req.Header.Set("Host", "account.ccnu.edu.cn")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", "https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	return len(resp.Cookies()) != 0
}

// jsessionidPattern 从 script src 中提取 jsessionid 值，大小写不敏感，遇分隔符截断。
var jsessionidPattern = regexp.MustCompile(`(?i)jsessionid=([^;?&#]*)`)

// parseCasLogin 从 CCNU CAS 登录页解析 jsessionid（js）与 lt 参数（st）。
// 页面结构不匹配或 HTML 缺失时返回 ok=false，避免对 soup 结果越界访问。
func parseCasLogin(html string) (js, st string, ok bool) {
	doc := soup.HTMLParse(html)
	casBody := doc.Find("body", "id", "cas")
	if casBody.Pointer == nil {
		return "", "", false
	}
	for _, sc := range casBody.FindAll("script") {
		if m := jsessionidPattern.FindStringSubmatch(sc.Attrs()["src"]); m != nil {
			js = m[1]
			break
		}
	}
	if js == "" {
		return "", "", false
	}
	logo := doc.Find("div", "class", "logo")
	if logo.Pointer == nil {
		return "", "", false
	}
	for _, in := range logo.FindAll("input") {
		attrs := in.Attrs()
		if attrs["name"] == "lt" && attrs["value"] != "" {
			return js, attrs["value"], true
		}
	}
	return "", "", false
}
