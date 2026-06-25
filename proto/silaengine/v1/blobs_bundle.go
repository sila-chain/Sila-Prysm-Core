package silaenginev1

type BlobsBundler interface {
	GetKzgCommitments() [][]byte
	GetProofs() [][]byte
	GetBlobs() [][]byte
}
