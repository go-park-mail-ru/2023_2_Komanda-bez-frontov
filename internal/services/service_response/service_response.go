package serviceresponse

type Response struct {
	StatusCode int
	Body       interface{}
}

func NewResponse(statusCode int, body interface{}) *Response {
	return &Response{
		StatusCode: statusCode,
		Body:       body,
	}
}
