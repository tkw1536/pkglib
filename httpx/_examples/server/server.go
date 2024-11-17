//spellchecker:words main
package main

//spellchecker:words http github pkglib httpx recovery
import (
	"log"
	"net/http"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/recovery"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			httpx.RenderErrorPage(recovery.Recover(recover()), httpx.Response{}, w, r)
		}()
		panic("stuff")
	})

	log.Fatal(http.ListenAndServe("localhost:3000", nil))
}
