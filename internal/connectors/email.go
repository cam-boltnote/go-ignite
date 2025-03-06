package connectors

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/mail.v2"
)

// EmailSender handles sending emails using SMTP
type EmailSender struct {
	dialer  *mail.Dialer
	from    string
	Enabled bool
}

// IsEnabled returns whether the email sender is enabled
func (e *EmailSender) IsEnabled() bool {
	return e != nil && e.Enabled
}

// NewEmailSender creates a new instance of EmailSender
func NewEmailSender() (*EmailSender, error) {
	log.Println("Initializing EmailSender...")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading .env file: %v", err)
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Get required environment variables
	requiredVars := []string{"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM_EMAIL"}
	config := make(map[string]string)

	for _, v := range requiredVars {
		if value := os.Getenv(v); value != "" {
			config[v] = value
			// Log the first few characters of sensitive info
			if v == "SMTP_PASSWORD" {
				log.Printf("%s: %s...", v, value[:min(len(value), 3)])
			} else {
				log.Printf("%s: %s", v, value)
			}
		} else {
			log.Printf("SMTP configuration missing: %s. Email functionality will be disabled.", v)
			return &EmailSender{Enabled: false}, nil
		}
	}

	port, err := strconv.Atoi(config["SMTP_PORT"])
	if err != nil {
		log.Printf("Invalid SMTP port: %v. Email functionality will be disabled.", err)
		return &EmailSender{Enabled: false}, nil
	}

	dialer := mail.NewDialer(
		config["SMTP_HOST"],
		port,
		config["SMTP_USERNAME"],
		config["SMTP_PASSWORD"],
	)

	// For Gmail SMTP relay, use TLS instead of SSL
	dialer.SSL = false
	dialer.TLSConfig = &tls.Config{ServerName: config["SMTP_HOST"]}

	// Test the connection
	s, err := dialer.Dial()
	if err != nil {
		log.Printf("Failed to connect to SMTP server: %v. Email functionality will be disabled.", err)
		return &EmailSender{Enabled: false}, nil
	}
	s.Close()

	log.Println("EmailSender initialized successfully - SMTP connection test passed")
	return &EmailSender{
		dialer:  dialer,
		from:    config["SMTP_FROM_EMAIL"],
		Enabled: true,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SendEmail sends an email with the given parameters
func (e *EmailSender) SendEmail(to string, subject string, body string) error {
	if !e.Enabled {
		log.Printf("Email functionality is disabled. Skipping email to: %s", to)
		return nil
	}

	log.Printf("Starting email send process to: %s", to)
	log.Printf("Using SMTP configuration - Host: %s, Port: %d, Username: %s, From: %s",
		e.dialer.Host, e.dialer.Port, e.dialer.Username, e.from)

	m := mail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Add retry logic
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempt %d: Trying to send email...", i+1)

		// Create a connection to test SMTP settings
		s, err := e.dialer.Dial()
		if err != nil {
			log.Printf("Failed to connect to SMTP server: %v", err)
			lastErr = err
			continue
		}
		s.Close()

		if err := e.dialer.DialAndSend(m); err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to send email: %v", i+1, err)
			log.Printf("Email sending failed: %v. Retrying...", lastErr)
			continue
		}
		log.Printf("Email sent successfully to %s", to)
		return nil
	}

	log.Printf("All attempts to send email failed. Last error: %v", lastErr)
	return fmt.Errorf("failed to send email after %d attempts: %v", maxRetries, lastErr)
}

// SendPasswordReset sends a password reset email
func (e *EmailSender) SendPasswordReset(to string, resetToken string, resetURL string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2c3e50;">Password Reset Request</h2>
				<p>You have requested to reset your password. Click the link below to proceed:</p>
				<p style="margin: 25px 0;">
					<a href="%s" style="background-color: #3498db; color: white; padding: 12px 25px; text-decoration: none; border-radius: 4px;">Reset Password</a>
				</p>
				<p style="color: #7f8c8d; font-size: 0.9em;">If you didn't request this, please ignore this email.</p>
				<p style="color: #7f8c8d; font-size: 0.9em;">This link will expire in 1 hour.</p>
			</div>
		</body>
		</html>
	`, resetURL)

	return e.SendEmail(to, subject, body)
}

// SendFollowUpReminder sends a reminder email for follow-up items
func (e *EmailSender) SendFollowUpReminder(to string, entryTitle string, dueDate string) error {
	subject := "Follow-up Reminder"
	body := fmt.Sprintf(`
		<h2>Follow-up Reminder</h2>
		<p>This is a reminder for your entry: <strong>%s</strong></p>
		<p>Due date: %s</p>
		<p>Please check your activity tracker for more details.</p>
	`, entryTitle, dueDate)

	return e.SendEmail(to, subject, body)
}
