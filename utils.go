package carbon

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/container_stats"
	"github.com/docker/docker/client"
	"github.com/gosuri/uilive"
	log "github.com/sirupsen/logrus"
)

// Outputs live updating console log of Carbon Data as stored in "totalCarbon"
// Note, outputs should be a list of keys in totalCarbon you wish to log
// Iterate over outputs instead of totalCarbon keys as map key ordering is not consistent so the output is harder to read
func outputLiveConsumption(writer *uilive.Writer, totalCarbon map[string]float64, outputs *[]string, name string) {
	for _, output := range *outputs {
		_, _ = fmt.Fprintf(writer, "%s: %s Carbon Consumed: %fmgCo2Eq\n", name, output, totalCarbon[output])
	}
}

/*
		OutputTotalCarbon Generic function to Output Total Carbon split by some iterable, ie region or time ranges
		@param: iterableName Name of the variable you are iterating over, only used in logging
		@param: Iterable The variable you iterate over e.g. regions, time zones. Note we will refer to an instance in this as "item"
		@param computeFn: Function which takes as input containerCarbon -> map(containerName, item) -> carbon consumed, containerName, power (consumed by container), item.
	                      Should update containerCarbon with the correct carbon value for containerName, item tuple
*/
//goland:noinspection GoPrintFunctions
func OutputTotalCarbon(iterableName string, iterable *[]string, computeFn func(map[ContainerRegion]float64, string, float64, string, *sync.WaitGroup)) {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal("Failed to Initialise Docker Client", err)
	}

	wgCount = len(*iterable)

	var iterationDurationInSeconds int64
	var totalTimeInSeconds int64
	var lastIterationStartTimeUnix int64
	containerPower := make(map[string]float64)
	containerCarbon := make(map[ContainerRegion]float64)
	totalCarbon := make(map[string]float64)

	writer := uilive.New()
	writer.Start()

	fmt.Println("Total Carbon consumption of running containers:")
	startTimeUnix := time.Now().Unix()
	for {
		lastIterationStartTimeUnix = time.Now().Unix()
		container_stats.GetDockerStats(cli, containerPower)
		carbon_emissions.RefreshCarbonCache()

		log.Debug(containerPower)
		for container := range containerPower {
			log.Debug(containerPower[container])
			var wg sync.WaitGroup
			wg.Add(wgCount)
			for _, item := range *iterable {
				go computeFn(containerCarbon, container, containerPower[container], item, &wg)
			}
			wg.Wait()
		}
		log.Debug(containerCarbon)

		iterationDurationInSeconds = time.Now().Unix() - lastIterationStartTimeUnix
		totalTimeInSeconds = time.Now().Unix() - startTimeUnix

		log.Debug(fmt.Sprintf("Time taken for iteration: %v Seconds", iterationDurationInSeconds))
		log.Debug(fmt.Sprintf("Total Time Spend: %v Seconds", totalTimeInSeconds))

		for _, item := range *iterable {
			for container := range containerPower {
				totalCarbon[item] += containerCarbon[ContainerRegion{container, item}] * float64(iterationDurationInSeconds)
			}
		}

		// Note, Live logging requires the loglevel be set to error as info logging gets in the way
		outputLiveConsumption(writer, totalCarbon, iterable, iterableName)
		log.Debug(totalCarbon)

		// Empty the maps at the end of every iteration to prevent old reports staying
		containerPower = make(map[string]float64)
		containerCarbon = make(map[ContainerRegion]float64)
	}

	writer.Stop()
}

func computeAndUpdateCarbonConsumption(containerCarbon map[ContainerRegion]float64, container string, power float64, item string, carbon float64) {
	carbonConsumed := power * carbon * 10 / 216 // Carbon is in gCo2/H converting here to mgCo2/S
	log.Debug(fmt.Sprintf("Region: %s Rating: %f Power: %f", region, carbon, power))
	mutex.Lock() // Map write operations are not thread safe and this function is called in parallel
	containerCarbon[ContainerRegion{container, item}] = carbonConsumed
	mutex.Unlock()
}

func GetOrElsEnvVars(ENV_VAR string, defaultVar string) string {
	envVar := os.Getenv(ENV_VAR)
	if len(envVar) > 0 {
		return envVar
	}

	os.Setenv(ENV_VAR, defaultVar)
	return defaultVar
}
