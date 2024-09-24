package i18n

//go:generate go run ../../cmd/genschema/genschema.go

// Messages defines all the i18n strings to be localized.
type Messages struct {
	Email struct {
		MissedCallSubject string `json:"missedCallSubject"`
		MissedCallContent string `json:"missedCallContent"`
		NameFrom          string `json:"nameFrom"`
		NameTo            string `json:"nameTo"`
		VoicemailSubject  string `json:"voicemailSubject"`
		VoicemailContent  string `json:"voicemailContent"`
	} `json:"email"`
	Voice struct {
		AcceptCall       string `json:"acceptCall"`
		ConfirmConnected string `json:"confirmConnected"`
		LangSelect       string `json:"langSelect"`
		PleaseHold       string `json:"pleaseHold"`
		RecordAfterTone  string `json:"recordAfterTone"`
		ReRecord         string `json:"rerecord"`
		Voicemail        string `json:"voicemail"`
		VoicemailRepeat  string `json:"voicemailRepeat"`
		Welcome          string `json:"welcome"`
	} `json:"voice"`
}
