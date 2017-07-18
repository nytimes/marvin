package readinglist

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"

	"github.com/NYTimes/marvin"
	"github.com/pkg/errors"
)

type Service interface {
	GetLinks(context.Context, *GetListProtoJSONRequest) (*Links, error)
	PutLink(context.Context, *PutLinkProtoJSONRequest) (*Message, error)
}

type linkService struct {
	db DB
}

func NewService(db DB) Service {
	return linkService{db: db}
}

const userKey = "ae-user"

func AddUser(ctx context.Context, usr *user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

func getUser(ctx context.Context) *user.User {
	return ctx.Value(userKey).(*user.User)
}

func (s linkService) GetLinks(ctx context.Context, r *GetListProtoJSONRequest) (*Links, error) {
	if r.Limit == 0 {
		r.Limit = 50
	}
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

func (s linkService) PutLink(ctx context.Context, r *PutLinkProtoJSONRequest) (*Message, error) {
	// nyt URLs only!
	if !strings.HasPrefix(r.Request.Link.Url, "https://www.nytimes.com/") {
		return nil, marvin.NewProtoStatusResponse(
			&Message{"only https://www.nytimes.com URLs accepted"},
			http.StatusBadRequest)
	}
	var err error
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
