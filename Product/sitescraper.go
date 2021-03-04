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
	fmt.Fprint(w, "Url: "+j.Uri+"\nExtension: "+j.Extension+"\nRecursion Depth: "+j.RecursionDepth+"\nMinimum File Size: "+j.MinimumFileSize+"\n\n\n")

	// main recursion loop
	l := []string{}
	j.GetUrisFromPage(&w, j.RecursionDepthInt(), j.RecursionDepthInt(), &l)

	// Print results
	fmt.Fprint(w, "\n---------------------------------------------------", "\n")
	fmt.Fprint(w, "\n Work finished; results:", "\n\n")

	for i := range l {
		fmt.Fprint(w, "  -  ", j.GetShortenedUri(l[i], 75), "\n")
		//	fmt.Fprint(w, l[i][:Min(75, len(l[i]))], "\n")
	}

}

// GetShortenedUri ...
func (j job) GetShortenedUri(str string, truncateLength int) string {
	return ShortenText(str, truncateLength)
	//	fmt.Fprint((*w), str[:Min(truncateLength, len(str[truncateLength]))], "\n")
}

// ShortenText ...
func ShortenText(str string, truncateLength int) string {
	return str[:Min(truncateLength, len(str))]
}

// bytestoString ...
func bytesToString(data []byte) string {
	return string(data[:])
}

// Min returns the smaller of x or y.
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type job struct {
	Uri             string `json:"uri"`
	Extension       string `json:"ext"`
	RecursionDepth  string `json:"recursiondepth"`
	MinimumFileSize string `json:"minfilesize"`
}

func (j job) RecursionDepthInt() (r int) {
	r, _ = strconv.Atoi(j.RecursionDepth)
	return r
	/// To do: error checking
}

// GetUrisFromPage ...
// uriList *[]string is a growing list of URIs
func (j job) GetUrisFromPage(w *http.ResponseWriter, remainingDepth int, maxDepth int, uriList *[]string) {

	// debug
	fmt.Fprint((*w), "\n---------------------------------------------------", "\n")
	fmt.Fprint((*w), " - ", remainingDepth, "\n")
	fmt.Fprint((*w), " - Searching under: ", j.Uri, "\n")

	if remainingDepth > 0 {

		// For element-n, issue GET to uri
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

		// Use REGEX to search HTML BODY for URIs, and append them to uriList
		htmlStr := bytesToString(html)
		foundThisInvocation := regex.FindAllString(htmlStr, -1)
		*uriList = append(*uriList, foundThisInvocation...)

		// For each of the Urls we read, do the same thing (recurse), and dive deeper

		for a := range foundThisInvocation {
			j.Uri = foundThisInvocation[a]
			fmt.Fprint((*w), " +--- ", a, " ", ShortenText(j.Uri, 75), "\n")
			j.GetUrisFromPage(w, remainingDepth-1, maxDepth, uriList)
		}
	} else {
		// reset
		remainingDepth = maxDepth
		fmt.Fprint((*w), " - Reached the end of ", ShortenText(j.Uri, 75), "\n")
	}
}
