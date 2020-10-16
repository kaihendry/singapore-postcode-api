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

// Buildings is postcode lookup map
type Buildings map[string]Building

type Building struct {
	Latitude  float64 `json:"LATITUDE,string"`
	Longitude float64 `json:"LONGITUDE,string"`
	Building  string  `json:"BUILDING"`
	Postcode  string  `json:"POSTAL"`
}

func main() {
	logger := log.New(os.Stderr, "", log.Lshortfile)

	server, err := NewServer(logger, "templates/*.html", "data/buildings.json")
	if err != nil {
		logger.Fatalf("failed to create server: %v", err)
	}

	err = http.ListenAndServe(":"+os.Getenv("PORT"), server)
	if err != nil {
		logger.Fatalf("failed to start server: %v", err)
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
		log:       logger,
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

	b, ok := server.postcodes[postcode]
	if !ok {
		server.log.Printf("Postcode not found: %s", postcode)
		http.Error(w, "Postcode not found", http.StatusBadRequest)
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

// rewrite to a GeoJSON geopoint
func geoPoint(data Building) (geo interface{}) {
	return struct {
		Type       string `json:"type"`
		Properties struct {
			Place string `json:"Place"`
		} `json:"properties"`
		Geometry struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	}{
		Type: "Feature",
		Properties: struct {
			Place string `json:"Place"`
		}{
			Place: data.Building,
		},
		Geometry: struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		}{
			Type:        "Point",
			Coordinates: []float64{data.Longitude, data.Latitude},
		},
	}
}

func (server *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(data)
}

func BuildingsFromFile(path string) (Buildings, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed read %q: %w", path, err)
	}

	var buildings []Building
	err = json.Unmarshal(content, &buildings)
	if err != nil {
		return nil, fmt.Errorf("failed parse: %w", err)
	}

	buildingmap := make(map[string]Building)

	for _, v := range buildings {
		buildingmap[v.Postcode] = v
	}

	return buildingmap, nil
}
