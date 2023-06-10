package httpx

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
)

type dummyKey string

const (
	dummyKeyValue dummyKey = "example"
)

func ExampleContextHandler() {

	// dummy handler to print the value of dummy key
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.Context().Value(dummyKeyValue))
	})

	// wrap it
	wrapped := ContextHandler{
		Handler: handler,
		Replacer: func(r *http.Request) context.Context {
			if r.URL.Path == "/nil/" {
				return nil
			}
			return context.WithValue(r.Context(), dummyKeyValue, "replaced")
		},
	}

	// create a test server
	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	// read the root url (which replaces the context)
	{
		// read the root url
		res, err := http.Get(ts.URL)
		if err != nil {
			panic(err)
		}

		// get the result
		result, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(result))
	}

	// read /nil/ (which does not replace the context)
	{
		// read the root url
		res, err := http.Get(ts.URL + "/nil/")
		if err != nil {
			panic(err)
		}

		// get the result
		result, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(result))
	}

	// Output: replaced
	// <nil>
}

func ExampleContextHandler_nil() {

	// dummy handler to print the value of dummy key
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.Context().Value(dummyKeyValue))
	})

	// wrap it, but don't replace the context
	wrapped := ContextHandler{
		Handler:  handler,
		Replacer: nil,
	}

	// create a test server
	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	// read the root url (context not replaced)
	{
		// read the root url
		res, err := http.Get(ts.URL)
		if err != nil {
			panic(err)
		}

		// get the result
		result, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(result))
	}

	// Output: <nil>
}
