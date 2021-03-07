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
	var alreadyChecked map[string]bool

	GetUrisFromPage(j.Uri, &w, j.RecursionDepthInt(), j.RecursionDepthInt(), &l, &j.ValidDomainsRegex, &alreadyChecked)

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
func GetUrisFromPage(uri string, w *http.ResponseWriter, remainingDepth int, maxDepth int, uriList *[]string, validDomainsRegex *string, alreadyChecked *map[string]bool) {

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

		// Use REGEX to search HTML BODY for URIs, and append them to uriList
		urlRegexSyntax := `[(http(s)?):\/\/(www\.)?a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`
		regex := regexp.MustCompile(urlRegexSyntax)
		htmlStr := bytesToString(html)
		foundThisInvocation := regex.FindAllString(htmlStr, -1)
		fmt.Fprint((*w), "this invocation 1\n", foundThisInvocation)

		regex2 := regexp.MustCompile(`[^\s]*(` + (*validDomainsRegex) + `)[^\s]*`)
		foundThisInvocation = regex2.FindAllString(strings.Join(foundThisInvocation, " "), -1)

		fmt.Fprint((*w), "this invocation 2\n", foundThisInvocation)
		*uriList = append(*uriList, foundThisInvocation...)

		// For each of the Urls we read, do the same thing (recurse), and dive deeper
		for n, foundUri := range foundThisInvocation {
			// Did we process this already?
			if (*alreadyChecked)[foundUri] == false {
				uri = foundThisInvocation[n]
				GetUrisFromPage(foundUri, w, remainingDepth-1, maxDepth, uriList, validDomainsRegex, alreadyChecked)

				// Switch hash table to indicate this URI has already been checked
				(*alreadyChecked)[foundUri] = true
			}
		}

	} else {
		// reset
		remainingDepth = maxDepth
	}
}
