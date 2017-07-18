package readinglist

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"google.golang.org/appengine"

	"github.com/NYTimes/marvin"
	"github.com/NYTimes/marvin/marvintest"
	"github.com/kr/pretty"
)

type testStep struct {
	name                string
	givenURL            string
	givenMethod         string
	givenPayloadFixture string

	wantCode    int
	wantFixture string
}

func TestService(t *testing.T) {

	tests := []struct {
		name string

		steps []testStep
	}{
		{
			name: "success-get_set_get_delete_get",

			steps: []testStep{
				{
					name:        "initial empty get",
					givenURL:    "/list.json",
					givenMethod: http.MethodGet,

					wantCode:    http.StatusOK,
					wantFixture: "list-empty.json",
				},
				{
					name:                "inserting initial link",
					givenURL:            "/link.json",
					givenMethod:         http.MethodPut,
					givenPayloadFixture: "put-1.json",

					wantCode:    http.StatusOK,
					wantFixture: "success-msg.json",
				},
				{
					name:        "verifying initial insert",
					givenURL:    "/list.json",
					givenMethod: http.MethodGet,

					wantCode:    http.StatusOK,
					wantFixture: "get-1.json",
				},
				{
					name:                "deleting initial link",
					givenURL:            "/link.json",
					givenMethod:         http.MethodPut,
					givenPayloadFixture: "delete-1.json",

					wantCode:    http.StatusOK,
					wantFixture: "success-msg.json",
				},
				{
					name:        "verifying delete",
					givenURL:    "/list.json",
					givenMethod: http.MethodGet,

					wantCode:    http.StatusOK,
					wantFixture: "list-empty.json",
				},
			},
		},
		{
			name: "success-set_set-dupe_set_get",

			steps: []testStep{
				{
					name:                "inserting initial link",
					givenURL:            "/link.json",
					givenMethod:         http.MethodPut,
					givenPayloadFixture: "put-1.json",

					wantCode:    http.StatusOK,
					wantFixture: "success-msg.json",
				},
				{
					name:                "duping initial link",
					givenURL:            "/link.json",
					givenMethod:         http.MethodPut,
					givenPayloadFixture: "put-1.json",

					wantCode:    http.StatusOK,
					wantFixture: "success-msg.json",
				},
				{
					name:                "put second link",
					givenURL:            "/link.json",
					givenMethod:         http.MethodPut,
					givenPayloadFixture: "put-2.json",

					wantCode:    http.StatusOK,
					wantFixture: "success-msg.json",
				},
				{
					name:        "verifying list",
					givenURL:    "/list.json",
					givenMethod: http.MethodGet,

					wantCode:    http.StatusOK,
					wantFixture: "get-2.json",
				},
				{
					name:                "delete 1",
					givenURL:            "/link.json",
					givenMethod:         http.MethodPut,
					givenPayloadFixture: "delete-1.json",

					wantCode:    http.StatusOK,
					wantFixture: "success-msg.json",
				},
				{
					name:        "verify delete",
					givenURL:    "/list.json",
					givenMethod: http.MethodGet,

					wantCode:    http.StatusOK,
					wantFixture: "get-2-only.json",
				},
			},
		},
		{
			name: "bad_requests",

			steps: []testStep{
				{
					name:                "bad put",
					givenURL:            "/link.json",
					givenMethod:         http.MethodPut,
					givenPayloadFixture: "bad-put.json",

					wantCode:    http.StatusBadRequest,
					wantFixture: "bad-req-msg.json",
				},
			},
		},
	}
	// deal with AE context/dev_appserver initialization.
	// this call is expensive so we only want to call it once.
	ctx, ctxdone := marvintest.SetupTestContextWithContext(t)
	defer ctxdone()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctx, err := appengine.Namespace(ctx, test.name)
			if err != nil {
				t.Fatalf("unable to init namespace: %s", err)
			}
			// override context to set a namespace for this test step
			// to isolate it from the other test steps' data
			marvintest.SetServerContext(ctx)

			// init the server so we can call ServeHTTP on it
			svr := marvin.NewServer(NewService(NewDB()))

			//go through each step and call the server/verify the response
			for _, step := range test.steps {

				var payload []byte
				// set up the request body if a fixture is given
				if step.givenPayloadFixture != "" {
					var err error
					payload, err = ioutil.ReadFile("fixtures/" + step.givenPayloadFixture)
					if err != nil {
						t.Fatalf("unable to read test fixture: %s", err)
					}
				}
				r, err := http.NewRequest(step.givenMethod, step.givenURL, bytes.NewBuffer(payload))
				if err != nil {
					t.Fatalf("unable to create test request: %s", err)
				}

				w := httptest.NewRecorder()
				// hit the server and capture the response
				svr.ServeHTTP(w, r)

				// check status code for what we want
				if w.Code != step.wantCode {
					t.Errorf("expected response of %d, got %d", step.wantCode, w.Code)
				}

				// compare response against expected fixture
				if step.wantFixture != "" {
					var resp map[string]interface{}
					err = json.NewDecoder(w.Body).Decode(&resp)
					if err != nil {
						t.Fatalf("unable to read response: %s", err)
					}
					compareFixture(t, step.name, resp, "fixtures/"+step.wantFixture)
				}

				// le sigh, local datastore be slow :(
				time.Sleep(1 * time.Second)
			}
		})
	}
}

// compareFixture takes a struct and compares it to a test fixture by
// converting them both to map[string]interface{}'s and doing a
// reflect.DeepEqual. It will fail the test if they are not the same and
// neatly log the differences.
func compareFixture(t *testing.T, name string, obj interface{}, fixture string) {
	// read and convert input object into map[string]interface{}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Errorf("[%s] could not marshal input to json", name)
		return
	}
	var actual map[string]interface{}
	err = json.Unmarshal(b, &actual)
	if err != nil {
		t.Errorf("[%s] could not unmarshal actual to map[string]interface{}", name)
		return
	}

	// read and convert fixture into map[string]interface{}
	b, err = ioutil.ReadFile(fixture)
	if err != nil {
		t.Fatalf("[%s] could not open test fixture: %s", name, fixture)
	}
	var expected map[string]interface{}
	err = json.Unmarshal(b, &expected)
	if err != nil {
		t.Errorf("[%s] could not unmarshal expected to map[string]interface{}: %s", name, err)
		return
	}

	// compare them
	if !reflect.DeepEqual(actual, expected) {
		// marshal back to json strings for useful diff output
		diffs := pretty.Diff(actual, expected)
		t.Errorf("[%s] found %d difference(s) in actual vs. expected:", name, len(diffs))
		for _, diff := range diffs {
			t.Log(diff)
		}
	}
}
