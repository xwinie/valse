package valse

import "github.com/valyala/fasthttp"

//Entity 结构体
type Entity struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

// New new entity
func (r Entity) New(newCode int, newMsg string) Entity {
	r.Msg = newMsg
	r.Code = newCode
	return r
}

//NewHTTPMessage new http message
func NewHTTPMessage(code int, msg ...string) error {
	m := StatusText(code)
	if len(msg) != 0 {
		m = msg[0]
	}

	return &Entity{code, m}
}

// WithMsg set msg
func (r *Entity) WithMsg(newMsg string) *Entity {
	r.Msg = newMsg
	return r
}

// WithAttachMsg  添加msg
func (r *Entity) WithAttachMsg(newMsg string) *Entity {
	r.Msg = r.Msg + newMsg
	return r
}

//WithCode set code
func (r *Entity) WithCode(newCode int) *Entity {
	r.Code = newCode
	return r
}

//MarshalJSON entity json
func (r *Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"code":    r.Code,
		"message": r.Msg,
	})
}

func (r *Entity) Error() string {
	return r.Msg
}

//Message 当前消息
func (r *Entity) Message() string {
	return r.Msg
}

//EntityCode 当前code
func (r *Entity) EntityCode() int {
	return r.Code
}

//BuildEntity 新建实体函数
func BuildEntity(newCode int, newMsg string) *Entity {
	return &Entity{newCode, newMsg}
}

//ResponseEntity 返回实体
type ResponseEntity struct {
	StatusCode int
	Data       interface{}
}

//NewBuild 新建实体
func (r *ResponseEntity) NewBuild(StatusCode int, Data interface{}) *ResponseEntity {
	r.StatusCode = StatusCode
	r.Data = Data
	return r
}

//Build 无code实体
func (r *ResponseEntity) Build(Data interface{}) *ResponseEntity {
	r.StatusCode = fasthttp.StatusOK
	r.Data = Data
	return r
}

//BuildError 错误实体
func (r *ResponseEntity) BuildError(Data interface{}) *ResponseEntity {
	r.StatusCode = fasthttp.StatusBadRequest
	r.Data = Data
	return r
}

//BuildFormatError 格式化实体错误
func (r *ResponseEntity) BuildFormatError(Data interface{}) *ResponseEntity {
	r.StatusCode = fasthttp.StatusNotAcceptable
	r.Data = Data
	return r
}

//BuildPostAndPut post和put实体
func (r *ResponseEntity) BuildPostAndPut(Data interface{}) *ResponseEntity {
	r.StatusCode = fasthttp.StatusCreated
	r.Data = Data
	return r
}

//BuildDelete Response delete删除实体
func (r *ResponseEntity) BuildDelete(Data interface{}) *ResponseEntity {
	r.StatusCode = fasthttp.StatusNoContent
	r.Data = Data
	return r
}

//BuildDeleteGone StatusGone状态实体
func (r *ResponseEntity) BuildDeleteGone(Data interface{}) *ResponseEntity {
	r.StatusCode = fasthttp.StatusGone
	r.Data = Data
	return r
}
