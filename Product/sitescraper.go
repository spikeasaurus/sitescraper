package sitescraper

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math/rand"
	"net/http"
)

// Sitescraper ...
func Sitescraper(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, "Hi. I am your Google Assitant. Say Help for a list of things I can do.")

	var d struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		switch err {
		case io.EOF:
			fmt.Fprint(w, "Hello World!")
			return
		default:
			log.Printf("json.NewDecoder: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	phrases := []string{"I am not Alexa; I am your Google Assistant",
		"Read me a bedtime story Home device not found",
		"Turning off foyer lights",
		"Confirming emergency 911 call",
		"Restarting to install firmware"}

	if d.Message != "" {
		fmt.Fprint(w, phrases[rand.Intn(4)])
		return
	}

	fmt.Fprint(w, html.EscapeString(d.Message))
}
