package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/NYTimes/marvin"
	"github.com/golang/protobuf/proto"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

func (s httpService) PutLink(ctx context.Context, r interface{}) (interface{}, error) {
	return s.svc.PutLink(ctx, r.(*readinglist.PutLinkProtoJSONRequest))
}

func decodePutRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var lr readinglist.LinkRequest
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil || lr.Link == nil {
		return nil, marvin.NewProtoStatusResponse(
			&readinglist.Message{Message: "bad request"},
			http.StatusBadRequest)
	}
	return &readinglist.PutLinkProtoJSONRequest{
		Request: &lr}, nil
}

func decodePutProtoRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, marvin.NewProtoStatusResponse(
			&readinglist.Message{Message: "unable to read request"},
			http.StatusBadRequest)
	}
	var lr readinglist.LinkRequest
	err = proto.Unmarshal(b, &lr)
	if err != nil {
		return nil, marvin.NewProtoStatusResponse(
			&readinglist.Message{Message: "bad request"},
			http.StatusBadRequest)
	}
	return &readinglist.PutLinkProtoJSONRequest{
		Request: &lr}, nil
}
