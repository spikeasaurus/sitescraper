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

// DEBUG ...
// Debug flag
const DEBUG = true

// Sitescraper ...
func Sitescraper(w http.ResponseWriter, r *http.Request) {

	j := job{}

	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		switch err {
		case io.EOF:
			// j.Debug(w, "EOF")
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
	alreadyChecked := *new(map[string]bool)
	alreadyChecked = make(map[string]bool)
	finalCleanup := *new(map[string]bool)
	finalCleanup = make(map[string]bool)

	// Extensions in map format:
	exts := *new(map[string]bool)
	exts = make(map[string]bool)
	// *exts = j.GetExtensions(&w) TO-DO: Add unmarshaller
	exts["jpg"] = true
	exts["jpeg"] = true
	exts["png"] = true

	j.Debug(&w, 2, "Testing Extensions")
	j.Debug(&w, 3, "jpg: ", exts["jpg"])
	j.Debug(&w, 3, "jpeg: ", exts["jpeg"])
	j.Debug(&w, 3, "pdf: ", exts["pdf"])
	j.Debug(&w, 3, "txt: ", exts["txt"])
	j.Debug(&w, 3, "png: ", exts["png"], "\n")

	// Main recursion entry point
	j.GetURIsFromPage(j.URI, &w, j.RecursionDepthInt(), j.RecursionDepthInt(), &l, &j.ValidDomainsRegex, alreadyChecked, exts)

	j.Debug(&w, 1, "\nDEBUG\t---Generating list of downloadable files: ", len(l), " collected")

	// out is the variable for keeping track of which of the URIs from l we actually want to keep
	out := []string{}

	for n, listItem := range l {
		j.Debug(&w, 2, "Examining item ", fmt.Sprint(n), ": ", listItem)

		if j.MatchesExtension(&w, listItem, exts) {
			j.Debug(&w, 3, "Adding ", listItem, " to ", out)
			out = append(out, listItem)
		}
	}

	// Clean up duplicates
	finalList := []string{}
	j.RemoveDuplicates(&w, &out, &finalList, finalCleanup)
	j.Debug(&w, 1, "URIs found:", len(finalList))

	// Final output
	///j.Debug(&w, 1, strings.Trim(fmt.Sprint(finalList), "[]"))
	j.Debug(&w, 1, finalList)

}

// Debug ...
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
	j.Debug(w, 1, "str = ", str)
	j.Debug(w, 1, "ext = ", ext)

	// The shortest possible file name is something like A.jpg; anything shorter, and idk.
	if len(str) < 6 {
		return false
	} else if ext[str[len(str)-3:]] == true || ext[str[len(str)-4:]] == true {
		return true
	} else {
		return false
	}
}

// GetShortenedURI ...
// Truncate a URI safety from whatever length to another shorter length
func (j job) GetShortenedURI(str string, truncateLength int) string {
	return ShortenText(str, truncateLength)
	//	// j.Debug((*w), str[:Min(truncateLength, len(str[truncateLength]))], "\n")
}

// GetExtensions ...
func (j job) GetExtensions(w *http.ResponseWriter) (extensions map[string]bool) {
	j.Debug(w, 2, "Importing extensions from user input")

	for _, ext := range j.Extensions {
		j.Debug(w, 2, "ext = ", ext)
	}
	return extensions
}

// RemoveDuplicates ....
func (j job) RemoveDuplicates(w *http.ResponseWriter, inStr *[]string, outStr *[]string, hash map[string]bool) {

	j.Debug(w, 1, "inStr: ", len(*inStr))

	for _, s := range *inStr {
		if hash[s] != true {
			j.Debug(w, 3, "Unique: ", s)
			*outStr = append((*outStr), s)
			hash[s] = true
		} else {
			j.Debug(w, 3, "Duplicate: ", s)
		}
	}

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
func (j job) GetURIsFromPage(URI string, w *http.ResponseWriter, remainingDepth int, maxDepth int, URIList *[]string, validDomainsRegex *string, alreadyChecked map[string]bool, extensions map[string]bool) {

	j.Debug(w, 1, "Current URI: ", URI)
	j.Debug(w, 1, "Remaining Depth: ", remainingDepth)
	j.Debug(w, 1, "Maximum Depth: ", maxDepth)

	//	if remainingDepth >= 0 {

	// For element-n, issue GET to URI
	//html, _ := func() ([]byte, error) {

	defer RecoverGetURIsFromPage()

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport, Timeout: 0 * time.Second}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	resp, _ := client.Get(URI)
	html, _ := ioutil.ReadAll(resp.Body)

	// Close
	defer resp.Body.Close()

	if err := recover(); err != nil {
		//	j.Debug((*w), "\nERROR\t------", err)
	}

	//	return html, err
	//}()

	// Use REGEX to search HTML BODY for URIs, and append them to URIList
	htmlStr := bytesToString(html)
	urlRegexSyntax := `(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`
	regex := regexp.MustCompile(urlRegexSyntax)
	foundThisInvocation := regex.FindAllString(htmlStr, -1)
	regex2 := regexp.MustCompile(`[^\s\"]*(` + (*validDomainsRegex) + `)[^\s\"]*`)
	j.Debug(w, 2, "Applying regex, uri: ", foundThisInvocation, ")")
	foundThisInvocation = regex2.FindAllString(strings.Join(foundThisInvocation, " "), -1)
	j.Debug(w, 2, "Applying regex, domain name: ", foundThisInvocation, ")")

	// For each of the Urls we read, do the same thing (recurse), and dive deeper
	j.Debug(w, 1, "htmlStr = ", htmlStr, ")")
	j.Debug(w, 1, "Iterating thru URIs found this innovaction (", len(foundThisInvocation), ")")

	for n, foundURI := range foundThisInvocation {

		j.Debug(w, 2, "n=", n, ", remaining depth= ", remainingDepth, ", foundURI=", foundURI)

		// Did we process this already?
		if alreadyChecked[foundURI] != true {
			j.Debug(w, 3, "foundURI is unique: ", ShortenText(foundURI, 125))

			*URIList = append(*URIList, foundURI)
			j.Debug(w, 3, "URIList: ", *URIList)

			// Recurse deeper
			if remainingDepth > 0 {
				j.GetURIsFromPage(foundURI, w, remainingDepth-1, maxDepth, URIList, validDomainsRegex, alreadyChecked, extensions)
			}
			// Switch hash table to indicate this URI has already been checked
			alreadyChecked[foundURI] = true
		} else {
			j.Debug(w, 3, "foundURI is not unique: ", ShortenText(foundURI, 125))
		}
	}
}
