package readinglist

import (
	"context"
	"strings"

	"google.golang.org/appengine/datastore"

	"github.com/pkg/errors"
	ocontext "golang.org/x/net/context"
)

type DB interface {
	GetLinks(ctx context.Context, userID string, limit int) ([]string, error)
	PutLink(ctx context.Context, userID string, url string) error
	DeleteLink(ctx context.Context, userID string, url string) error
}

type Datastore struct{}

const LinkKind = "Link"

func NewDB() Datastore {
	return Datastore{}
}

type linkData struct {
	UserID string
	URL    string `datastore:",noindex"`
}

func newKey(ctx context.Context, userID, url string) *datastore.Key {
	skey := strings.TrimPrefix(url, "https://www.nytimes.com/")
	return datastore.NewKey(ctx, LinkKind, reverse(userID)+"-"+skey, 0, nil)
}

func (d Datastore) GetLinks(ctx context.Context, userID string, limit int) ([]string, error) {
	var datas []*linkData
	_, err := datastore.NewQuery(LinkKind).Filter("UserID =", userID).
		Limit(limit).GetAll(ctx, &datas)
	links := make([]string, len(datas))
	for i, d := range datas {
		links[i] = d.URL
	}
	return links, errors.WithMessage(err, "unable to query links")
}

func (d Datastore) DeleteLink(ctx context.Context, userID string, url string) error {
	err := datastore.Delete(ctx, newKey(ctx, userID, url))
	return errors.Wrap(err, "unable to delete url")
}

func (d Datastore) PutLink(ctx context.Context, userID string, url string) error {
	key := newKey(ctx, userID, url)

	// run in transaction to avoid any dupes
	err := datastore.RunInTransaction(ctx, func(ctx ocontext.Context) error {
		var existing linkData
		err := datastore.Get(ctx, key, &existing)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return errors.Wrap(err, "unable to check if link already exists")
		}
		// link already exists, just return
		if err != datastore.ErrNoSuchEntity {
			return nil
		}
		// put new link
		_, err = datastore.Put(ctx, key, &linkData{
			UserID: userID,
			URL:    url,
		})
		return err
	}, nil)
	return errors.Wrap(err, "unable to put link")
}

// Using this to turn keys 12345, 12346 into 54321, 64321 which are easier for
// Datastore/BigTable to shard.
func reverse(id string) string {
	runes := []rune(id)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}
