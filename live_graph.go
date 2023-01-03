package carbon

import (
	"context"
	"fmt"
	"time"

	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/container_stats"
	errorhandler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gosuri/uilive"
	ag "github.com/guptarohit/asciigraph"
	"github.com/sirupsen/logrus"
)

var colours = ag.SeriesColors(
	ag.Red,
	ag.Yellow,
	ag.Green,
	ag.Blue,
	ag.Purple,
	ag.Red,
	ag.Yellow,
	ag.Green,
	ag.Blue,
	ag.Purple)

var colourNames = []string{
	"Red",
	"Yellow",
	"Green",
	"Blue",
	"Purple",
	"Red",
	"Yellow",
	"Green",
	"Blue",
	"Purple",
}

// Warning Ansi colours are not a standard format so there is no guarantee this will work everywhere
// Please rely on the colourNames as source of truth
var ansiColours = []string{
	"\033[31m", // Red
	"\033[33m", // Yellow
	"\033[32m", // Green
	"\033[34m", // Blue
	"\033[35m", // Purple
	"\033[31m", // Red
	"\033[33m", // Yellow
	"\033[32m", // Green
	"\033[34m", // Blue
	"\033[35m", // Purple
}

var ansiDefault = "\033[0m"

var colourCount = len(colourNames)

func asciPlot(region string) {
	cli, err := client.NewEnvClient()
	if err != nil {
		logrus.Fatal(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: false,
	})

	n := 0
	for _, container := range containers { // Filter to ignore carbon-plugin containers
		if !container_stats.FilterPluginContainers(container.Names[0]) {
			containers[n] = container
			n++
		}
	}
	containers = containers[:n]

	// Need a static list of names for the line graph, the maps they're stored in lower don't respect ordering
	var containerNames []string
	for i, container := range containers {
		name := container.Names[0]
		containerNames = append(containerNames, name)
		colourName := "Grey"
		ansiColour := ansiDefault
		if i < colourCount {
			colourName = colourNames[i]
			ansiColour = ansiColours[i]
		}
		legend := fmt.Sprintf("Container: %s Colour: %s\n", name, colourName)
		fmt.Print(ansiColour, legend)

	}
	fmt.Println(ansiDefault, "Live Carbon Consumption by Container (mgCo2Eq/s): ") // Reset colour

	if err != nil {
		logrus.Fatal("Failed to Initialise Docker Client", err)
	}

	var iterationDurationInSeconds int64
	var lastIterationStartTimeUnix int64
	containerPower := make(map[string]float64)
	graphData := make([][]float64, len(containerNames))
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()

	for {
		lastIterationStartTimeUnix = time.Now().Unix()
		container_stats.GetDockerStats(cli, containerPower)
		carbon_emissions.RefreshCarbonCache()
		carbon, err := carbon_emissions.GetCurrentCarbonEmissions(region)

		if err != nil {
			errorhandler.StdErrorHandler(fmt.Sprintf("Failed to get carbon data for region: %s", region), err)
		} else {
			for index, container := range containerNames {
				power := containerPower[container]

				carbonConsumed := float64(iterationDurationInSeconds) * power * carbon * 10 / 216 // Carbon is in gCo2/H converting here to mgCo2/S
				logrus.Debug(fmt.Sprintf("Region: %s Rating: %f Power: %f", region, carbon, power))
				graphData[index] = append(graphData[index], carbonConsumed)
			}
		}

		graph := ag.PlotMany(
			graphData,
			ag.Precision(0),
			ag.Height(10),
			ag.Width(30),
			colours)

		fmt.Fprintln(writer, graph)

		iterationDurationInSeconds = time.Now().Unix() - lastIterationStartTimeUnix
	}
}
