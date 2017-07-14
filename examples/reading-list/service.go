package readinglist

import (
	"context"

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

func (s linkService) GetLinks(ctx context.Context, r *GetListProtoJSONRequest) (*Links, error) {
	if r.Limit == 0 {
		r.Limit = 50
	}
	links, err := s.db.GetLinks(ctx, r.UserID, int(r.Limit))
	lks := make([]*Link, len(links))
	for i, l := range links {
		lks[i] = &Link{Url: l}
	}
	return &Links{Links: lks}, errors.WithMessage(err, "unable to get links")
}

func (s linkService) PutLink(ctx context.Context, r *PutLinkProtoJSONRequest) (*Message, error) {
	var err error
	if r.Request.Delete {
		err = s.db.DeleteLink(ctx, r.UserID, r.Request.Link.Url)
	} else {
		err = s.db.PutLink(ctx, r.UserID, r.Request.Link.Url)
	}
	if err != nil {
		return nil, errors.Wrap(err, "problems updating link")
	}
	return &Message{Message: "success"}, nil
}
