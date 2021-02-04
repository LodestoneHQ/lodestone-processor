package document

import (
	"net/http"
)

type TikaRoundTripper struct {
	r                   http.RoundTripper
	ocrLanguageOverride string
}

// https://cwiki.apache.org/confluence/display/tika/TikaJAXRS#TikaJAXRS-MultipartSupport TIKA must have an Accept header to return JSON responses.
func (mrt TikaRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Path == "/meta" {
		r.Header.Add("Accept", "application/json")
	}
	if r.URL.Path == "/tika" {
		r.Header.Add("Accept", "text/plain")

		if mrt.ocrLanguageOverride != "" {
			// Note: we are not using Header#Add here because it would mess up the key
			// Go HTTP expects all headers to be case insensitive and converts the case to avoid ambiguity.
			// Tika on the other hand treats headers case sensitive and ignores headers with messed up case
			r.Header["X-Tika-OCRLanguage"] = append(r.Header["X-Tika-OCRLanguage"], mrt.ocrLanguageOverride)
		}
	}

	return mrt.r.RoundTrip(r)
}
