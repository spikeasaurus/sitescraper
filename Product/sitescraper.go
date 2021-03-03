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
		URL             string `json:"url"`
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

	fmt.Fprint(w, "Url: "+d.URL+"\nRecursion Depth: "+d.RecursionDepth+"\nMinimum File Size: "+d.MinimumFileSize)

}
