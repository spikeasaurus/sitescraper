package sitescraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Sitescraper ...
func Sitescraper(w http.ResponseWriter, r *http.Request) {

	var d struct {
		Uri             string `json:"uri"`
		RecursionDepth  string `json:"recursiondepth"`
		MinimumFileSize string `json:"minfilesize"`
	}

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		switch err {
		case io.EOF:
			fmt.Fprint(w, "EOF")
			return
		default:
			log.Printf("json.NewDecoder: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	d.Uri = "asdf"
	ParseUri(d.Uri)

	// test output
	fmt.Fprint(w, "Url: "+d.Uri+"\nRecursion Depth: "+d.RecursionDepth+"\nMinimum File Size: "+d.MinimumFileSize)

}

// ParseUri ...
func ParseUri(urlString string) {

}

// ReadUri ...
func ReadUri(urlString string) {

}
