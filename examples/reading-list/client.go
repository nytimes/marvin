package readinglist

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/NYTimes/marvin"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type Client struct {
	key string
	l   log.Logger

	put endpoint.Endpoint
	get endpoint.Endpoint
}

func NewClient(host string, l log.Logger, opts ...httptransport.ClientOption) *Client {
	return &Client{
		put: retryEndpoint(httptransport.NewClient(
			http.MethodPut,
			mustParseURL(host, "/link.proto"),
			encodePut,
			decodePutResp,
			opts...,
		).Endpoint(), l),
		get: retryEndpoint(httptransport.NewClient(
			http.MethodGet,
			mustParseURL(host, "/list.proto"),
			encodeGet,
			decodeGetResp,
			opts...,
		).Endpoint(), l),
	}
}

func (c Client) GetLinks(ctx context.Context, limit int) (*Links, error) {
	out, err := c.get(ctx, &GetListProtoJSONRequest{Limit: int32(limit)})
	if out != nil {
		return out.(*Links), err
	}
	return nil, err
}

func (c Client) PutLink(ctx context.Context, url string, delete bool) (*Message, error) {
	out, err := c.put(ctx, &PutLinkProtoJSONRequest{
		Request: &LinkRequest{
			Link:   &Link{Url: url},
			Delete: delete,
		},
	})
	if out != nil {
		return out.(*Message), err
	}
	return nil, err
}

func encodePut(ctx context.Context, r *http.Request, req interface{}) error {
	pr := req.(*PutLinkProtoJSONRequest)
	return marvin.EncodeProtoRequest(ctx, r, pr.Request)
}

func encodeGet(ctx context.Context, r *http.Request, req interface{}) error {
	gr := req.(*GetListProtoJSONRequest)
	r.URL.RawQuery = "limit=" + strconv.FormatInt(int64(gr.Limit), 10)
	return nil
}

func decodeGetResp(ctx context.Context, r *http.Response) (interface{}, error) {
	switch r.StatusCode {
	case http.StatusOK:
		var res Links
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read response")
		}
		err = proto.Unmarshal(b, &res)
		return &res, err
	default:
		var res Message
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read response")
		}
		err = proto.Unmarshal(b, &res)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse response")
		}
		return nil, errors.New(res.Message)
	}
}

func decodePutResp(ctx context.Context, r *http.Response) (interface{}, error) {
	var res Message
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response")
	}
	err = proto.Unmarshal(b, &res)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse response")
	}

	switch r.StatusCode {
	case http.StatusOK:
		return &res, nil
	default:
		return nil, errors.New(res.Message)
	}
}

func mustParseURL(host, path string) *url.URL {
	r, err := url.Parse(host + path)
	if err != nil {
		panic("invalid url: " + err.Error())
	}
	return r
}

func retryEndpoint(e endpoint.Endpoint, l log.Logger) endpoint.Endpoint {
	bl := sd.NewEndpointer(
		sd.FixedInstancer{"1"},
		sd.Factory(func(_ string) (endpoint.Endpoint, io.Closer, error) {
			return e, nil, nil
		}),
		l,
	)
	return lb.Retry(3, 2*time.Second, lb.NewRoundRobin(bl))
}
