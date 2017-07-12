package readinglist

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/appengine/datastore"
)

type DB interface {
	GetLinks(ctx context.Context, userID string, limit int) ([]*Link, error)
	PutLink(ctx context.Context, userID string, link *Link) error
}

type Datastore struct{}

const LinkKind = "Link"

func NewDB() Datastore {
	return Datastore{}
}

type linkData struct {
	UserID string
	Link   Link
}

func (d Datastore) GetLinks(ctx context.Context, userID string, limit int) ([]*Link, error) {
	var datas []*linkData
	_, err := datastore.NewQuery(LinkKind).Filter("UserID =", userID).
		Limit(limit).GetAll(ctx, &datas)
	links := make([]*Link, len(datas))
	for i, d := range datas {
		links[i] = &d.Link
	}
	return links, errors.WithMessage(err, "unable to query links")
}

func (d Datastore) PutLink(ctx context.Context, userID string, link *Link) error {
	id, _, err := datastore.AllocateIDs(ctx, LinkKind, nil, 1)
	if err != nil {
		return errors.WithMessage(err, "unable to allocate IDs")
	}
	link.Id = userID + "-" + strconv.FormatInt(id, 10)
	key := datastore.NewKey(ctx, LinkKind, link.Id, 0, nil)
	_, err = datastore.Put(ctx, key, &linkData{
		UserID: userID,
		Link:   *link,
	})
	return errors.WithMessage(err, "unable to put link")
}
