package internal

import "net/http"

// Serve starts a minimal HTTP server on the given address.
// The root ("/") route responds with plain text.
func Serve(addr, text string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(text))
	})
	return http.ListenAndServe(addr, nil)
}
