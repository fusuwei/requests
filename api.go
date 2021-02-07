package requests

func Get(requestUrl string, args ...interface{}) (*Response, error) {
	req := &Request{}
	return req.get(requestUrl, args...)
}

func Post(requestUrl string, args ...interface{}) (*Response, error) {
	req := &Request{}
	return req.post(requestUrl, args...)
}
