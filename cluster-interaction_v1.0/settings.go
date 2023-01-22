package main

type Settings struct {
	ClusterType string `yaml:"clusterType"`
	ExploreDirectory string `yaml:"exploreDirectory"`
	MaxItemsStore int64 `yaml:"maxItemsStore"`
}