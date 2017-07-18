package marvin

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
	"google.golang.org/appengine"
)

var defaultDenial = NewJSONStatusResponse(
	map[string]string{"msg": "unauthorized"},
	http.StatusUnauthorized)

// Internal is a middleware handler meant to mark an endpoint or service for
// service-to-service use only. If the incoming request does not contain an
// 'X-Appengine-Inbound-Appid' header that matches the AppID of the current service,
// this handler will return with with the given denial response.
//
// If no denial is given, the server will respond with a 401 status code and a simple
// JSON response. If you supply your own denial, we recommend you use the Proto/JSONStatusResponse
// structs to respond with a specific status code and the appropriate serialization.
//
// More info on the 'X-Appengine-Inbound-Appid' header here:
// https://cloud.google.com/appengine/docs/standard/go/appidentity/#asserting_identity_to_other_app_engine_apps
func Internal(ep endpoint.Endpoint, denial error) endpoint.Endpoint {
	if denial == nil {
		denial = defaultDenial
	}
	return endpoint.Endpoint(func(ctx context.Context, r interface{}) (interface{}, error) {
		// only accept requests from our app
		if ctx.Value(ContextKeyInboundAppID).(string) != appengine.AppID(ctx) {
			return nil, denial
		}
		return ep(ctx, r)
	})
}

// CORSHandler is a middleware func for setting all headers that enable CORS.
// If an originSuffix is provided, a strings.HasSuffix check will be performed
// before adding any CORS header. If an empty string is provided, any Origin
// header found will be placed into the CORS header. If no Origin header is
// found, no headers will be added.
func CORSHandler(f http.Handler, originSuffix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" &&
			(originSuffix == "" || strings.HasSuffix(origin, originSuffix)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-requested-by, *")
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
		}
		// blanket response for all OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		f.ServeHTTP(w, r)
	})
}

// ParseIPNets will accept a comma delimited list of CIDR blocks, parse them and
// return a slice of net.IPNets.
func ParseIPNets(ipStr string) ([]*net.IPNet, error) {
	var ipnets []*net.IPNet
	if ipStr != "" {
		ips := strings.Split(ipStr, ",")
		for _, ip := range ips {
			_, ipnet, err := net.ParseCIDR(ip)
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse CIDR string")
			}
			ipnets = append(ipnets, ipnet)
		}
	}
	return ipnets, nil
}

// AllowIPNets is a middleware to only allow access to requests that exist in one of the
// given IPNets. If no IPNets are provided, all requests are allowed to pass through.
// If the request is denied access, the given response will be returned.
func AllowIPNets(ipnets []*net.IPNet, denial interface{}) endpoint.Middleware {
	return endpoint.Middleware(func(ep endpoint.Endpoint) endpoint.Endpoint {
		if len(ipnets) == 0 {
			return ep
		}
		return endpoint.Endpoint(func(ctx context.Context, r interface{}) (interface{}, error) {
			// TODO: add check for forwarded-for header
			addr := ctx.Value(httptransport.ContextKeyRequestRemoteAddr).(string)
			ip := net.ParseIP(addr)
			if ip == nil {
				ipStr, _, err := net.SplitHostPort(addr)
				if err != nil {
					return denial, nil
				}
				ip = net.ParseIP(ipStr)
			}
			var ok bool
			for _, ipnet := range ipnets {
				if ipnet.Contains(ip) {
					ok = true
					break
				}
			}
			if !ok {
				return denial, nil
			}
			// all clear, pass on through
			return ep(ctx, r)
		})
	})
}
