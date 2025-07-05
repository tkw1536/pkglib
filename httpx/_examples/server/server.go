//spellchecker:words main
package main

//spellchecker:words http github pkglib httpx recovery
import (
	"log"
	"net/http"

	"go.tkw01536.de/pkglib/httpx"
	"go.tkw01536.de/pkglib/recovery"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			httpx.RenderErrorPage(recovery.Recover(recover()), httpx.Response{}, w, r)
		}()
		panic("stuff")
	})

	log.Fatal(http.ListenAndServe("localhost:3000", nil)) // #nosec G114 -- this is example code and doesn't need timeouts
}

// spellchecker:words nosec
