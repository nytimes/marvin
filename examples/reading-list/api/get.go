package api

import (
	"context"
	"net/http"
	"strconv"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

func (s httpService) GetLinks(ctx context.Context, r interface{}) (interface{}, error) {
	return s.svc.GetLinks(ctx, r.(*readinglist.GetListProtoJSONRequest))
}

func decodeGetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var limit int64
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}
	return &readinglist.GetListProtoJSONRequest{
		Limit: int32(limit),
	}, nil
}
