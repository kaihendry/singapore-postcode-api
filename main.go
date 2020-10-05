package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	logger := log.New(os.Stderr, "", log.Lshortfile)

	server, err := NewServer(logger, "templates/*.html", "buildings.json")
	if err != nil {
		logger.Printf("failed to create server: %v", err)
		os.Exit(1)
	}

	err = http.ListenAndServe(":"+os.Getenv("PORT"), server)
	if err != nil {
		logger.Printf("failed to start server: %v", err)
		os.Exit(1)
	}
}

type Log interface {
	Printf(msg string, args ...interface{})
	Println(args ...interface{})
}

type Server struct {
	log       Log
	views     *template.Template
	postcodes Buildings
}

func NewServer(logger Log, templatesPath, postcodesPath string) (*Server, error) {
	views, err := template.ParseGlob(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates from %q: %w", templatesPath, err)
	}

	postcodes, err := BuildingsFromFile(postcodesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %q: %w", postcodesPath, err)
	}
	logger.Printf("Loaded %d buildings", len(postcodes))

	return &Server{
		log: logger,

		views:     views,
		postcodes: postcodes,
	}, nil
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	postcode := r.URL.Query().Get("postcode")
	if len(postcode) != 6 {
		server.views.ExecuteTemplate(w, "index.html", nil)
		return
	}

	wantedResponse := r.URL.Query().Get("r")
	server.log.Println("Postcode", postcode, "Wanted Response", wantedResponse)

	b, ok := server.postcodes.Lookup(postcode)
	if !ok {
		// handle missing postcode
		return
	}

	if wantedResponse == "gmap" {
		redirect := fmt.Sprintf("https://maps.google.com/?q=%f,%f", b.Latitude, b.Longitude)
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	} else {
		geojson := geoPoint(b)
		server.writeJSON(w, geojson)
	}
}

func geoPoint(data Building) (geo interface{}) {
	return struct {
		Type       string "json:\"type\""
		Properties struct {
			Place string "json:\"Place\""
		} "json:\"properties\""
		Geometry struct {
			Type        string    "json:\"type\""
			Coordinates []float64 "json:\"coordinates\""
		} "json:\"geometry\""
	}{
		Type: "Feature",
		Properties: struct {
			Place string "json:\"Place\""
		}{
			Place: data.Building,
		},
		Geometry: struct {
			Type        string    "json:\"type\""
			Coordinates []float64 "json:\"coordinates\""
		}{
			Type:        "Point",
			Coordinates: []float64{data.Latitude, data.Longitude},
		},
	}
}

func (server *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(data)
}

type Buildings []Building

type Building struct {
	Latitude  float64 `json:"LATITUDE,string"`
	Longitude float64 `json:"LONGITUDE,string"`
	Building  string  `json:"BUILDING"`
	Postcode  string  `json:"POSTAL"`
}

// curl -O https://raw.githubusercontent.com/xkjyeah/singapore-postal-codes/master/buildings.json
func BuildingsFromFile(path string) (Buildings, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed read %q: %w", path, err)
	}

	var buildings Buildings
	err = json.Unmarshal(content, &buildings)
	if err != nil {
		return nil, fmt.Errorf("failed parse: %w", err)
	}

	return buildings, nil
}

func (buildings Buildings) Lookup(postcode string) (Building, bool) {
	for _, building := range buildings {
		if postcode == building.Postcode {
			return building, true
		}
	}

	return Building{}, false
}
