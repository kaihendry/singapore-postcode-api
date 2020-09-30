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

type building struct {
	Latitude  float64 `json:"LATITUDE,string"`
	Longitude float64 `json:"LONGITUDE,string"`
	Postcode  string  `json:"POSTAL"`
}

type buildings []building

var postcodes buildings
var views = template.Must(template.ParseGlob("templates/*.html"))

func main() {

	var err error
	postcodes, err = loadBuildingJSON("buildings.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d buildings", len(postcodes))

	app := mux.NewRouter()
	app.HandleFunc("/", handleIndex)

	addr := ":" + os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(addr, app))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	postcode := r.URL.Query().Get("postcode")
	if len(postcode) != 6 {
		views.ExecuteTemplate(w, "index.html", nil)
		return
	}
	wantedResponse := r.URL.Query().Get("r")
	log.Println("Postcode", postcode, "Wanted Response", wantedResponse)

	b := postcodes.getLocation(postcode)

	if wantedResponse != "json" {
		http.Redirect(w, r,
			fmt.Sprintf("https://maps.google.com/?q=%f,%f", b.Latitude, b.Longitude), http.StatusSeeOther)
	} else {
		response.JSON(w, b)
	}
}

func (Buildings buildings) getLocation(postcode string) (b building) {
	for _, b = range Buildings {
		if postcode == b.Postcode {
			return b
		}
	}
	return
}

// curl -O https://raw.githubusercontent.com/xkjyeah/singapore-postal-codes/master/buildings.json
func loadBuildingJSON(jsonfile string) (bs buildings, err error) {
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
