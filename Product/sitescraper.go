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
	finalCleanup := new(map[string]bool)
	*finalCleanup = make(map[string]bool)

	// Extensions in map format:
	exts := new(map[string]bool)
	*exts = make(map[string]bool)
	// *exts = j.GetExtensions(&w) TO-DO: Add unmarshaller
	(*exts)["jpg"] = true
	(*exts)["jpeg"] = true
	(*exts)["png"] = true

	if DEBUG == true {
		fmt.Fprint(w, "\nDEBUG\t---Testing Extensions")
		fmt.Fprint(w, "\nDEBUG\t   +--- jpg: ", (*exts)["jpg"])
		fmt.Fprint(w, "\nDEBUG\t   +--- jpeg: ", (*exts)["jpeg"])
		fmt.Fprint(w, "\nDEBUG\t   +--- pdf: ", (*exts)["pdf"])
		fmt.Fprint(w, "\nDEBUG\t   +--- txt: ", (*exts)["txt"])
		fmt.Fprint(w, "\nDEBUG\t   +--- png: ", (*exts)["png"], "\n")
	}

	// Main recursion entry point
	GetUrisFromPage(j.Uri, &w, j.RecursionDepthInt(), j.RecursionDepthInt(), &l, &j.ValidDomainsRegex, alreadyChecked, exts)

	if DEBUG == true {
		fmt.Fprint(w, "\nDEBUG\t---Generating list of downloadable files: ", len(l), " collected")
	}
	// out is the variable for keeping track of which of the URIs from l we actually want to keep
	out := []string{}

	for n, listItem := range l {
		if DEBUG == true {
			fmt.Fprint(w, "\nDEBUG\t------Examining item ", n, ": ", listItem)
		}
		// URI extensions have 3 or 4 len
		if MatchesExtension(listItem, exts) {
			if DEBUG == true {
				fmt.Fprint(w, "\nDEBUG\t----------Adding ", listItem, " to ", out)
			}
			out = append(out, listItem)
		}
	}

	// Clean up duplicates
	finalList := []string{}
	RemoveDuplicates(&w, &out, &finalList, finalCleanup)

	if DEBUG == true {
		fmt.Fprint(w, "\nDEBUG\t---URIs found:", len(finalList), "\n\n\n\n\n")
	}

	// Final output
	fmt.Fprint(w, strings.Trim(fmt.Sprint(finalList), "[]"))

}

// MatchesExtension ...
// TO DO: A better way to check for too short file names
func MatchesExtension(str string, ext *map[string]bool) bool {
	// The shortest possible file name is something like A.jpg; anything shorter, and idk.
	if len(str) < 6 {
		return false
	}
	return (*ext)[str]
}

// GetShortenedUri ...
// Truncate a URI safety from whatever length to another shorter length
func (j job) GetShortenedUri(str string, truncateLength int) string {
	return ShortenText(str, truncateLength)
	//	// fmt.Fprint((*w), str[:Min(truncateLength, len(str[truncateLength]))], "\n")
}

// GetExtensions ...
func (j job) GetExtensions(w *http.ResponseWriter) (extensions map[string]bool) {
	if DEBUG == true {
		fmt.Fprint((*w), "\nDEBUG\t---Importing extensions from user input")
	}
	for _, ext := range j.Extensions {
		if DEBUG == true {
			fmt.Fprint((*w), "\nDEBUG\t---ext = ", ext)
			//	fmt.Fprint((*w), "\nDEBUG\t---extensions[ext] = ", extensions[ext])
		}
		//extensions[ext] = true
	}
	return extensions
}

