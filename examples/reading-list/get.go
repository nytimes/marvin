package readinglist

import (
	"context"
	"net/http"
	"strconv"

	"github.com/NYTimes/marvin"
	"github.com/pkg/errors"
	"google.golang.org/appengine/log"
)

// go-kit endpoint.Endpoint with core business logic
func (s service) getLinks(ctx context.Context, req interface{}) (interface{}, error) {
	r := req.(*GetListProtoJSONRequest)

	// set request defaults
	if r.Limit == 0 {
		r.Limit = 50
	}

	// get data from the service-injected DB interface
	links, err := s.db.GetLinks(ctx, getUser(ctx).ID, int(r.Limit))
	if err != nil {
		log.Errorf(ctx, "error getting links from DB: %s", err)
		return nil, marvin.NewProtoStatusResponse(
			&Message{"server error"},
			http.StatusInternalServerError)
	}
	lks := make([]*Link, len(links))
	for i, l := range links {
		lks[i] = &Link{Url: l}
	}
	return &Links{Links: lks}, errors.Wrap(err, "unable to get links")
}

// request decoder can be used for proto and JSON since there is no body
func decodeGetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var limit int64
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}
	return &GetListProtoJSONRequest{
		Limit: int32(limit),
	}, nil
}
