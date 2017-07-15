package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/NYTimes/marvin"
	"github.com/golang/protobuf/proto"
	"google.golang.org/appengine/log"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

func (s httpService) PutLink(ctx context.Context, r interface{}) (interface{}, error) {
	res, err := s.svc.PutLink(ctx, r.(*readinglist.PutLinkProtoJSONRequest))
	if err != nil {
		log.Errorf(ctx, "unable to update link: %s", err)
		return nil, marvin.NewProtoStatusResponse(
			&readinglist.Message{Message: "unable to update link"},
			http.StatusInternalServerError)
	}
	return res, nil
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
		UserID:  getUserID(ctx),
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
		UserID:  getUserID(ctx),
		Request: &lr}, nil
}
