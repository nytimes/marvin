package marvintest

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/appengine/aetest"

	"github.com/NYTimes/marvin/internal"
)

// SetupTestContext will start up dev_appserver.py via `appengine/aetest` and inject the
// context into the marvin server so users can call
// `marvin.NewServer(svc).ServeHTTP(w, r)` within tests.
// This call is very expensive and should be used sparingly in your test suite.
func SetupTestContext(t *testing.T) func() {
	_, done := SetupTestContextWithContext(t)
	return done
}

// SetupTestContextWithContext will start up dev_appserver.py via `appengine/aetest` and
// inject the context into the marvin server so users can call
// `marvin.NewServer(svc).ServeHTTP(w, r)` within tests. This function also returns the
// server context for use outside the server.
// This call is very expensive and should be used sparingly in your test suite.
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

// SetupBenchmarkContext will start up dev_appserver.py via `appengine/aetest` and inject the
// context into the marvin server so users can call `marvin.NewServer(svc).ServeHTTP(w, r)`
// within tests.
// This call is very expensive and should be used sparingly in your test suite.
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

// SetServerContext will override the server context and inject the given context
// into incoming requests. The effect of this function is global.
func SetServerContext(ctx context.Context) {
	internal.NewContext = func(r *http.Request) context.Context {
		return ctx
	}
}
