package requests

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Headers map[string]string
type Params map[string]string
type Cookie map[string]string
type Data map[string]string
type Json map[string]string
type Proxies map[string]string
type Verify bool
type AllowRedirects bool

type Request struct {
	//httpreq        *http.Request
	Url            string
	Method         string
	Headers        map[string]string
	Params         map[string]string
	Cookie         map[string]string
	Proxies        map[string]string
	Data           map[string]string
	Json           map[string]string
	Timeout        time.Duration
	Verify         Verify
	AllowRedirects AllowRedirects

	client *http.Client
	//cookies []*http.Cookie
}

type Response struct {
	R       *http.Response
	Cookie  http.CookieJar
	Request *Request
	Content []byte
	Text    string
}

func newClient(verify Verify, allowRedirects AllowRedirects, timeout time.Duration, proxies map[string]string) *http.Client {
	client := &http.Client{}
	tr := &http.Transport{}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	proxy := setProxy(proxies)
	if proxy != nil {
		tr.Proxy = http.ProxyURL(proxy)
	}
	client.Transport = tr
	if allowRedirects == true {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("Don't redirect!")
		}
	}
	client.Timeout = timeout
	return client
}

func (request *Request)setCookie(requestUrl string, c map[string]string)  {
	cookie, _ := cookiejar.New(nil)
	cookieList := make([]*http.Cookie, 0)
	u, err := url.Parse(requestUrl)
	if err != nil {
		return
	}
	if request.client.Jar == nil{
		cookie.SetCookies(u, cookieList)
		request.client.Jar = cookie
	}
	if c != nil{
		for k, v := range c {
			httpCookie := &http.Cookie{
				Name:     k,
				Value:    v,
				HttpOnly: false,
			}
			cookieList = append(cookieList, httpCookie)
		}
		cookie.SetCookies(u, cookieList)
		request.client.Jar = cookie
	}
}

func (request *Request) parseArgs(args ...interface{}) {

	for _, arg := range args {
		switch ty := arg.(type) {
		case Headers:
			request.Headers = ty
		case Params:
			request.Params = ty
		case Cookie:
			request.Cookie = ty
		case Proxies:
			request.Proxies = ty
		case Data:
			request.Data = ty
		case Json:
			request.Json = ty
		case Verify:
			request.Verify = ty
		case AllowRedirects:
			request.AllowRedirects = ty
		case time.Duration:
			request.Timeout = ty
		}
	}
}

func (request *Request) baseSend(requestUrl, method string, args ...interface{}) (*Response, error) {
	var err error
	request.Url = requestUrl
	request.Method = method
	if request.Headers == nil{
		request.Headers = make(map[string]string)
	}
	request.parseArgs(args...)
	if request.client == nil{
		client := newClient(request.Verify, request.AllowRedirects, request.Timeout, request.Proxies)
		request.client = client
	}
	request.setCookie(requestUrl, request.Cookie)

	requestUrl, err = buildURLParams(requestUrl, request.Params)
	if err != nil {
		return nil, err
	}
	var httpReq *http.Request
	switch method {
	case "GET":
		httpReq, err = http.NewRequest("GET", requestUrl, nil)
	case "POST":
		if request.Data != nil {
			body := buildForms(request.Data)
			httpReq, err = http.NewRequest("POST", requestUrl, body)
			request.Headers["Content-Type"] = "application/x-www-form-urlencoded"
		} else if request.Json != nil {
			jsonStr, _ := json.Marshal(request.Json)
			httpReq, err = http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonStr))
			request.Headers["Content-Type"] = "application/json;charset=utf-8"
		}

	}
	if err != nil {
		return nil, err
	}
	for key, value := range request.Headers {
		httpReq.Header.Add(key, value)
	}
	resp, err := request.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	response := &Response{}
	response.R = resp
	response.Cookie = request.client.Jar
	response.Content = response.setContent()
	response.Text = response.setText()
	return response, nil

}

func (request *Request) get(requestUrl string, args ...interface{}) (*Response, error) {
	return request.baseSend(requestUrl, "GET", args...)
}

func (request *Request) post(requestUrl string, args ...interface{}) (*Response, error) {
	return request.baseSend(requestUrl, "POST", args...)
}

func buildURLParams(userURL string, params map[string]string) (string, error) {
	parsedURL, err := url.Parse(userURL)

	if err != nil {
		return "", err
	}

	parsedQuery, err := url.ParseQuery(parsedURL.RawQuery)

	if err != nil {
		return "", nil
	}

	for key, value := range params {
		parsedQuery.Add(key, value)
	}
	return addQueryParams(parsedURL, parsedQuery), nil
}

func addQueryParams(parsedURL *url.URL, parsedQuery url.Values) string {
	if len(parsedQuery) > 0 {
		return strings.Join([]string{strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1), parsedQuery.Encode()}, "?")
	}
	return strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1)
}

func buildForms(m map[string]string) *strings.Reader {
	postValue := url.Values{}
	for k, v := range m {
		postValue.Add(k,  v)
	}
	postString := postValue.Encode()
	return  strings.NewReader(postString)
}

func setProxy(proxies map[string]string) *url.URL {
	if proxies == nil {
		return nil
	}
	var proxy string
	for key, value := range proxies {
		proxy = key + "://" + value
	}
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		return nil
	}
	return proxyUrl
}

func (resp *Response) setContent() []byte {
	defer resp.R.Body.Close()
	var err error
	var Body = resp.R.Body
	if resp.R.Header.Get("Content-Encoding") == "gzip" && resp.R.Header.Get("Accept-Encoding") != "" {
		// fmt.Println("gzip")
		reader, err := gzip.NewReader(Body)
		if err != nil {
			return nil
		}
		Body = reader
	}
	content, err := ioutil.ReadAll(Body)
	if err != nil {
		return nil
	}
	resp.Content = content
	return content
}

func (resp *Response) setText() string {
	if resp.Content == nil {
		resp.setContent()
	}
	resp.Text = string(resp.Content)
	return resp.Text
}
