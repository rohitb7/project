package core

type StorageManagerProgressEvent struct {
	Progress float64
}

type OperationState int

const (
	OperationUnknown OperationState = iota
	OperationSuccess
	OperationFailure
)

type StorageManagerOperationEvent struct {
	Operation OperationState
	Err       error
}

type HierarchyIdentifier struct {
	Bucket string
}

type FileInfo struct {
	Size     int64
	MetaData map[string]string
}

type BlobContext struct {
	RemotePathKey       string
	HierarchyIdentifier HierarchyIdentifier
}

func (b *BlobContext) GetRemotePathKey() string {
	if b == nil {
		return ""
	}
	return b.RemotePathKey
}

func (b *BlobContext) GetBucket() string {
	if b == nil {
		return ""
	}
	return b.HierarchyIdentifier.Bucket
}
