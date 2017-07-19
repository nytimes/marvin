package main

import (
	"google.golang.org/appengine"

	"github.com/NYTimes/marvin"

	readinglist "github.com/NYTimes/marvin/examples/reading-list"
)

// a tiny main package that simply initializes the service
// and registers with marvin/GAE.
func main() {
	marvin.Init(readinglist.NewService(readinglist.NewDB()))
	appengine.Main()
}
