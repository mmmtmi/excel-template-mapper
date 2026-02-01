package model

type Template struct {
	ID           string
	Name         string
	Target       string
	SheetName    *string
	HeaderRow    int
	DataStartRow int
}
