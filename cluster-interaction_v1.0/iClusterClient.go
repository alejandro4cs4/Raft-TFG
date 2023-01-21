package main

type IClusterClient interface {
	closeClient()
	clearAllClusterData(measureElapsedTime bool)
	clearClusterDataOneByOne(exploreDirectory string)
	storeDataInCluster(exploreDirectory string)
	listClusterData()
}