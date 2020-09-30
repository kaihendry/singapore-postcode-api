package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/tj/go/http/response"
)

type Building struct {
	Latitude  float64 `json:"LATITUDE,string"`
	Longitude float64 `json:"LONGITUDE,string"`
	Postcode  string  `json:"POSTAL"`
}

type Buildings []Building

var BS Buildings
var views = template.Must(template.ParseGlob("templates/*.html"))

func main() {

	var err error
	BS, err = loadBuildingJSON("buildings.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d buildings", len(BS))

	app := mux.NewRouter()
	app.HandleFunc("/", handleIndex)

	addr := ":" + os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(addr, app))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	PostcodesParam, ok := r.URL.Query()["postcode"]
	if !ok || len(PostcodesParam[0]) != 6 {
		views.ExecuteTemplate(w, "index.html", nil)
		return
	}
	postcode := PostcodesParam[0]
	wantedResponse := r.URL.Query().Get("r")
	log.Println("Postcode", postcode)
	log.Println("Wanted Response", wantedResponse)

	b := BS.lookupPostcode(postcode)

	if wantedResponse != "json" {
		http.Redirect(w, r,
			fmt.Sprintf("https://maps.google.com/?q=%f,%f", b.Latitude, b.Longitude), http.StatusSeeOther)
	} else {
		response.JSON(w, b)
	}
}

func (Buildings Buildings) lookupPostcode(postcode string) (b Building) {
	for _, b = range Buildings {
		if postcode == b.Postcode {
			return b
		}
	}
	return
}

// curl -O https://raw.githubusercontent.com/xkjyeah/singapore-postal-codes/master/buildings.json
func loadBuildingJSON(jsonfile string) (bs Buildings, err error) {
	content, err := ioutil.ReadFile(jsonfile)
	if err != nil {
		return
	}
	err = json.Unmarshal(content, &bs)
	if err != nil {
		return
	}
	return
}
