package readinglist

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/NYTimes/marvin"
	"github.com/golang/protobuf/proto"
)

// go-kit endpoint.Endpoint with core business logic
func (s service) putLink(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*PutLinkProtoJSONRequest)

	// validate the request
	if !strings.HasPrefix(r.Request.Link.Url, "https://www.nytimes.com/") {
		return nil, marvin.NewProtoStatusResponse(
			&Message{"only https://www.nytimes.com URLs accepted"},
			http.StatusBadRequest)
	}

	var err error
	// call the service-injected DB interface
	if r.Request.Delete {
		err = s.db.DeleteLink(ctx, getUser(ctx).ID, r.Request.Link.Url)
	} else {
		err = s.db.PutLink(ctx, getUser(ctx).ID, r.Request.Link.Url)
	}
	if err != nil {
		return nil, marvin.NewProtoStatusResponse(
			&Message{"problems updating link"},
			http.StatusInternalServerError)
	}

	return &Message{Message: "success"}, nil
}

// JSON request decoder
func decodePutRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var lr LinkRequest
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil || lr.Link == nil {
		return nil, marvin.NewProtoStatusResponse(
			&Message{Message: "bad request"},
			http.StatusBadRequest)
	}
	return &PutLinkProtoJSONRequest{
		Request: &lr}, nil
}

// Protobuf request decoder
func decodePutProtoRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, marvin.NewProtoStatusResponse(
			&Message{Message: "unable to read request"},
			http.StatusBadRequest)
	}
	var lr LinkRequest
	err = proto.Unmarshal(b, &lr)
	if err != nil {
		return nil, marvin.NewProtoStatusResponse(
			&Message{Message: "bad request"},
			http.StatusBadRequest)
	}
	return &PutLinkProtoJSONRequest{
		Request: &lr}, nil
}
