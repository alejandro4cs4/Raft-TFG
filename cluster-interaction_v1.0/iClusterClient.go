package main

type IClusterClient interface {
	closeClient()
	clearAllClusterData(measureElapsedTime bool)
	clearClusterDataOneByOne()
	storeDataInCluster()
	listClusterData()
}