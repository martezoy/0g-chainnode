package types

type DASRequest struct {
	RequestId       uint64 `json:"request_id"`
	StreamId        string `json:"stream_id"`
	BatchHeaderHash string `json:"batch_header_hash"`
	NumBlobs        uint64 `json:"num_blobs"`
}
