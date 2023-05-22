package storageclient

type IStorageClient interface {
	StoreObject(objectName string, filePath string)
}
