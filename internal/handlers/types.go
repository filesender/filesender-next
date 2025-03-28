package handlers

type uploadTemplate struct {
	MinDate     string
	DefaultDate string
	MaxDate     string
}

type uploadDoneTemplate struct {
	UserID     string
	TransferID string
	FileCount  int
	BytesSize  int
}

type getTransferTemplate struct {
	FileCount int
	ByteSize  int
	Files     []getTransferTemplateFile
}

type getTransferTemplateFile struct {
	FileName string
	ByteSize int
}
