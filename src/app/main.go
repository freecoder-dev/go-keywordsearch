package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

var Revision string

type Suggestion struct {
	Data string `xml:"data,attr"`
}

type ComplereSuggestion struct {
	Suggestions []Suggestion `xml:"suggestion"`
}

type Toplevel struct {
	CompleteSuggestions []ComplereSuggestion `xml:"CompleteSuggestion"`
}

type QueryResponse struct {
	Suggestions []string `json:"suggestions"`
}

func main() {
	http.HandleFunc("/search", handleSuggestions)

	fmt.Println("API server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func handleSuggestions(w http.ResponseWriter, r *http.Request) {
	var suggestions []string
	var questions = []string{
		"",
		"what ",
		"when ",
		"where ",
		"how ",
		"is ",
		"vs ",
	}

	// get the query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	for _, question := range questions {
		q := question + query

		apiURL := fmt.Sprintf("https://google.com/complete/search?output=toolbar&gl=us&q=%s", url.QueryEscape(q))
		resp, err := http.Get(apiURL)
		if err != nil {
			log.Println("Error sending http request: ", err)
			http.Error(w, "Failed to fetch suggestions", http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		// check the response status code
		if resp.StatusCode != http.StatusOK {
			log.Println("HTTP request failed with status code: ", resp.StatusCode)
			http.Error(w, "Failed to fetch suggestions", http.StatusInternalServerError)
			return
		}

		// read rge response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response body", resp.StatusCode)
			http.Error(w, "Failed to read the body", http.StatusInternalServerError)
			return
		}

		// parse the xml response
		var toplevel Toplevel
		err = xml.Unmarshal(body, &toplevel)

		if err != nil {
			log.Println("Error parsing the XML response: ", err)
			http.Error(w, "failed to parse response", http.StatusInternalServerError)
			return
		}

		// etxract the suggestions if available
		if len(toplevel.CompleteSuggestions) > 0 {
			for _, completeSuggestion := range toplevel.CompleteSuggestions {
				for _, suggestion := range completeSuggestion.Suggestions {
					suggestions = append(suggestions, suggestion.Data)
				}
			}
		}
	}

	// build and send the response
	BuildResponse(w, suggestions)
}

func BuildResponse(w http.ResponseWriter, suggestions []string) {

	// prepare the JSON response
	response := QueryResponse{
		Suggestions: suggestions,
	}

	// marshal the JSON response
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling JSON response: ", err)
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	// set the response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// write the JSON response
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Println("Error writning response: ", err)
	}

}
