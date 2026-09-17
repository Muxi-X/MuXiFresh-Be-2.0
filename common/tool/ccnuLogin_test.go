package tool

import (
	"fmt"
	"github.com/anaskhan96/soup"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"
)

func TestCCNULogin(t *testing.T) {
	t.Skip("手动调试脚本：依赖真实 CCNU 网络与写死账号，默认跳过")
	username := "xxx"
	password := "xxx"
	resp, _ := soup.Get("https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	js, st, ok := parseCasLogin(resp)
	if !ok {
		t.Fatal("parseCasLogin failed")
	}
	jar, _ := cookiejar.New(&cookiejar.Options{})

	client := &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://account.ccnu.edu.cn/cas/login;jsessionid=%v?service=http", js) + "%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal"
	text := fmt.Sprintf("username=%v&password=%v&lt=%v&execution=e1s1&_eventId=submit&submit=", username, password, st) + "%E7%99%BB%E5%BD%95"
	body := strings.NewReader(text)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Cookie", "JSESSIONID="+js)
	req.Header.Set("Host", "account.ccnu.edu.cn")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", "https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fone.ccnu.edu.cn%2Fcas%2Flogin_portal")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, _ := client.Do(req)
	fmt.Println(len(res.Cookies()))
}

func TestParseCasLogin(t *testing.T) {
	validHTML := `<html><body id="cas">` +
		`<script src="//cdnjs.cloudflare.com/ajax/libs/html5shiv/3.6.1/html5shiv.js"></script>` +
		`<script src="/cas/js/jquery.min.js;jsessionid=SESSION123"></script>` +
		`<script src="/cas/js/cas.js;jsessionid=SESSION123"></script>` +
		`<div class="logo"><input name="username"/><input name="password"/><input name="lt" value="LT-12345"/></div>` +
		`</body></html>`

	captchaHTML := `<html><body id="cas">` +
		`<script src="/cas/js/jquery.min.js;jsessionid=SESSION123"></script>` +
		`<script src="/cas/js/cas.js;jsessionid=SESSION123"></script>` +
		`<div class="logo"><input name="username"/><input name="password"/><input name="captcha"/><input name="lt" value="LT-999"/></div>` +
		`</body></html>`

	cases := []struct {
		name   string
		html   string
		ok     bool
		wantJS string
		wantST string
	}{
		{"empty", "", false, "", ""},
		{"short", "<html><body></body></html>", false, "", ""},
		{"no-cas", "<html><body><div>hello</div></body></html>", false, "", ""},
		{"no-jsessionid", `<html><body id="cas"><script src="/cas/js/cas.js"></script><div class="logo"><input name="lt" value="LT-1"/></div></body></html>`, false, "", ""},
		{"empty-jsessionid", `<html><body id="cas"><script src="/cas/js/cas.js;jsessionid="></script><div class="logo"><input name="lt" value="LT-1"/></div></body></html>`, false, "", ""},
		{"jsessionid-trailing", `<html><body id="cas"><script src="/cas/js/cas.js;jsessionid=S1;other=2"></script><div class="logo"><input name="lt" value="LT-1"/></div></body></html>`, true, "S1", "LT-1"},
		{"jsessionid-uppercase", `<html><body id="cas"><script src="/cas/js/cas.js;JSESSIONID=S9"></script><div class="logo"><input name="lt" value="LT-1"/></div></body></html>`, true, "S9", "LT-1"},
		{"jsessionid-query", `<html><body id="cas"><script src="/cas/js/cas.js;jsessionid=S1?foo=1"></script><div class="logo"><input name="lt" value="LT-1"/></div></body></html>`, true, "S1", "LT-1"},
		{"jsessionid-boundary", `<html><body id="cas"><script src="/cas/js/foojsessionid=BAD;jsessionid=GOOD"></script><div class="logo"><input name="lt" value="LT-1"/></div></body></html>`, true, "GOOD", "LT-1"},
		{"no-logo", `<html><body id="cas"><script src="/cas/js/cas.js;jsessionid=S1"></script><div>no logo</div></body></html>`, false, "", ""},
		{"no-lt", `<html><body id="cas"><script src="/cas/js/cas.js;jsessionid=S1"></script><div class="logo"><input/><input/><input value="x"/></div></body></html>`, false, "", ""},
		{"empty-lt-value", `<html><body id="cas"><script src="/cas/js/cas.js;jsessionid=S1"></script><div class="logo"><input name="lt" value=""/></div></body></html>`, false, "", ""},
		{"lt-no-value-attr", `<html><body id="cas"><script src="/cas/js/cas.js;jsessionid=S1"></script><div class="logo"><input name="lt"/></div></body></html>`, false, "", ""},
		{"valid", validHTML, true, "SESSION123", "LT-12345"},
		{"captcha-shift", captchaHTML, true, "SESSION123", "LT-999"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			js, st, ok := parseCasLogin(tc.html)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v (js=%q st=%q)", tc.ok, ok, js, st)
			}
			if tc.ok && (js != tc.wantJS || st != tc.wantST) {
				t.Fatalf("expected js=%q st=%q, got js=%q st=%q", tc.wantJS, tc.wantST, js, st)
			}
		})
	}
}
