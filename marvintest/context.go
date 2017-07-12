package marvintest

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/appengine/aetest"

	"github.com/NYTimes/marvin/internal"
)

// SetupTestContext is helpful for unit testing `ae` servers.
func SetupTestContext(t *testing.T) func() {
	_, done := SetupTestContextWithContext(t)
	return done
}

func SetupTestContextWithContext(t *testing.T) (context.Context, func()) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatalf("unable to init aetest context: %s", err)
	}
	internal.NewContext = func(r *http.Request) context.Context {
		return ctx
	}
	return ctx, done
}

// SetupBenchmarkContext is helpful for benchmarking `ae` servers.
func SetupBenchmarkContext(t *testing.B) func() {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatalf("unable to init aetest context: %s", err)
	}
	internal.NewContext = func(r *http.Request) context.Context {
		return ctx
	}
	return done
}
