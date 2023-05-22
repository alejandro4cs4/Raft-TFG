package metadataclient

type IMetadataClient interface {
	CloseClient()
	StoreKeyValue()
}
