package main

import (
	carbon "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)
	//carbon.RegionMode(&[]string{"uksouth", "foo"})
	carbon.RegionMode(&[]string{"westus"})
	//carbon.GraphMode("westus")
	//carbon.TimeMode("westus")
	//regions := strings.Join(carbon.ListValidRegions(), "\n")
	//log.Println("Available regions for measuring carbon consumption:")
	//log.Println(regions)
}
