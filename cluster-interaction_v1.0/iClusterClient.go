package main

type IClusterClient interface {
	clearAllClusterData(measureElapsedTime bool)
	clearClusterDataOneByOne()
	storeDataInCluster()
	listClusterData()
}