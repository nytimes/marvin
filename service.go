package marvin

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// HTTPEndpoint encapsulates everything required to build
// an endpoint hosted on an HTTP server.
type HTTPEndpoint struct {
	Endpoint endpoint.Endpoint
	Decoder  httptransport.DecodeRequestFunc
	Encoder  httptransport.EncodeResponseFunc
	Options  []httptransport.ServerOption
}

// Service is the most basic interface of a service that can be received and
// hosted by Google App Engine.
// By default, if no warmup request is provided, marvin will add a basic
// handler to the /_ah/warmup route for you.
type Service interface {
	// HTTPMiddleware is for service-wide http specific middlewares
	// for easy integration with 3rd party http.Handlers.
	HTTPMiddleware(http.Handler) http.Handler

	// Middleware is for any service-wide go-kit middlewares
	Middleware(endpoint.Endpoint) endpoint.Endpoint

	// Options are service-wide go-kit options
	Options() []httptransport.ServerOption

	// RouterOptions allows users to override the default
	// behavior and use of the GorillaRouter.
	RouterOptions() []RouterOption
}

// JSONService endpoints are for HTTP endpoints with JSON serialization.
// This service will add default decoder/encoders if none provided.
type JSONService interface {
	Service
	JSONEndpointer
}

// ProtoService endpoints are for HTTP endpoints with Protobuf serialization.
// This service will add default decoder/encoders if none provided.
type ProtoService interface {
	Service
	ProtoEndpointer
}

// MixedService combines the Proto and JSON services to allow users to
// expose endpoints on both JSON and Protobuf.
type MixedService interface {
	Service
	JSONEndpointer
	ProtoEndpointer
}

// JSONEndpointer is for HTTP endpoints with JSON serialization
// The first map's string is for the HTTP path, the second is for the http method.
// For example:
//
//	return map[string]map[string]marvin.HTTPEndpoint{
//		"/cat/{id}.json": {
//			"GET": {
//				Endpoint: s.GetCatByID,
//				Decoder:  decodeGetCatRequest,
//			},
//		},
//		"/cats.json": {
//			"PUT": {
//				Endpoint: s.PutCats,
//				Decoder:  decodePutCatsProtoRequest,
//			},
//			"GET": {
//				Endpoint: s.GetCats,
//				Decoder:  decodeGetCatsRequest,
//			},
//		},
//	}
//
type JSONEndpointer interface {
	JSONEndpoints() map[string]map[string]HTTPEndpoint
}

// ProtoEndpointer is for HTTP endpoints with protobuf serialization.
// The first map's string is for the HTTP path, the second is for the http method.
// For example:
//
//	return map[string]map[string]marvin.HTTPEndpoint{
//		"/cat/{id}.proto": {
//			"GET": {
//				Endpoint: s.GetCatByID,
//				Decoder:  decodeGetCatRequest,
//			},
//		},
//		"/cats.proto": {
//			"PUT": {
//				Endpoint: s.PutCats,
//				Decoder:  decodePutCatsRequest,
//			},
//			"GET": {
//				Endpoint: s.GetCats,
//				Decoder:  decodeGetCatsRequest,
//			},
//		},
//	}
//
type ProtoEndpointer interface {
	ProtoEndpoints() map[string]map[string]HTTPEndpoint
}
