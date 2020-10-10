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
	fmt.Println(string(requestDump))
}
