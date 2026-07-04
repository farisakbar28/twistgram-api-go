package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
	"twistgram-api-go/internal/config"
)

// SendEmail mengirim email SMTP sederhana (via Mailtrap dll).
func SendEmail(to, subject, body string) error {
	cfg := config.LoadConfig()
	if cfg.SMTPHost == "" {
		// Bypass pengiriman jika SMTP tidak diset, tapi print ke console (Berguna untuk testing/MVP)
		fmt.Printf("--- EMAIL DISPATCH BYPASS ---\nTo: %s\nSubject: %s\nBody: %s\n-----------------------------\n", to, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	address := cfg.SMTPHost + ":" + cfg.SMTPPort

	headers := make(map[string]string)
	headers["From"] = cfg.SMTPFromEmail
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + body)

	return smtp.SendMail(address, auth, cfg.SMTPFromEmail, []string{to}, msg.Bytes())
}

func SendOTPEmail(to, otp string) error {
	subject := "Kode OTP Verifikasi Twistgram Anda"
	t, _ := template.New("otp").Parse(`
		<h2>Kode Verifikasi</h2>
		<p>Halo,</p>
		<p>Gunakan kode berikut untuk memverifikasi akun Anda:</p>
		<h1 style="background:#eee;padding:10px;text-align:center;letter-spacing:5px;">{{.OTP}}</h1>
		<p>Kode ini berlaku selama 10 menit. Jangan berikan kode ini kepada siapapun.</p>
	`)
	var body bytes.Buffer
	t.Execute(&body, map[string]string{"OTP": otp})
	return SendEmail(to, subject, body.String())
}
