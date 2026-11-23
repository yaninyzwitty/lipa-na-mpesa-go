package daraja

import (
	"os"
	"strings"
)

type Config struct {
	BaseURL         string
	ConsumerKey     string
	ConsumerSecret  string
	ShortCode       string
	ResponseType    string
	ConfirmationURL string
	ValidationURL   string
}

func ConfigFromEnv() Config {
	return Config{
		BaseURL:         strings.TrimRight(os.Getenv("DARAJA_SANDBOX_URL"), "/"),
		ConsumerKey:     os.Getenv("DARAJA_CONSUMER_KEY"),
		ConsumerSecret:  os.Getenv("DARAJA_CONSUMER_SECRET"),
		ShortCode:       os.Getenv("DARAJA_C2B_SHORTCODE"),
		ResponseType:    os.Getenv("DARAJA_C2B_RESPONSE_TYPE"),
		ConfirmationURL: os.Getenv("DARAJA_C2B_CONFIRMATION_URL"),
		ValidationURL:   os.Getenv("DARAJA_C2B_VALIDATION_URL"),
	}
}
