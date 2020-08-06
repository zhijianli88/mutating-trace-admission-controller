package print

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

// Request print http request header and body
func Request(r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println("Request Error: ", err)
	}
	fmt.Printf("\x1b[31m%s\x1b[0m\n", string(requestDump))
}
