package model

type RequestOTP struct {
	EmailAddress string `json:"email_address"`
}

type ValidateOTP struct {
	EmailAddress string `json:"email_address"`
	OTP          string `json:"otp"`
}
