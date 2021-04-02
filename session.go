package requests

import "net/http"

type Session struct {
	Request
	cookies *http.Cookie
}

func NewSession() *Session {
	return &Session{}
}

func (session *Session) Get(requestUrl string, args ...interface{}) (*Response, error) {
	resp, err := session.get(requestUrl, args...)
	if err != nil {
		return nil, err
	}
	if session.Cookie == nil{
		session.Cookie = make(map[string]string)
	}
	cookies := resp.Cookie
	for _, cook := range cookies {
		session.Cookie[cook.Name] = cook.Value
	}
	return resp, err
}

func (session *Session) Post(requestUrl string, args ...interface{}) (*Response, error) {
	resp, err := session.post(requestUrl, args...)
	if err != nil {
		return nil, err
	}
	if session.Cookie == nil{
		session.Cookie = make(map[string]string)
	}
	cookies := resp.Cookie
	for _, cook := range cookies {
		session.Cookie[cook.Name] = cook.Value
	}
	return resp, err
}
