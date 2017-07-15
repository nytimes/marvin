package marvin

import (
	"encoding/json"
	"net/http"

	"github.com/golang/protobuf/proto"
)

// NewProtoStatusResponse allows users to respond with a specific HTTP status code and
// a Protobuf or JSON serialized response.
func NewProtoStatusResponse(res proto.Message, code int) *ProtoStatusResponse {
	return &ProtoStatusResponse{res: res, code: code}
}

// ProtoStatusResponse implements:
// `httptransport.StatusCoder` to allow users to respond with the given
// response with a non-200 status code.
// `proto.Marshaler` and proto.Message so it can wrap a proto Endpoint responses.
// `json.Marshaler` so it can wrap JSON Endpoint responses.
// `error` so it can be used to respond as an error within the go-kit stack.
type ProtoStatusResponse struct {
	code int
	res  proto.Message
}

// to implement httptransport.StatusCoder
func (c *ProtoStatusResponse) StatusCode() int {
	return c.code
}

// to implement proto.Marshaler
func (c *ProtoStatusResponse) Marshal() ([]byte, error) {
	return proto.Marshal(c.res)
}

// to implement proto.Message
func (c *ProtoStatusResponse) Reset()         { c.res.Reset() }
func (c *ProtoStatusResponse) String() string { return c.res.String() }
func (c *ProtoStatusResponse) ProtoMessage()  { c.res.ProtoMessage() }

// to implement json.Marshaler
func (c *ProtoStatusResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.res)
}

var _ proto.Marshaler = &ProtoStatusResponse{}

// to implement error
func (c *ProtoStatusResponse) Error() string {
	return http.StatusText(c.code)
}

// NewJSONStatusResponse allows users to respond with a specific HTTP status code and
// a JSON serialized response.
func NewJSONStatusResponse(res interface{}, code int) *JSONStatusResponse {
	return &JSONStatusResponse{res: res, code: code}
}

// JSONStatusResponse implements:
// `httptransport.StatusCoder` to allow users to respond with the given
// response with a non-200 status code.
// `json.Marshaler` so it can wrap JSON Endpoint responses.
// `error` so it can be used to respond as an error within the go-kit stack.
type JSONStatusResponse struct {
	code int
	res  interface{}
}

// StatusCode is to implement httptransport.StatusCoder
func (c *JSONStatusResponse) StatusCode() int {
	return c.code
}

// MarshalJSON is to implement json.Marshaler
func (c *JSONStatusResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.res)
}

// Error is to implement error
func (c *JSONStatusResponse) Error() string {
	return http.StatusText(c.code)
}
