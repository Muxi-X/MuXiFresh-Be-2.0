package ccnusvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type Req struct {
	StudentId string `json:"student_id"`
	Password  string `json:"password"`
}

//func (c *ccnuService) Login(ctx context.Context, studentId string, password string) (bool, error) {
//	client, err := c.loginClient(ctx, studentId, password)
//	fmt.Println(err.Error())
//	return client != nil, err
//}

func (c *ccnuService) Login(ctx context.Context, studentId string, password string) (bool, error) {
	url := "https://kstack.muxixyz.com/users/login_ccnu"
	method := "POST"
	param := &Req{
		StudentId: studentId,
		Password:  password,
	}
	raw, err := json.Marshal(param)
	if err != nil {
		return false, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(raw))

	if err != nil {
		return false, err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "kstack.muxixyz.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	fmt.Println(string(body))
	return strings.Contains(string(body), "Success"), nil
}

func (c *ccnuService) client() *http.Client {
	j, _ := cookiejar.New(&cookiejar.Options{})
	return &http.Client{
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Jar:     j,
		Timeout: c.timeout,
	}
}

func (c *ccnuService) loginClient(ctx context.Context, studentId string, password string) (*http.Client, error) {
	params, err := c.makeAccountPreflightRequest()
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("username", studentId)
	v.Set("password", password)
	v.Set("lt", params.lt)
	v.Set("execution", params.execution)
	v.Set("_eventId", params._eventId)
	v.Set("submit", params.submit)

	request, err := http.NewRequest("POST", "https://account.ccnu.edu.cn/cas/login/cas/login;jsessionid="+params.JSESSIONID, strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	request.WithContext(ctx)

	client := c.client()
	resp, err := client.Do(request)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return nil, ccnuv1.ErrorNetworkToXkError("网络异常")
		}
		return nil, err
	}
	if len(resp.Header.Get("Set-Cookie")) == 0 {
		return nil, ccnuv1.ErrorInvalidSidOrPwd("学号或密码错误")
	}
	return client, nil
}

type ClientKey struct{} // 用于 context 的键

// 将 http.Client 添加到 context 中
func (c *ccnuService) addClientToContext(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, ClientKey{}, client)
}

// 从 context 中获取 http.Client
func (c *ccnuService) getClientFromContext(ctx context.Context) *http.Client {
	client, ok := ctx.Value(ClientKey{}).(*http.Client)
	if !ok {
		return nil // 这里可以处理默认逻辑或错误
	}
	return client
}
