package models

// LibreViewCredentials contains authentication information for the LibreView API
type LibreViewCredentials struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

// NightscoutCredentials contains authentication information for the Nightscout API
type NightscoutCredentials struct {
	URL       string `json:"url" bson:"url"`
	APISecret string `json:"apiSecret" bson:"apiSecret"`
}
