package pojo

type Response struct {
	Body struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg,omitempty"`
		Data interface{} `json:"data,omitempty"`
	}
}

func NewResponse(Code int, Msg string, Data interface{}) *Response {
	responseBody := new(Response)
	responseBody.Body.Code = Code
	responseBody.Body.Msg = Msg
	responseBody.Body.Data = Data
	return responseBody
}
