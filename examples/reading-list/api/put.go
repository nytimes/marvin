package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/NYTimes/marvin"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

func (s httpService) PutLink(ctx context.Context, r interface{}) (interface{}, error) {
	return s.svc.PutLink(ctx, r.(*readinglist.PutLinkProtoJSONRequest))
}

func decodePutRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var ls readinglist.Link
	err := json.NewDecoder(r.Body).Decode(&ls)
	if err != nil {
		return nil, marvin.NewProtoStatusResponse(
			&readinglist.Message{Message: "these are some bad cats"},
			http.StatusBadRequest)
	}
	r.Body.Close()

	return &readinglist.PutLinkProtoJSONRequest{
		UserID: getUserID(ctx),
		Link:   &ls}, nil
}
