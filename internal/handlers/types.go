package handlers

type uploadTemplate struct {
	AppRoot     string
	MinDate     string
	DefaultDate string
	MaxDate     string
	UserID      string
}

type downloadTemplate struct {
	AppRoot  string
	ByteSize int64
	UserID   string
	FileID   string
}
