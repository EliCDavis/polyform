package endpoint

type ContentType string

const (
	BinaryContentType    ContentType = "application/octet-stream"
	JsonContentType      ContentType = "application/json"
	PlainTextContentType ContentType = "text/plain"
)
