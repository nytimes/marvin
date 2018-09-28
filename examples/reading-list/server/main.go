package main

import (
	"google.golang.org/appengine"

	"github.com/nytimes/marvin"

	readinglist "github.com/nytimes/marvin/examples/reading-list"
)

// a tiny main package that simply initializes the service
// and registers with marvin/GAE.
func main() {
	marvin.Init(readinglist.NewService(readinglist.NewDB()))
	appengine.Main()
}
