package handlers

type uploadTemplate struct {
	MinDate     string
	DefaultDate string
	MaxDate     string
}

type uploadDoneTemplate struct {
	UserID    string
	FileID    string
	BytesSize int
}

type downloadTemplate struct {
	ByteSize int
	UserID   string
	FileID   string
}
