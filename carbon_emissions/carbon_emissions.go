package carbon_emissions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	errorhandler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
)

var baseUrl string

// Map[region, emission rating] local cache to reduce network i/o
var carbonRegionCache map[string]float64

var mutex = &sync.Mutex{}

type carbonAwareResponse struct {
	Rating float64 `json:"rating"`
}

type errorResponse struct {
	Detail string `json:"detail"`
}

func handleResponse[_ io.ReadCloser, T []carbonAwareResponse | errorResponse](responseBody io.ReadCloser, data *T) error {
	defer responseBody.Close()
	body, err := io.ReadAll(responseBody)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}
	return nil
}

func getCarbonEmissions(region string, prevTime string, toTime string) (float64, error) {
	url := fmt.Sprintf("%s/emissions/bylocation?location=%s&time=%s&toTime=%s", baseUrl, region, prevTime, toTime)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	} else if resp.StatusCode != 200 {
		var errData errorResponse
		err = handleResponse[io.ReadCloser, errorResponse](resp.Body, &errData)
		if err != nil {
			return 0, err
		}

		return 0, errors.New(fmt.Sprintf("Invalid Response from Carbon SDK: %s \n Cause: %s", resp.Status, errData.Detail))
	}

	var data []carbonAwareResponse
	err = handleResponse[io.ReadCloser, []carbonAwareResponse](resp.Body, &data)
	if err != nil {
		return 0, err
	}
	return data[0].Rating, nil
}

func formatTimeAsString(time time.Time) string {
	return time.Format("2006-01-02T15:04")
}

func GetCarbonEmissionsByTime(region string, utcTime time.Time) (float64, error) {
	toTime := formatTimeAsString(utcTime)
	prevTime := formatTimeAsString(utcTime.Add(-time.Minute))

	rating := carbonRegionCache[region]
	if rating > 0 {
		return rating, nil
	} else {
		newRating, err := getCarbonEmissions(region, prevTime, toTime)
		if err != nil {
			return 0, err
		}
		mutex.Lock()
		carbonRegionCache[region] = newRating
		mutex.Unlock()
		return newRating, nil
	}
}

func GetCurrentCarbonEmissions(region string) (float64, error) {
	rating, err := GetCarbonEmissionsByTime(region, time.Now().UTC())
	if err != nil {
		errorhandler.StdErrorHandler(fmt.Sprintf("Failure fetching emission data for Region: %s", region), err)
		return 0, err
	} else {
		log.Debug(fmt.Sprintf("Region: %s Rating: %f", region, rating))
		return rating, nil
	}
}

func RefreshCarbonCache() {
	carbonRegionCache = make(map[string]float64)
}

func LoadSettings() {
	log.Info("Loading settings")
	RefreshCarbonCache()

	url := os.Getenv("CARBON_SDK_URL")
	host := os.Getenv("CARBON_SDK_HOST")
	port := os.Getenv("CARBON_SDK_PORT")
	if len(url) > 4 {
		baseUrl = url
	} else {
		if host == "" || port == "" {
			log.Fatal("Error loading env variables, please set either CARBON_SDK_URL or CARBON_SDK_HOST and CARBON_SDK_PORT")
		}
		baseUrl = fmt.Sprintf("http://%s:%s", host, port)
	}
	log.Info(fmt.Sprintf("Using Carbon Aware SDK at %s", baseUrl))
}
