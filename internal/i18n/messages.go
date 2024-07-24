package i18n

//go:generate go run ../../cmd/genschema/genschema.go

// Messages defines all the i18n strings to be localized.
type Messages struct {
	Voice struct {
		GatherOutbound string `json:"gatherOutbound"`
		LangSelect     string `json:"langSelect"`
		Welcome        string `json:"welcome"`
	} `json:"voice"`
}
