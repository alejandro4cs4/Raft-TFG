package storageclient

type IStorageClient interface {
	CloseClient()
	StoreObject()
}
