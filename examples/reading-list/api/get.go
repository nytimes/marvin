package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/NYTimes/marvin"
	"google.golang.org/appengine/log"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

func (s httpService) GetLinks(ctx context.Context, r interface{}) (interface{}, error) {
	res, err := s.svc.GetLinks(ctx, r.(*readinglist.GetListProtoJSONRequest))
	if err != nil {
		log.Errorf(ctx, "unable to get links: %s", err)
		return nil, marvin.NewProtoStatusResponse(
			&readinglist.Message{Message: "unable to update link"},
			http.StatusInternalServerError)
	}
	return res, nil
}

func decodeGetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var limit int64
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}
	return &readinglist.GetListProtoJSONRequest{
		Limit:  int32(limit),
		UserID: getUserID(ctx),
	}, nil
}
