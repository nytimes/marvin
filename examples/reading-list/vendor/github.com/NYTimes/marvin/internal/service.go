package internal

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
)

var NewContext = func(r *http.Request) context.Context {
	return appengine.NewContext(r)
}
