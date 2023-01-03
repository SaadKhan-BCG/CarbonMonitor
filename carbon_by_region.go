package carbon

import (
	"fmt"
	"sync"

	carbonemissions "github.com/SaadKhan-BCG/CarbonMonitor/carbon_emissions"
	errorhandler "github.com/SaadKhan-BCG/CarbonMonitor/error_handler"
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
