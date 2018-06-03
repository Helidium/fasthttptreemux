package fasthttptreemux

import (
	"reflect"
	"testing"

	"github.com/valyala/fasthttp"
)

// When we find a node with a matching path but no handler for a method,
// we should fall through and continue searching the tree for a less specific
// match, i.e. a wildcard or catchall, that does have a handler for that method.
func TestMethodNotAllowedFallthrough(t *testing.T) {
	var matchedMethod string
	var matchedPath string
	var matchedParams map[string]string

	router := New()

	addRoute := func(method, path string) {
		router.Handle(method, path, func(ctx *fasthttp.RequestCtx) {
			matchedMethod = method
			matchedPath = path
			matchedParams = make(map[string]string)
			ctx.VisitUserValues(func(key []byte, value interface{}) {
				matchedParams[string(key)] = value.(string)
			})
		})
	}

	checkRoute := func(method, path, expectedMethod, expectedPath string,
		expectedCode int, expectedParams map[string]string) {

		matchedMethod = ""
		matchedPath = ""
		matchedParams = nil

		rh := fasthttp.RequestHeader{}
		rh.SetMethod(method)
		rh.SetRequestURI(path)
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{Header: rh}}
		router.Handler(ctx)
		if expectedCode != ctx.Response.StatusCode() {
			t.Errorf("%s %s expected code %d, saw %d", method, path, expectedCode, ctx.Response.StatusCode())
		}

		if ctx.Response.StatusCode() == 200 {
			if matchedMethod != method || matchedPath != expectedPath {
				t.Errorf("%s %s expected %s %s, saw %s %s", method, path,
					expectedMethod, expectedPath, matchedMethod, matchedPath)
			}

			if !reflect.DeepEqual(matchedParams, expectedParams) {
				t.Errorf("%s %s expected params %+v, saw %+v", method, path, expectedParams, matchedParams)
			}
		}
	}

	addRoute("GET", "/apple/banana/cat")
	addRoute("GET", "/apple/potato")
	addRoute("POST", "/apple/banana/:abc")
	addRoute("POST", "/apple/ban/def")
	addRoute("DELETE", "/apple/:seed")
	addRoute("DELETE", "/apple/*path")
	addRoute("OPTIONS", "/apple/*path")

	checkRoute("GET", "/apple/banana/cat", "GET", "/apple/banana/cat", 200, map[string]string{})
	checkRoute("POST", "/apple/banana/cat", "POST", "/apple/banana/:abc", 200,
		map[string]string{"abc": "cat"})
	checkRoute("POST", "/apple/banana/dog", "POST", "/apple/banana/:abc", 200,
		map[string]string{"abc": "dog"})

	// Wildcards should be checked before catchalls
	checkRoute("DELETE", "/apple/banana", "DELETE", "/apple/:seed", 200,
		map[string]string{"seed": "banana"})
	checkRoute("DELETE", "/apple/banana/cat", "DELETE", "/apple/*path", 200,
		map[string]string{"path": "banana/cat"})

	checkRoute("POST", "/apple/ban/def", "POST", "/apple/ban/def", 200, map[string]string{})
	checkRoute("OPTIONS", "/apple/ban/def", "OPTIONS", "/apple/*path", 200,
		map[string]string{"path": "ban/def"})
	checkRoute("GET", "/apple/ban/def", "", "", 405, map[string]string{})

	// Always fallback to the matching handler no matter how many other
	// nodes without proper handlers are found on the way.
	checkRoute("OPTIONS", "/apple/banana/cat", "OPTIONS", "/apple/*path", 200,
		map[string]string{"path": "banana/cat"})
	checkRoute("OPTIONS", "/apple/bbbb", "OPTIONS", "/apple/*path", 200,
		map[string]string{"path": "bbbb"})

	// Nothing matches on patch
	checkRoute("PATCH", "/apple/banana/cat", "", "", 405, map[string]string{})
	checkRoute("PATCH", "/apple/potato", "", "", 405, map[string]string{})

	// And some 404 tests for good measure
	checkRoute("GET", "/abc", "", "", 404, map[string]string{})
	checkRoute("OPTIONS", "/apple", "", "", 404, map[string]string{})
}
