package sitescraper

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Sitescraper ...
func Sitescraper(w http.ResponseWriter, r *http.Request) {

	j := job{}

	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		switch err {
		case io.EOF:
			// fmt.Fprint(w, "EOF")
			return
		default:
			log.Printf("json.NewDecoder: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	// test output
	// fmt.Fprint(w, "Url: "+j.Uri+"\nExtension: "+j.Extension+"\nRecursion Depth: "+j.RecursionDepth+"\nMinimum File Size: "+j.MinimumFileSize+"\n\n\n")

	// main recursion loop
	l := []string{}

	GetUrisFromPage(j.Uri, &w, j.RecursionDepthInt(), j.RecursionDepthInt(), &l, &j.ValidDomainsRegex)

	// Print results
	// fmt.Fprint(w, "\n---------------------------------------------------", "\n")
	// fmt.Fprint(w, "\n Work finished; results:", "\n\n")

	out := []string{}
	for _, listItem := range l {
		// fmt.Fprint(w, "  -  ", j.GetShortenedUri(listItem, 75), "\n")
		length := len(listItem)
		if listItem[length-3:] == "jpg" || listItem[length-4:] == "jpeg" {
			out = append(out, listItem)
		}
	}
	fmt.Fprint(w, strings.Trim(fmt.Sprint(out), "[]"))

}

// GetShortenedUri ...
func (j job) GetShortenedUri(str string, truncateLength int) string {
	return ShortenText(str, truncateLength)
	//	// fmt.Fprint((*w), str[:Min(truncateLength, len(str[truncateLength]))], "\n")
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
	Uri               string `json:"uri"`
	Extension         string `json:"ext"`
	RecursionDepth    string `json:"recursiondepth"`
	MinimumFileSize   string `json:"minfilesize"`
	ValidDomainsRegex string `json:"validdomains"`
}

func (j job) RecursionDepthInt() (r int) {
	r, _ = strconv.Atoi(j.RecursionDepth)
	return r
	/// To do: error checking
}

//

func RecoverGetUrisFromPage() {
	if r := recover(); r != nil {
		// recovered
	}
}

// GetUrisFromPage ...
// uriList *[]string is a growing list of URIs
func GetUrisFromPage(uri string, w *http.ResponseWriter, remainingDepth int, maxDepth int, uriList *[]string, validDomainsRegex *string) {

	// debug
	// fmt.Fprint((*w), "\n---------------------------------------------------", "\n")
	// fmt.Fprint((*w), " - ", remainingDepth, "\n")
	// fmt.Fprint((*w), " - Searching under: ", uri, "\n")

	if remainingDepth > 0 {

		// For element-n, issue GET to uri
		html, _ := func() ([]byte, error) {

			defer RecoverGetUrisFromPage()

			customTransport := http.DefaultTransport.(*http.Transport).Clone()
			customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			client := &http.Client{Transport: customTransport, Timeout: 3 * time.Second}

			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			resp, err := client.Get(uri)
			html, err := ioutil.ReadAll(resp.Body)

			// Close
			resp.Body.Close()

			if err := recover(); err != nil {
				fmt.Println("Error!! ", err)
			}

			return html, err
		}()

		// Define what Url might look like

		urlRegexSyntax := `(?=((https?:\/\/)?)(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*))`
		regex := regexp.MustCompile(urlRegexSyntax)
		var urlSubSyntax string = `([\s]*` + (*validDomainsRegex) + `[^\s]*)`
		urlRegexSyntaxSubmatched := regex.FindStringSubmatch(urlSubSyntax)

		fmt.Fprint((*w), urlRegexSyntaxSubmatched)
		//				   (?=((https?:\/\/)?)(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*))([^\s]*`  (imagevenue)           [^\s]*)

		// Use REGEX to search HTML BODY for URIs, and append them to uriList
		htmlStr := bytesToString(html)
		foundThisInvocation := regex.FindAllString(htmlStr, -1)
		*uriList = append(*uriList, foundThisInvocation...)

		// fmt.Fprint((*w), " - Total URIs: ", len(*uriList), " (", len(foundThisInvocation), ") found this pass\n")
		// fmt.Fprint((*w), " - Read BODY: ", len(htmlStr), " characters\n")
		// fmt.Fprint((*w), " - Found this invocation: ", len(foundThisInvocation), "\n")

		// For each of the Urls we read, do the same thing (recurse), and dive deeper

		for n, foundUri := range foundThisInvocation {
			uri = foundThisInvocation[n]
			// fmt.Fprint((*w), " +--- ", n, " ", ShortenText(foundUri, 75), "\n")
			GetUrisFromPage(foundUri, w, remainingDepth-1, maxDepth, uriList, validDomainsRegex)
		}
	} else {
		// reset
		remainingDepth = maxDepth
		// fmt.Fprint((*w), " - Reached the end of ", ShortenText(uri, 75), "\n")
	}
}
