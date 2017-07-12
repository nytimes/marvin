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
	return &Links{Links: links}, errors.WithMessage(err, "unable to get links")
}

func (s linkService) PutLink(ctx context.Context, r *PutLinkProtoJSONRequest) (*Message, error) {
	err := s.db.PutLink(ctx, r.UserID, r.Link)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to save link")
	}
	return &Message{Message: "success"}, nil
}
