package handlers

type uploadTemplate struct {
	MinDate     string
	DefaultDate string
	MaxDate     string
	UserID      string
}

type downloadTemplate struct {
	ByteSize int
	UserID   string
	FileID   string
}
