package handlers

type uploadTemplate struct {
	AppRoot string
}

type downloadTemplate struct {
	AppRoot  string
	ByteSize int64
	UserID   string
	FileID   string
}
