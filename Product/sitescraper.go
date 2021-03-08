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

// Debug flag
const DEBUG = true

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

	// List of URIs collected is a slice of strings
	l := []string{}

	// Map/hastable for tracking whether a URI has already been seen, to avoid visiting the same URI multiple times
	alreadyChecked := new(map[string]bool)
	*alreadyChecked = make(map[string]bool)

	// Main recursion entry point
	GetUrisFromPage(j.Uri, &w, j.RecursionDepthInt(), j.RecursionDepthInt(), &l, &j.ValidDomainsRegex, alreadyChecked)

	if DEBUG == true {
		fmt.Fprint(w, "\nDEBUG\tGenerating list of downloadable files")
	}
	// out is the variable for keeping track of which of the URIs from l we actually want to keep
	out := []string{}
	for _, listItem := range l {
		length := len(listItem)
		// URI extensions have 3 or 4 len
		if listItem[length-3:] == "jpg" || listItem[length-4:] == "jpeg" {
			out = append(out, listItem)
			if DEBUG == true {
				fmt.Fprint(w, "\nDEBUG\t---Adding to list of downloadable files: ", listItem)
			}
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
			client := &http.Client{Transport: customTransport, Timeout: 0 * time.Second}

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
		urlRegexSyntax := `((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)`
		regex := regexp.MustCompile(urlRegexSyntax)
		htmlStr := bytesToString(html)
		foundThisInvocation := regex.FindAllString(htmlStr, -1)
		regex2 := regexp.MustCompile(`[^\s\"]*(` + (*validDomainsRegex) + `)[^\s\"]*`)
		foundThisInvocation = regex2.FindAllString(strings.Join(foundThisInvocation, " "), -1)

		// For each of the Urls we read, do the same thing (recurse), and dive deeper
		if DEBUG == true {
			fmt.Fprint((*w), "\nDEBUG\tIterating thru URIs found this innovaction")
		}
		for n, foundUri := range foundThisInvocation {

			if DEBUG == true {
				fmt.Fprint((*w), "\nDEBUG\t---n=", n, ", foundUri=", foundUri)
			}
			// Did we process this already?
			if (*alreadyChecked)[foundUri] == false {
				if DEBUG == true {
					fmt.Fprint((*w), "\nDEBUG\t------foundUri is unique: ", ShortenText(foundUri, 50))
				}
				uri = foundUri

				// Recurse deeper
				GetUrisFromPage(foundUri, w, remainingDepth-1, maxDepth, uriList, validDomainsRegex, alreadyChecked)

				// Switch hash table to indicate this URI has already been checked
				(*alreadyChecked)[foundUri] = true
			} else {
				if DEBUG == true {
					fmt.Fprint((*w), "\nDEBUG\t------foundUri is not unique: ", ShortenText(foundUri, 50))
				}
			}

		}

		*uriList = append(*uriList, foundThisInvocation...)

	} else {
		// reset
		if DEBUG == true {
			fmt.Fprint((*w), "\nDEBUG\t---Reached end of max depth (", remainingDepth, ")")
		}
		remainingDepth = maxDepth
	}
}
