package metadataclient

type IMetadataClient interface {
	CloseClient()
	StoreKeyValue(key string, value string)
	GetByKey(key string) MetadataClientGetResponse
}
