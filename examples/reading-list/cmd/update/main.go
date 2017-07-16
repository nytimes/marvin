package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

var (
	host    = flag.String("host", "http://localhost:8080", "the host of the reading list server")
	article = flag.String("url", "", "the URL to add or delete")
	delete  = flag.Bool("delete", false, "delete this URL from the list")
	creds   = flag.String("creds", "/opt/nyt/etc/gcp.json", "the path of the service account credentials file")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	c := readinglist.NewClient(*host, log.NewJSONLogger(os.Stdout),
		httptransport.SetClient(httpClient(ctx)))

	aurl := *article
	if len(aurl) == 0 {
		panic("please provide a valid URL")
	}

	l, err := c.PutLink(ctx, aurl, *delete)
	if err != nil {
		panic("unable to get links: " + err.Error())
	}

	fmt.Printf("%#v", l)
}

func httpClient(ctx context.Context) *http.Client {
	jsonKey, err := ioutil.ReadFile(*creds)
	if err != nil {
		panic("unable to get credentials: " + err.Error())
	}

	conf, err := google.JWTConfigFromJSON(
		jsonKey,
	)
	if err != nil {
		panic("unable to parse credentials: " + err.Error())
	}

	return oauth2.NewClient(ctx, conf.TokenSource(ctx))
}
