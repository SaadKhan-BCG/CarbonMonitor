package carbon

import (
	"fmt"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	errorhandler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
	"sync"
)

func ComputeCurrentCarbonConsumption(containerCarbon map[ContainerRegion]float64, container string, power float64, region string, wg *sync.WaitGroup) {
	defer wg.Done()
	carbon, err := carbonemissions.GetCurrentCarbonEmissions(region)
	if err != nil {
		errorhandler.StdErrorHandler(fmt.Sprintf("Failed fetching emissions data for Container: %s Region: %s ", container, region), err)
	} else {
		computeAndUpdateCarbonConsumption(containerCarbon, container, power, region, carbon)
	}
}
