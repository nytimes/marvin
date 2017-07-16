package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/NYTimes/marvin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"google.golang.org/appengine/user"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

func init() {
	marvin.Init(newService())
}

type httpService struct {
	svc readinglist.Service
}

func newService() httpService {
	return httpService{
		svc: readinglist.NewService(readinglist.NewDB()),
	}
}

// adding a service-wide error handler that can check the path
// suffix to determine how to serialize the error
func (s httpService) Options() []httptransport.ServerOption {
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
func (s httpService) RouterOptions() []marvin.RouterOption {
	return []marvin.RouterOption{
		marvin.RouterSelect("stdlib"),
	}
}

// in this example, we're tossing a simple CORS middleware in the mix
func (s httpService) HTTPMiddleware(h http.Handler) http.Handler {
	return marvin.CORSHandler(h, "")
}

func (s httpService) Middleware(ep endpoint.Endpoint) endpoint.Endpoint {
	return endpoint.Endpoint(func(ctx context.Context, r interface{}) (interface{}, error) {
		usr, err := user.CurrentOAuth(ctx, "https://www.googleapis.com/auth/userinfo.profile")
		if usr == nil || err != nil {
			// reject if user is not logged in
			return nil, marvin.NewProtoStatusResponse(
				&readinglist.Message{"please provide oauth token"},
				http.StatusUnauthorized,
			)
		}
		// add the user to the request context and continue
		return ep(readinglist.AddUser(ctx, usr), r)
	})
}

func (s httpService) JSONEndpoints() map[string]map[string]marvin.HTTPEndpoint {
	return map[string]map[string]marvin.HTTPEndpoint{
		"/link.json": {
			"PUT": {
				Endpoint: s.PutLink,
				Decoder:  decodePutRequest,
			},
		},
		"/list.json": {
			"GET": {
				Endpoint: s.GetLinks,
				Decoder:  decodeGetRequest,
			},
		},
	}
}

func (s httpService) ProtoEndpoints() map[string]map[string]marvin.HTTPEndpoint {
	return map[string]map[string]marvin.HTTPEndpoint{
		"/link.proto": {
			"PUT": {
				Endpoint: s.PutLink,
				Decoder:  decodePutProtoRequest,
			},
		},
		"/list.proto": {
			"GET": {
				Endpoint: s.GetLinks,
				Decoder:  decodeGetRequest,
			},
		},
	}
}
