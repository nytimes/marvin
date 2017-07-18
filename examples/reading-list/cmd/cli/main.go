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
	delete  = flag.Bool("delete", false, "delete this URL from the list (requires -mode update)")

	creds = flag.String("creds", "", "the path of the service account credentials file. if empty, uses Google Application Default Credentials.")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	hc, err := httpClient(ctx)
	if err != nil {
		exitf("%v", err)
	}

	c := readinglist.NewClient(*host, log.NewJSONLogger(os.Stdout),
		httptransport.SetClient(hc))

	switch *mode {
	case "list":
		l, err := c.GetLinks(ctx, *limit)
		if err != nil {
			exitf("unable to get links: %v", err)
		}
		fmt.Printf("successful request with %d links returned\n", len(l.Links))
		for _, lk := range l.Links {
			fmt.Println("* " + lk.Url)
		}
	case "update":
		aurl := *article
		if len(aurl) == 0 {
			exitf("missing -url flag")
		}
		fmt.Println("saving URL:", aurl)
		m, err := c.PutLink(ctx, aurl, *delete)
		if err != nil {
			exitf("unable to update link: %v", err)
		}
		fmt.Println(m.Message)
	default:
		fmt.Println("INVALID MODE. Please choose 'update' or 'list'")
	}
}

func httpClient(ctx context.Context) (*http.Client, error) {
	scopes := []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}

	if *creds == "" {
		return google.DefaultClient(ctx, scopes...)
	}
	jsonKey, err := ioutil.ReadFile(*creds)
	if err != nil {
		return nil, fmt.Errorf("unable to get credentials: %v", err)
	}

	conf, err := google.JWTConfigFromJSON(jsonKey, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}
	return oauth2.NewClient(ctx, conf.TokenSource(ctx)), nil
}

func exitf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	fmt.Fprintln(os.Stderr)
	os.Exit(2)
}
