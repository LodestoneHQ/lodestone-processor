package document

import (
	"net/http"
)

type TikaRoundTripper struct {
	r http.RoundTripper
}

// https://cwiki.apache.org/confluence/display/tika/TikaJAXRS#TikaJAXRS-MultipartSupport TIKA must have an Accept header to return JSON responses.
func (mrt TikaRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Path == "/meta" {
		r.Header.Add("Accept", "application/json")
	}
	if r.URL.Path == "/tika" {
		r.Header.Add("Accept", "text/plain")
	}

	return mrt.r.RoundTrip(r)
}