// RemoveDuplicates ....
func RemoveDuplicates(w *http.ResponseWriter, inStr *[]string, outStr *[]string, hash *map[string]bool) {

	if DEBUG == true {
		fmt.Fprint((*w), "\nDEBUG\t---inStr: ", len(*inStr))
	}

	for _, s := range *inStr {
		if (*hash)[s] != true {
			if DEBUG == true {
				fmt.Fprint((*w), "\nDEBUG\t------Unique: ", s)
			}
			*outStr = append((*outStr), s)
			(*hash)[s] = true
		} else {
			if DEBUG == true {
				fmt.Fprint((*w), "\nDEBUG\t------Duplicate: ", s)
			}
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
	Uri               string `json:"uri"`
	Extensions        string `json:"ext"`
	RecursionDepth    string `json:"recursiondepth"`
	MinimumFileSize   string `json:"minfilesize"`
	ValidDomainsRegex string `json:"validdomains"`
}

// RecursionDepthInt ...
// Here is how we convert json:"recursiondepth", a string, to int
func (j job) RecursionDepthInt() (r int) {
	r, _ = strconv.Atoi(j.RecursionDepth)
	return r
	/// To do: error checking
}

// RecoverGetUrisFromPage ...
// This is how we avoid our recursion from dying in disgrace
func RecoverGetUrisFromPage() {
	if r := recover(); r != nil {
		// recovered
	}
}

// GetUrisFromPage ...
// uriList *[]string is a growing list of URIs
func GetUrisFromPage(uri string, w *http.ResponseWriter, remainingDepth int, maxDepth int, uriList *[]string, validDomainsRegex *string, alreadyChecked *map[string]bool, extensions *map[string]bool) {

	if DEBUG == true {
		fmt.Fprint((*w), "\nDEBUG\tRemaining Depth: ", remainingDepth)
		fmt.Fprint((*w), "\nDEBUG\tMaximum Depth: ", maxDepth)
	}

	if remainingDepth > 0 {

		// For element-n, issue GET to uri
		html, _ := func() ([]byte, error) {

			defer RecoverGetUrisFromPage()

			customTransport := http.DefaultTransport.(*http.Transport).Clone()
			customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			client := &http.Client{Transport: customTransport, Timeout: 15 * time.Second}

			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			resp, err := client.Get(uri)
			html, err := ioutil.ReadAll(resp.Body)

			// Close
			defer resp.Body.Close()

			if err := recover(); err != nil {
				fmt.Fprint((*w), "ERROR\t", err)
			}

			return html, err
		}()

		// Use REGEX to search HTML BODY for URIs, and append them to uriList
		urlRegexSyntax := `(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`
		regex := regexp.MustCompile(urlRegexSyntax)
		htmlStr := bytesToString(html)
		foundThisInvocation := regex.FindAllString(htmlStr, -1)
		regex2 := regexp.MustCompile(`[^\s\"]*(` + (*validDomainsRegex) + `)[^\s\"]*`)
		foundThisInvocation = regex2.FindAllString(strings.Join(foundThisInvocation, " "), -1)

		// For each of the Urls we read, do the same thing (recurse), and dive deeper
		if DEBUG == true {
			fmt.Fprint((*w), "\nDEBUG\tIterating thru URIs found this innovaction (", len(foundThisInvocation), ")")
		}
		for n, foundUri := range foundThisInvocation {

			if DEBUG == true {
				fmt.Fprint((*w), "\nDEBUG\t------n=", n, ", foundUri=", foundUri)
			}
			// Did we process this already?
			if (*alreadyChecked)[foundUri] != true {
				if DEBUG == true {
					fmt.Fprint((*w), "\nDEBUG\t---------foundUri is unique: ", ShortenText(foundUri, 125))
				}

				// Recurse deeper
				GetUrisFromPage(foundUri, w, remainingDepth-1, maxDepth, uriList, validDomainsRegex, alreadyChecked, extensions)

				// Switch hash table to indicate this URI has already been checked
				(*alreadyChecked)[foundUri] = true
			} else {
				if DEBUG == true {
					fmt.Fprint((*w), "\nDEBUG\t---------foundUri is not unique: ", ShortenText(foundUri, 125))
				}
			}

		}
		// Take everything we've found this invocation, including the duplicates, and append it to the master list
		// We are okay with the duplicates because we check for duplicates using the hash table (though in the
		// future, we'll add extra features to reduce wait times)
		*uriList = append(*uriList, foundThisInvocation...)

	} else {
		// Reached the end of depth; reset the remainingDepth back to the max amount for the next "root node"
		if DEBUG == true {
			fmt.Fprint((*w), "\nDEBUG\t---Reached end of max depth (", remainingDepth, ")")
		}
	}
}
