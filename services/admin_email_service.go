package services

import (
	"fmt"
	"strconv"

	"github.com/go-gomail/gomail"
)

type EmailService struct {
	SettingsService *SettingsService
}

func NewEmailService(settingsService *SettingsService) *EmailService {
	return &EmailService{SettingsService: settingsService}
}

func (s *EmailService) SendEmail(to string, subject string, body string) error {
	// Fetch SMTP settings from the settings table
	smtpHost, err := s.SettingsService.GetSetting("smtp_host")
	if err != nil {
		return err
	}
	smtpPortStr, err := s.SettingsService.GetSetting("smtp_port")
	if err != nil {
		return err
	}
	smtpEmail, err := s.SettingsService.GetSetting("smtp_email")
	if err != nil {
		return err
	}
	smtpPassword, err := s.SettingsService.GetSetting("smtp_password")
	if err != nil {
		return err
	}
	// Convert smtpPort string to int
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

	// Prepare the email message
	m := gomail.NewMessage()
	m.SetHeader("From", smtpEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Create the SMTP dialer
	d := gomail.NewDialer(smtpHost, smtpPort, smtpEmail, smtpPassword)

	// Send the email
	return d.DialAndSend(m)
}
