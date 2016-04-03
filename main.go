package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/skiesel/thermometers/sensors"
)

var (
	lastSentEmail = time.Now()
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/getTemperatures", getTemperatures)
	http.HandleFunc("/temperatureOutOfRange", temperatureOutOfRange)
	http.HandleFunc("/", renderIndex)

	http.ListenAndServe(":8080", nil)
}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func getTemperatures(w http.ResponseWriter, r *http.Request) {
	temperatures := []float64{}
	readings := sensors.GetThermometerReadings()
	for _, reading := range readings {
		temperatures = append(temperatures, reading.Fahrenheit)
	}

	js, err := json.Marshal(temperatures)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func temperatureOutOfRange(w http.ResponseWriter, r *http.Request) {
	type outOfRange struct {
		Which int64   `json:"which"`
		Value float64 `json:"value"`
		Min   float64 `json:"min"`
		Max   float64 `json:"max"`
	}

	decoder := json.NewDecoder(r.Body)
	var data outOfRange
	err := decoder.Decode(&data)
	if err != nil {
		panic(err)
	}

	body := fmt.Sprintf("Thermometer %d measured %g which is outside of specified range of %g - %g", data.Which, data.Value, data.Min, data.Max)

	if time.Now().Sub(lastSentEmail).Minutes() > 5 {
		lastSentEmail = time.Now()
		SendMail("Designated Brewer: Temperature Monitor Alert", body)
	}

}
