package fasthttptreemux

import (
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestEmptyGroupAndMapping(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			//everything is good, it paniced
		} else {
			t.Error(`Expected NewGroup("")`)
		}
	}()
	New().GET("", func(ctx *fasthttp.RequestCtx) {})
}
func TestSubGroupSlashMapping(t *testing.T) {
	r := New()
	r.NewGroup("/foo").GET("/", func(ctx *fasthttp.RequestCtx) {
		//w.WriteHeader(200)
		ctx.Response.SetStatusCode(200)
	})

	var rh fasthttp.RequestHeader
	var ctx *fasthttp.RequestCtx

	rh = fasthttp.RequestHeader{}
	rh.SetMethod("GET")
	rh.SetRequestURI("/foo")
	ctx = &fasthttp.RequestCtx{Request: fasthttp.Request{Header: rh}}
	r.Handler(ctx)
	if ctx.Response.StatusCode() != 301 { //should get redirected
		t.Error(`/foo on NewGroup("/foo").GET("/") should result in 301 response, got:`, ctx.Response.StatusCode())
	}

	rh = fasthttp.RequestHeader{}
	rh.SetMethod("GET")
	rh.SetRequestURI("/foo/")
	ctx = &fasthttp.RequestCtx{Request: fasthttp.Request{Header: rh}}
	r.Handler(ctx)
	if ctx.Response.StatusCode() != 200 {
		t.Error(`/foo/ on NewGroup("/foo").GET("/"") should result in 200 response, got:`, ctx.Response.StatusCode())
	}
}

func TestSubGroupEmptyMapping(t *testing.T) {
	r := New()
	r.NewGroup("/foo").GET("", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.SetStatusCode(200)
	})
	rh := fasthttp.RequestHeader{}
	rh.SetMethod("GET")
	rh.SetRequestURI("/foo")
	ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{Header: rh}}
	r.Handler(ctx)
	if ctx.Response.StatusCode() != 200 {
		t.Error(`/foo on NewGroup("/foo").GET("") should result in 200 response, got:`, ctx.Response.StatusCode())
	}
}

func TestGroupMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testGroupMethods(t, scenario.RequestCreator, false)
		testGroupMethods(t, scenario.RequestCreator, true)
	}
}

func TestInvalidHandle(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad handle path should have caused a panic")
		}
	}()
	New().NewGroup("/foo").GET("bar", nil)
}

func TestInvalidSubPath(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad sub-path should have caused a panic")
		}
	}()
	New().NewGroup("/foo").NewGroup("bar")
}

func TestInvalidPath(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad path should have caused a panic")
		}
	}()
	New().NewGroup("foo")
}

//Liberally borrowed from router_test
func testGroupMethods(t *testing.T, reqGen RequestCreator, headCanUseGet bool) {
	var result string
	makeHandler := func(method string) HandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			result = method
		}
	}
	router := New()
	router.HeadCanUseGet = headCanUseGet
	// Testing with a sub-group of a group as that will test everything at once
	g := router.NewGroup("/base").NewGroup("/user")
	g.GET("/:param", makeHandler("GET"))
	g.POST("/:param", makeHandler("POST"))
	g.PATCH("/:param", makeHandler("PATCH"))
	g.PUT("/:param", makeHandler("PUT"))
	g.DELETE("/:param", makeHandler("DELETE"))

	testMethod := func(method, expect string) {
		result = ""
		rh := fasthttp.RequestHeader{}
		rh.SetMethod(method)
		rh.SetRequestURI("/base/user/" + method)
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{Header: rh}}

		router.Handler(ctx)
		if expect == "" && ctx.Response.StatusCode() != http.StatusMethodNotAllowed {
			t.Errorf("Method %s not expected to match but saw code %d", method, ctx.Response.StatusCode())
		}

		if result != expect {
			t.Errorf("Method %s got result %s", method, result)
		}
	}

	testMethod("GET", "GET")
	testMethod("POST", "POST")
	testMethod("PATCH", "PATCH")
	testMethod("PUT", "PUT")
	testMethod("DELETE", "DELETE")
	if headCanUseGet {
		t.Log("Test implicit HEAD with HeadCanUseGet = true")
		testMethod("HEAD", "GET")
	} else {
		t.Log("Test implicit HEAD with HeadCanUseGet = false")
		testMethod("HEAD", "")
	}

	router.HEAD("/base/user/:param", makeHandler("HEAD"))
	testMethod("HEAD", "HEAD")
}

// Ensure that setting a GET handler doesn't overwrite an explciit HEAD handler.
func TestSetGetAfterHead(t *testing.T) {
	var result string
	makeHandler := func(method string) HandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			result = method
		}
	}

	router := New()
	router.HeadCanUseGet = true
	router.HEAD("/abc", makeHandler("HEAD"))
	router.GET("/abc", makeHandler("GET"))

	testMethod := func(method, expect string) {
		result = ""
		rh := fasthttp.RequestHeader{}
		rh.SetMethod(method)
		rh.SetRequestURI("/abc")
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{Header: rh}}

		router.Handler(ctx)

		if result != expect {
			t.Errorf("Method %s got result %s", method, result)
		}
	}

	testMethod("HEAD", "HEAD")
	testMethod("GET", "GET")
}
