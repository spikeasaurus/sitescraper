package sitescraper

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	//"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DEBUG ...
// Debug flag
const DEBUG = true

// Sitescraper ...
func Sitescraper(w http.ResponseWriter, r *http.Request) {

	j := job{}

	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		switch err {
		case io.EOF:
			j.Debug(&w, 1, "EOF")
			return
		default:
			log.Printf("json.NewDecoder: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	// Map/hastable for tracking whether a URI has already been seen, to avoid visiting the same URI multiple times
	checkedURIs := *new(map[string]bool)
	checkedURIs = make(map[string]bool)

	// Extensions in map format:
	exts := *new(map[string]bool)
	exts = make(map[string]bool)
	// *exts = j.GetExtensions(&w) TO-DO: Add unmarshaller

	exts["jpg"] = true
	exts["jpeg"] = true
	exts["png"] = true

	j.Debug(&w, 1, "Testing Extensions")
	j.Debug(&w, 1, "jpg: ", exts["jpg"])
	j.Debug(&w, 1, "jpeg: ", exts["jpeg"])
	j.Debug(&w, 1, "pdf: ", exts["pdf"])
	j.Debug(&w, 1, "txt: ", exts["txt"])
	j.Debug(&w, 1, "png: ", exts["png"], "\n")

	// Main recursion entry point
	j.GetURIsFromPage(j.URI, &w, j.RecursionDepthInt(), &j.ValidDomainsRegex, checkedURIs, exts)

	j.Debug(&w, 1, "Generating list of downloadable files: ", len(checkedURIs), " collected")

	// out is the variable for keeping track of which of the URIs from l we actually want to keep
	out := []string{}

	for uri := range checkedURIs {
		j.Debug(&w, 1, "Examining item ", uri)

		if j.MatchesExtension(&w, uri, exts) {
			j.Debug(&w, 3, "Adding ", uri, " to ", out)
			out = append(out, uri)
		}
	}

	j.Debug(&w, 1, "URIs found:", len(out))
	fmt.Fprint(w, out)
}

// Debug ...
// This is how we print debug messages at different levels of detail.
// debuglevel: level of detail
// str ...interface{}: enter strings separated by commas
func (j job) Debug(w *http.ResponseWriter, debugLevel int, str ...interface{}) {
	debugLevelRequested, err := strconv.Atoi(j.DebugLevel)
	if err != nil {
		return
	}
	if debugLevelRequested >= debugLevel {
		fmt.Fprint((*w), "\nDEBUG\t", strings.Repeat("---", debugLevel), strings.Trim(fmt.Sprint(str...), "[]"))
	}
}

// MatchesExtension ...
// TO DO: A better way to check for too short file names
func (j job) MatchesExtension(w *http.ResponseWriter, str string, ext map[string]bool) bool {

	j.Debug(w, 2, "This URI = ", str)
	j.Debug(w, 2, "Valid Extensions = ", ext)

	// The shortest possible file name is something like A.jpg; anything shorter, and idk.
	if len(str) < 6 {
		j.Debug(w, 2, "Invalid extension (file length too short)")
		return false
	} else if ext[str[len(str)-3:]] == true || ext[str[len(str)-4:]] == true {
		j.Debug(w, 2, "Valid extension")
		return true
	} else {
		j.Debug(w, 2, "Invalid extension (other possibilities exhausted)")
		return false
	}
}

// GetShortenedURI ...
// Truncate a URI safety from whatever length to another shorter length
func (j job) GetShortenedURI(str string, truncateLength int) string {
	return ShortenText(str, truncateLength)
}

// GetExtensions ...
func (j job) GetExtensions(w *http.ResponseWriter) (extensions map[string]bool) {
	j.Debug(w, 2, "Importing extensions from user input")

	for _, ext := range j.Extensions {
		j.Debug(w, 2, "ext = ", ext)
	}
	return extensions
}

// ShortenText ...
// Truncate a string safety from whatever length to another shorter length
func ShortenText(str string, truncateLength int) string {
	return str[:Min(truncateLength, len(str))]
}

// bytestoString ...
func bytesToString(data []byte) string {
	return string(data[:])
}

// Min ...
// Min returns the smaller of x or y.
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// These are the things the user can POST to us.
type job struct {
	URI               string `json:"uri"`
	Extensions        string `json:"ext"`
	RecursionDepth    string `json:"recursiondepth"`
	MinimumFileSize   string `json:"minfilesize"`
	ValidDomainsRegex string `json:"validdomains"`
	DebugLevel        string `json:"debug"`
}

// RecursionDepthInt ...
// Here is how we convert json:"recursiondepth", a string, to int
func (j job) RecursionDepthInt() (r int) {
	r, _ = strconv.Atoi(j.RecursionDepth)
	return r
	/// To do: error checking
}

// RecoverGetURIsFromPage ...
// This is how we avoid our recursion from dying in disgrace
func RecoverGetURIsFromPage() {
	if r := recover(); r != nil {
		// recovered
	}
}

// GetURIsFromPage ...
// URIList *[]string is a growing list of URIs
func (j job) GetURIsFromPage(URI string, w *http.ResponseWriter, remainingDepth int, validDomainsRegex *string, checkedURIs map[string]bool, extensions map[string]bool) {

	defer RecoverGetURIsFromPage()

	j.Debug(w, 1, "URI: ", URI)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport, Timeout: 0 * time.Second}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	resp, errClientGet := client.Get(URI)
	if errClientGet != nil {
		j.Debug(w, 2, "HTTP Response Code is ", errClientGet, "when navigating to: ", URI)
	}
	bodyAsHTMLInBytes, errIOUtilRead := ioutil.ReadAll(resp.Body)
	if errIOUtilRead != nil {
		j.Debug(w, 2, "IOUtil error is ", errIOUtilRead, "when navigating to: ", URI)
	}

	// Close
	defer resp.Body.Close()

	if err := recover(); err != nil {
		//	j.Debug((*w), "\nERROR\t------", err)
	}

	// Convert HTML of Body from []bytes to string
	bodyAsHTMLInString := bytesToString(bodyAsHTMLInBytes)

	// Use REGEX to search HTML BODY for URIs, and append them to URIList
	regexURIFilter := `(https?:\/\/|\/)([\w\.]*)([a-z\.]{2,6})([\/\w \.\-\#]*)*\/?`
	URIFilter := regexp.MustCompile(regexURIFilter)
	foundThisInvocation := URIFilter.FindAllString(bodyAsHTMLInString, -1)
	j.Debug(w, 2, "Applying regex, uri: ", foundThisInvocation)

	// Use REGEX a second time to filter by substrings that appear in the URI (this helps prevent searches from wandering off)
	regexNameFilter := regexp.MustCompile(`[^\s\"]*(` + (*validDomainsRegex) + `)[^\s\"]*`)
	foundThisInvocation = regexNameFilter.FindAllString(strings.Join(foundThisInvocation, " "), -1)
	j.Debug(w, 2, "Applying regex, domain name: ", foundThisInvocation)

	// For each of the Urls we read, do the same thing (recurse), and dive deeper
	j.Debug(w, 4, "htmlStr = ", bodyAsHTMLInString, ")")
	j.Debug(w, 2, "Iterating thru URIs found this innovaction (", len(foundThisInvocation), ")")

	for n, foundURI := range foundThisInvocation {

		j.Debug(w, 2, "URI: ", URI, " >  n: ", n, " > remaining depth: ", remainingDepth, " > foundURI: ", ShortenText(foundURI, 125))

		// Is the foundURI a relative URI or an absolute URI? If it's a relative URI, we should append the stem

		var parentURI, relativeURI *url.URL
		relativeURI, relativeURIError := relativeURI.Parse(foundURI)
		j.Debug(w, 3, "relativeURI: ", ShortenText(relativeURI.String(), 125), "; error: ", relativeURIError)
		if relativeURI.Scheme == "" {
			j.Debug(w, 3, "link is relative")
			parentURI, parentURIError := parentURI.Parse(URI)
			j.Debug(w, 3, "parentURI: ", ShortenText(parentURI.String(), 125), "; error: ", parentURIError)
			foundURI = parentURI.ResolveReference(relativeURI).String()
		} else {
			j.Debug(w, 3, "link is absolute")
		}

		// Did we process this already?
		if checkedURIs[foundURI] == true {
			// Case: Duplicate URI
			j.Debug(w, 3, "foundURI is a duplicate: ", ShortenText(foundURI, 125))
		} else {
			// Case: Novel URI
			checkedURIs[foundURI] = true
			j.Debug(w, 3, "foundURI is novel: ", ShortenText(foundURI, 125))

			// Recurse deeper
			if remainingDepth > 0 {
				j.Debug(w, 4, ShortenText(foundURI, 125), " -- recursing deeper")
				j.GetURIsFromPage(foundURI, w, remainingDepth-1, validDomainsRegex, checkedURIs, extensions)

				// Switch hash table to indicate this URI has already been checked
				//alreadyChecked[foundURI] = true
			} else {
				j.Debug(w, 4, ShortenText(foundURI, 125), " -- recursion depth reached limit")
			}
		}
	}
}
