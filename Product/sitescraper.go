package sitescraper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

// Sitescraper ...
func Sitescraper(w http.ResponseWriter, r *http.Request) {

	j := job{}

	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
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

	// test output
	fmt.Fprint(w, "Url: "+j.Uri+"\nExtension: "+j.Extension+"\nRecursion Depth: "+j.RecursionDepth+"\nMinimum File Size: "+j.MinimumFileSize)

	// main loop
	j.Scrape(&w)
}

type job struct {
	Uri             string `json:"uri"`
	Extension       string `json:"ext"`
	RecursionDepth  string `json:"recursiondepth"`
	MinimumFileSize string `json:"minfilesize"`
}

// Scrape ...
func (j job) Scrape(w *http.ResponseWriter, remainingDepth *int) {
	if *remainingDepth > 0 {
		*remainingDepth--

		// Issues GET to uri.
		resp, err := http.Get(j.Uri)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		// Read the html contents
		html, err := ioutil.ReadAll(resp.Body)

		// Define what Url might look like
		const urlRegexSyntax = `https?://[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`
		acceptableImageTypes := "jpe?g|png|gif|bmp"
		acceptableURLTypes := "html?"
		fileExtensionSyntax := "(" + acceptableImageTypes + "|" + acceptableURLTypes + ")"
		regex := regexp.MustCompile(urlRegexSyntax + fileExtensionSyntax)

		// Use htmlSource from which to search
		htmlStr := bytesToString(html)
		urls := regex.FindAllString(htmlStr, -1)

		// Visit every uri
		// u -- every uri
		// i -- numbers every uri
		for i, u := range urls {
			// Print current value of remainingDepth:
			fmt.Fprint(*w, "Remaining Depth: " + strconv.Itoa(remainingDepth))
			//fmt.Fprint(*w, "u,z="+strconv.Itoa(i)+" , "+z)
			j.Uri = u
			j.Scrape(w, remainingDepth)
		}

		// Visit every url found and pull pix, "1 level deep"
		/*for i, z := range urls {
			//	fmt.Println(i, z)

			if z[len(z)-3:] == "jpg" || z[len(z)-4:] == "jpeg" {
				fmt.Println(i, " Saving "+z)
				//	downloadFile(path+"/"+i+"jpg", z)
			} else {
				fmt.Println(i, "Visiting "+z)
				//	savePictures(z)
			}
			//downloadFile(path, z)*/

			// For each link, attempt to open
		}
	}
}

func bytesToString(data []byte) string {
	return string(data[:])
}
