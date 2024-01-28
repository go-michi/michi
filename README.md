<div align="center">

<img src="https://raw.githubusercontent.com/go-michi/michi/assets/michi-black.png" width=128 />

<h1>michi</h1>

<p>michi is a true 100% compatible with net/http router for Go web applications.</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/go-michi/michi.svg)](https://pkg.go.dev/github.com/go-michi/michi) [![Go Report Card](https://goreportcard.com/badge/github.com/go-michi/michi)](https://goreportcard.com/report/github.com/go-michi/michi) [![MIT](https://img.shields.io/github/license/go-michi/michi)](https://img.shields.io/github/license/go-michi/michi) ![Code size](https://img.shields.io/github/languages/code-size/go-michi/michi)

</div>

## Features

- **True 100% compatible with net/http** - **http.ServerMux**, http.Handler and http.HandlerFunc
- **Enhanced http.ServeMux** - [After Go 1.22](https://go.dev/blog/routing-enhancements), it is possible to use http method and path values
- **API like chi** - Route, Group, With and  middlewares
- **No external dependencies** - Only use standard package
- **Lightweight** - Only 160 lines
- **Performance** - Fast michi == http.ServeMux

## Why michi?

After Go 1.22, HTTP routing in the standard library is now more expressive. The patterns used by net/http.ServeMux have been enhanced to accept methods and wildcards.
But these were already in 3rd party Routing libraries. So, we removed these overlapping features and provide a lack of http.ServeMux features.

<img src="https://raw.githubusercontent.com/go-michi/michi/assets/michi-architecture.png" />
<img src="https://raw.githubusercontent.com/go-michi/michi/assets/chi-architecture.png" />

## About michi(道)
- michi(道) means routes in Japanese.

## Getting Started

```
go get -u github.com/go-michi/michi
```

```go
package main

import (
  "fmt"
  "net/http"

  "github.com/go-chi/chi/v5/middleware"
  "github.com/go-michi/michi"
)

func main() {
  r := michi.NewRouter()
  r.Use(middleware.Logger)
  r.HandleFunc("POST /a/{id}/{$}", func(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("Hello " + req.PathValue("id")))
  })
  http.ListenAndServe(":3000", r)
}
```

Before using michi, read the [http.ServeMux](https://pkg.go.dev/net/http#ServeMux) GoDoc.
For more detailed usage, check the Example in [michi](https://pkg.go.dev/github.com/go-michi/michi) GoDoc.

## Migrating to michi(http.ServeMux) from chi

There are several changes, but rather than changing from chi to michi, it's about changing from chi to http.ServeMux.
Therefore, what you need to understand is how to use the standard library's http.ServeMux, and the knowledge specific to michi is kept to a minimum.

### import michi package

This change is due to michi.

```diff
- import  "github.com/go-chi/chi"
+ import  "github.com/go-michi/michi"

func main() {
-   r := chi.NewRouter()
+   r := michi.NewRouter()
}
```

### Use Handle or HandleFunc method instead of Get, Post, Put, Delete, Patch, Options, Head method

This change is due to http.ServeMux.

```diff
func main() {
-   r.Get("/user/{id}", userHandler)
+   r.HandleFunc("GET /user/{id}", userHandler)
}
```

### Use http.Request.PathValue method

This change is due to http.ServeMux.

```diff
func Handler(w http.ResponseWriter, r *http.Request) {
-   id := chi.URLParam(r, "id")
+   id := r.PathValue("id")
}
```

### Use {$} suffix for exact match

This change is due to http.ServeMux.

- with `{$}`, routing pattern match rule is same as chi
  - `/a/{$}` matches request `/a/`
- without `{$}`, routing pattern match rule is same as old http.ServeMux
  - `/a/` matches request `/a/` and `/a/b`

```diff
func main() {
-   r.Handle("/a/", userHandler)
+   r.Handle("/a/{$}", userHandler)
}
```


### Sub Router

This change is due to http.ServeMux.
http.ServeMux doesn't have Mount method, use Handle method instead of Mount method.
Handle method can't omit parent path.

```go
// chi
func main() {
   r := chi.NewRouter()
   // omit /a/ path
   r.Handle("/hello", handler("hello"))
   r2 := chi.NewRouter()
   r2.Mount("/a", r)
 }
```

```go
// michi
func main() {
    r := michi.NewRouter()
    // can't omit /a/ path
    r.Handle("/a/hello", handler("hello"))
    r2 := michi.NewRouter()
    r2.Handle("/a/", r)
}
```

```diff
func main() {
-   r.Handle("/hello", handler("hello"))
+   r.Handle("/a/hello", handler("hello"))
-   r2.Mount("/a", r)
+   r2.Handle("/a/", r)
 }
```

or using Route

```go
// michi
func main() {
    r := michi.NewRouter()
    // can't omit /a/ path
    r.Route("/a", func(r michi.Router) {
        r.Handle("/hello", handler("hello"))
    })
}
```

## Support version
michi only supports Go 1.22 or later and the two latest versions.
Currently, supports Go 1.22.

## Reference
- https://go.dev/blog/routing-enhancements
- https://pkg.go.dev/net/http#ServeMux

## Credits
- [Peter Kieltyka](https://github.com/pkieltyka) for https://github.com/go-chi/chi
  - michi's middleware interface from chi.
