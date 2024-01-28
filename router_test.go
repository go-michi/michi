package michi_test

import (
	"fmt"
	"github.com/go-michi/michi/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-michi/michi"
)

func TestServeHTTP(t *testing.T) {
	type args struct {
		requestURL string
	}
	type fields struct {
		handler http.Handler
	}
	type want struct {
		result     string
		statusCode int
		redirect   string
	}
	var result string
	h := func(name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result += name + "h"
		})
	}
	m := func(name string) func(next http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				result += name + "1"
				next.ServeHTTP(w, r)
				result += name + "2"
			})
		}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "/ handler",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/", h("/"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/",
			},
			want: want{
				result:     "/h",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "/a handler",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "a handler not found",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "a handler and middleware.StripSlashes",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(middleware.StripSlashes)
					r.Handle("/a", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "root 2 handlers",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a", h("a"))
					r.Handle("/b", h("b"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "root handler and a middleware",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("a"))
					r.Handle("/", h("/"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/",
			},
			want: want{
				result:     "a1/ha2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "handler not found, exec middleware with sub router",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					rt2 := michi.NewRouter()
					rt2.Handle("/a/{$}", h("a"))
					r.Handle("/", m("a")(rt2))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "a1a2",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "root handler with / not found",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("a"))
					r.Handle("/a/{$}", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "a1a2",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "handler and url with trailing slash",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a/", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "handler and url without trailing slash",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "handler with trailing slash, url without trailing slash. redirect /a to /a/",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a/", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "",
				statusCode: 301,
				redirect:   "/a/",
			},
		},
		{
			name: "url not found but prefix match",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a/", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "url not found with {$}",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("/a/{$}", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "handler not found, doesn't exec middleware with Use without sub router",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("/"))
					r.Handle("/a/b/{$}", h("b"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/c",
			},
			want: want{
				result:     "/1/2",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "handler not found, exec middleware with sub router",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					srt := michi.NewRouter()
					srt.Handle("/b/{$}", h("b"))
					r.Handle("/a/", m("a")(srt))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/c",
			},
			want: want{
				result:     "a1a2",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "handler not found, doesn't exec middleware with With without sub router",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.With(m("a")).Handle("/a", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "handler not found, doesn't exec middleware with With without sub router",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.With(m("a")).Handle("/a", h("a"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "",
				statusCode: 404,
				redirect:   "",
			},
		},
		// Route
		{
			name: "Route / and Handle /",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/", func(r *michi.Router) {
						r.Handle("/", h("/"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/",
			},
			want: want{
				result:     "/h",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a and Handle /",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/a/", func(r *michi.Router) {
						r.Handle("/", h("a"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a and Handler /b",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/a", func(r *michi.Router) {
						r.Handle("/b", h("b"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "bh",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a/ and Handle /b",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/a/", func(r *michi.Router) {
						r.Handle("/b", h("b"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "bh",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a/ Route and /b/ Handle",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/a/", func(r *michi.Router) {
						r.Handle("/b/", h("b"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b/",
			},
			want: want{
				result:     "bh",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "handler not found, exec middleware with Route",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("/"))
					r.Route("/a/", func(r *michi.Router) {
						r.Use(m("a"))
						r.With(m("b")).Handle("/b/{$}", h("b"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/c",
			},
			want: want{
				result:     "/1a1a2/2",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "handler not found, exec middleware with Route",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("/"))
					r.Route("/a/", func(r *michi.Router) {
						r.Use(m("a"))
						r.With(m("b")).Handle("/b", h("b"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b",
			},
			want: want{
				result:     "/1a1b1bhb2a2/2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a/ Route, middleware a, /b/ Route and / Handle",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("a"))
					r.Route("/a/", func(r *michi.Router) {
						r.Route("/b/", func(r *michi.Router) {
							r.Handle("/", h("b"))
						})
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "a1a2",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "Route / Route and /handle, middlewares in Route",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/", func(r *michi.Router) {
						r.Use(m("a"), m("b"))
						r.Handle("/", h("/"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "a1b1/hb2a2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a/ Route and middleware, /b/ Route and middleware. request /a/",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/a/", func(r *michi.Router) {
						r.Use(m("a"))
						r.Handle("/", h("a"))
						r.Route("/b/", func(r *michi.Router) {
							r.Use(m("b"))
							r.Handle("/", h("b"))
						})
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "a1aha2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a/ Route, /b/ Route and / Handle",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/a/", func(r *michi.Router) {
						r.Route("/b/", func(r *michi.Router) {
							r.Handle("/", h("b"))
						})
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b/",
			},
			want: want{
				result:     "bh",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route /a/ and middleware, Route /b/ and middleware. request /a/b/",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/a/", func(r *michi.Router) {
						r.Use(m("a"))
						r.Handle("/", h("a"))
						r.Route("/b/", func(r *michi.Router) {
							r.Use(m("b"))
							r.Handle("/", h("b"))
						})
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/b/",
			},
			want: want{
				result:     "a1b1bhb2a2",
				statusCode: 200,
				redirect:   "",
			},
		},
		// With
		{
			name: "Use after With",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/", func(r *michi.Router) {
						r.With(m("a")).Use(m("b")) // a and b are not affected
						r.Handle("/a", h("a"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Use and Handle after With",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/", func(r *michi.Router) {
						r.With(m("a")).Handle("/a", h("a"))
						r.Use(m("b"))
						r.Handle("/b", h("b")) // not affected With a middleware
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/b",
			},
			want: want{
				result:     "b1bhb2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Use a middleware and With b middleware",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/", func(r *michi.Router) {
						r.Use(m("a"))
						r.With(m("b")).Handle("/b", h("b"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/b",
			},
			want: want{
				result:     "a1b1bhb2a2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "With a and b middlewares",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.With(m("a"), m("b")).Handle("/", h("/"))
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/",
			},
			want: want{
				result:     "a1b1/hb2a2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Use a middleware, a not affected b middleware",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/", func(r *michi.Router) {
						r.Use(m("a"))
						r.With(m("b")).Handle("/b", h("b"))
						r.Handle("/a", h("a"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "a1aha2",
				statusCode: 200,
				redirect:   "",
			},
		},
		// Group
		{
			name: "/ middleware affect to b handler",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("/"))
					r.Handle("/a/", h("a"))
					r.Group(func(r *michi.Router) {
						r.Use(m("b"))
						r.Handle("/b/", h("b"))
					})

					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/b/",
			},
			want: want{
				result:     "/1b1bhb2/2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "a not affected by b middleware",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Group(func(r *michi.Router) {
						r.Use(m("a"))
						r.Handle("/a/", h("a"))
					})
					r.Group(func(r *michi.Router) {
						r.Use(m("b"))
						r.Handle("/b/", h("b"))
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/a/",
			},
			want: want{
				result:     "a1aha2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "a and b not affected to c in other Groups",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("/"))
					r.Handle("/a/", h("a"))
					r.Group(func(r *michi.Router) {
						r.Route("/b/", func(r *michi.Router) {
							r.Use(m("b"))
							r.Handle("/b1/", h("b1"))
						})
					})
					r.Route("/c/", func(r *michi.Router) {
						r.Group(func(r *michi.Router) {
							r.Use(m("c"))
							r.Handle("/c1/", h("c1"))
						})
					})
					return r
				}(),
			},
			args: args{
				requestURL: "https://example.com/c/",
			},
			want: want{
				result:     "/1/2",
				statusCode: 404,
				redirect:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result = ""
			w := httptest.NewRecorder()
			r := httptest.NewRequest("", tt.args.requestURL, nil)
			tt.fields.handler.ServeHTTP(w, r)
			if result != tt.want.result {
				t.Errorf("Result got: %v want: %v", result, tt.want.result)
			}
			if got := w.Result().StatusCode; got != tt.want.statusCode {
				t.Errorf("Result got: %v want: %v", got, tt.want.statusCode)
			}
			if got := w.Header().Get("Location"); got != tt.want.redirect {
				t.Errorf("Location got: %v want: %v", got, tt.want.redirect)
			}
		})
	}
}

func TestServeMuxWithHostAndMethod(t *testing.T) {
	type args struct {
		method     string
		requestURL string
	}
	type fields struct {
		handler http.Handler
	}
	type want struct {
		result     string
		statusCode int
		redirect   string
	}
	var result string
	h := func(name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result += name + "h" + r.PathValue("id") + r.PathValue("id2")
		})
	}
	m := func(name string) func(next http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				result += name + "1"
				next.ServeHTTP(w, r)
				result += name + "2"
			})
		}
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		want         want
		wantRedirect string
	}{
		{
			name: "Handler with host",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("example.com/a", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Handler with host, not found",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("example1.com/a", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "Route with host",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("example.com", func(r *michi.Router) {
						r.Handle("/a", h("a"))
					})
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Route with host, not found",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("example1.com", func(r *michi.Router) {
						r.Handle("/a", h("a"))
					})
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "",
				statusCode: 404,
				redirect:   "",
			},
		},
		{
			name: "Handler POST, request POST",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("POST /a", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Handler POST and multi spaces, request POST",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("POST   /a", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "ah",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Handler POST, request GET",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("POST /a", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodGet,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "",
				statusCode: 405,
				redirect:   "",
			},
		},
		{
			name: "Handler POST, request POST and middleware",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Use(m("a"))
					r.Handle("POST /a", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a",
			},
			want: want{
				result:     "a1aha2",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Handler POST, request POST PathValue",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("POST /a/{id}/", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a/12345/",
			},
			want: want{
				result:     "ah12345",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Handler POST, request POST PathValue",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Route("/{id}/", func(r *michi.Router) {
						r.Handle("POST /{id2}", h("a"))
					})
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/1/2",
			},
			want: want{
				result:     "ah12",
				statusCode: 200,
				redirect:   "",
			},
		},
		{
			name: "Handler POST, request POST wildcard PathValue",
			fields: fields{
				handler: func() http.Handler {
					r := michi.NewRouter()
					r.Handle("POST /a/{id...}", h("a"))
					return r
				}(),
			},
			args: args{
				method:     http.MethodPost,
				requestURL: "https://example.com/a/1/2/3",
			},
			want: want{
				result:     "ah1/2/3",
				statusCode: 200,
				redirect:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result = ""
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.args.method, tt.args.requestURL, nil)
			tt.fields.handler.ServeHTTP(w, r)
			if result != tt.want.result {
				t.Errorf("Result got: %v want: %v", result, tt.want.result)
			}
			if got := w.Result().StatusCode; got != tt.want.statusCode {
				t.Errorf("Result got: %v want: %v", got, tt.want.statusCode)
			}
			if got := w.Header().Get("Location"); got != tt.want.redirect {
				t.Errorf("Location got: %v want: %v", got, tt.want.redirect)
			}
		})
	}
}

func Example() {
	h := func(name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(name + " handler")
		})
	}
	mid := func(name string) func(next http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(name + " start")
				next.ServeHTTP(w, r)
				fmt.Println(name + " end")
			})
		}
	}
	r := michi.NewRouter()
	r.Use(mid("a"))
	r.Handle("/a/", h("a"))
	r.Route("/b/", func(r *michi.Router) {
		r.Use(mid("b"))
		r.Handle("/", h("b"))
		r.With(mid("c1")).Handle("/c1/", h("c1"))
		r.Group(func(r *michi.Router) {
			r.Use(mid("c2"))
			r.Handle("/c2/", h("c2"))
		})
	})
	{
		w := httptest.NewRecorder()
		target := "https://example.com/a/"
		req := httptest.NewRequest(http.MethodPost, target, nil)
		fmt.Println(target)
		r.ServeHTTP(w, req)
		fmt.Println()
	}
	{
		w := httptest.NewRecorder()
		target := "https://example.com/b/"
		req := httptest.NewRequest(http.MethodPost, target, nil)
		fmt.Println(target)
		r.ServeHTTP(w, req)
		fmt.Println()
	}
	{
		w := httptest.NewRecorder()
		target := "https://example.com/b/c1/"
		req := httptest.NewRequest(http.MethodPost, target, nil)
		fmt.Println(target)
		r.ServeHTTP(w, req)
		fmt.Println()
	}
	{
		w := httptest.NewRecorder()
		target := "https://example.com/b/c2/"
		req := httptest.NewRequest(http.MethodPost, target, nil)
		fmt.Println(target)
		r.ServeHTTP(w, req)
		fmt.Println()
	}
	// Output:
	// https://example.com/a/
	// a start
	// a handler
	// a end
	//
	// https://example.com/b/
	// a start
	// b start
	// b handler
	// b end
	// a end
	//
	// https://example.com/b/c1/
	// a start
	// b start
	// c1 start
	// c1 handler
	// c1 end
	// b end
	// a end
	//
	// https://example.com/b/c2/
	// a start
	// b start
	// c2 start
	// c2 handler
	// c2 end
	// b end
	// a end
}
