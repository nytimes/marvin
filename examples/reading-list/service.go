package readinglist

import (
	"context"
	"net/http"
	"strings"

	"github.com/NYTimes/marvin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"

	"google.golang.org/appengine/user"
)

type service struct {
	db DB
}

func NewService(db DB) marvin.MixedService {
	return service{db: db}
}

// adding a service-wide error handler that can check the path
// suffix to determine how to serialize the error
func (s service) Options() []httptransport.ServerOption {
	return []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(func(ctx context.Context, err error, w http.ResponseWriter) {
			// check proto/json by inspecting url
			path := ctx.Value(httptransport.ContextKeyRequestPath).(string)
			if strings.HasSuffix(path, ".json") {
				httptransport.EncodeJSONResponse(ctx, w, err)
				return
			}
			marvin.EncodeProtoResponse(ctx, w, err)
		}),
	}
}

// override the default gorilla router and select the stdlib
func (s service) RouterOptions() []marvin.RouterOption {
	return []marvin.RouterOption{
		marvin.RouterSelect("stdlib"),
	}
}

// in this example, we're tossing a simple CORS middleware in the mix
func (s service) HTTPMiddleware(h http.Handler) http.Handler {
	return marvin.CORSHandler(h, "")
}

// the go-kit middleware is used for checking user authentication and
// injecting the current user into the request context.
func (s service) Middleware(ep endpoint.Endpoint) endpoint.Endpoint {
	return endpoint.Endpoint(func(ctx context.Context, r interface{}) (interface{}, error) {
		usr, err := user.CurrentOAuth(ctx, "https://www.googleapis.com/auth/userinfo.profile")
		if usr == nil || err != nil {
			// reject if user is not logged in
			return nil, marvin.NewProtoStatusResponse(
				&Message{"please provide oauth token"},
				http.StatusUnauthorized,
			)
		}
		// add the user to the request context and continue
		return ep(addUser(ctx, usr), r)
	})
}

func (s service) JSONEndpoints() map[string]map[string]marvin.HTTPEndpoint {
	return map[string]map[string]marvin.HTTPEndpoint{
		"/link.json": {
			"PUT": {
				Endpoint: s.putLink,
				Decoder:  decodePutRequest,
			},
		},
		"/list.json": {
			"GET": {
				Endpoint: s.getLinks,
				Decoder:  decodeGetRequest,
			},
		},
	}
}

func (s service) ProtoEndpoints() map[string]map[string]marvin.HTTPEndpoint {
	return map[string]map[string]marvin.HTTPEndpoint{
		"/link.proto": {
			"PUT": {
				Endpoint: s.putLink,
				Decoder:  decodePutProtoRequest,
			},
		},
		"/list.proto": {
			"GET": {
				Endpoint: s.getLinks,
				Decoder:  decodeGetRequest,
			},
		},
	}
}

const userKey = "ae-user"

func addUser(ctx context.Context, usr *user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

func getUser(ctx context.Context) *user.User {
	return ctx.Value(userKey).(*user.User)
}
