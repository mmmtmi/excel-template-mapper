package model

type Rule struct {
	SourceType  string
	SourceKey   string
	TargetLabel string
	Transform   *string
	Required    bool
	Priority    int
}
