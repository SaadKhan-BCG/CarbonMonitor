package carbon

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	errorhandler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
)

var region string

func SetRegion(r string) {
	region = r
}

func ComputeCarbonConsumptionByTime(containerCarbon map[ContainerRegion]float64, container string, power float64, hour string, wg *sync.WaitGroup) {
	defer wg.Done()
	h, err := strconv.ParseInt(hour, 10, 64)
	if err != nil {
		errorhandler.StdErrorHandler(fmt.Sprintf("Failed as could not parse hour value %s as int", hour), err)
		return
	}

	currentTime := time.Now()
	currentLocation := currentTime.Location()

	// Take the day before as the reference since today's values are not yet fully available
	yesterdayTime := currentTime.AddDate(0, 0, -1)
	measureStartTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), int(h), 0, 0, 0, currentLocation)

	carbon, err := carbonemissions.GetCarbonEmissionsByTime(region, measureStartTime)
	if err != nil {
		errorhandler.StdErrorHandler(fmt.Sprintf("Failed fetching emissions data for Container: %s Region: %s Hour: H%s ", container, region, hour), err)
	} else {
		computeAndUpdateCarbonConsumption(containerCarbon, container, power, hour, carbon)
	}
}
