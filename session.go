package requests


type Session struct {
	Request
}

func NewSession() *Session {
	return &Session{}
}

func (session *Session) Get(requestUrl string, args ...interface{}) (*Response, error) {
	resp, err := session.get(requestUrl, args...)
	if err != nil {
		return nil, err
	}
	session.client.Jar = resp.Cookie
	return resp, err
}

func (session *Session) Post(requestUrl string, args ...interface{}) (*Response, error) {
	resp, err := session.post(requestUrl, args...)
	if err != nil {
		return nil, err
	}
	session.client.Jar = resp.Cookie
	return resp, err
}
