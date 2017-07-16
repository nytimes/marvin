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
	host = flag.String("host", "http://localhost:8080", "the host of the reading list server")

	mode = flag.String("mode", "list", "(list|update)")

	// list
	limit = flag.Int("limit", 20, "limit for the number of links to return when listing links")

	// update
	article = flag.String("url", "", "the URL to add or delete")
	delete  = flag.Bool("delete", false, "delete this URL from the list")

	creds = flag.String("creds", "gcp.json", "the path of the service account credentials file")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	c := readinglist.NewClient(*host, log.NewJSONLogger(os.Stdout),
		httptransport.SetClient(httpClient(ctx)))

	switch *mode {
	case "list":
		l, err := c.GetLinks(ctx, *limit)
		if err != nil {
			panic("unable to get links: " + err.Error())
		}
		fmt.Printf("successful request with %d links returned\n", len(l.Links))
		for _, lk := range l.Links {
			fmt.Println("* " + lk.Url)
		}
	case "update":
		aurl := *article
		if len(aurl) == 0 {
			panic("please provide a valid URL")
		}
		fmt.Println("saving URL:", aurl)
		m, err := c.PutLink(ctx, aurl, *delete)
		if err != nil {
			panic("unable to update link: " + err.Error())
		}
		fmt.Println(m.Message)
	default:
		fmt.Println("INVALID MODE. Please choose 'update' or 'list'")
	}
}

func httpClient(ctx context.Context) *http.Client {
	jsonKey, err := ioutil.ReadFile(*creds)
	if err != nil {
		panic("unable to get credentials: " + err.Error())
	}

	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	)
	if err != nil {
		panic("unable to parse credentials: " + err.Error())
	}
	return oauth2.NewClient(ctx, conf.TokenSource(ctx))
}
