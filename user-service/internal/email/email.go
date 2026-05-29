package email

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendVerificationEmailAsync(toEmail, firstName, verifyURL string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("email.panic", "error", r)
			}
		}()

		apiKey := os.Getenv("SENDGRID_API_KEY")
		senderEmail := os.Getenv("SENDGRID_SENDER_EMAIL")
		if apiKey == "" || senderEmail == "" {
			slog.Warn("email.config_missing", "note", "SendGrid env tidak diset, skip kirim email")
			return
		}

		from := mail.NewEmail("Majadigi", senderEmail)
		to := mail.NewEmail(firstName, toEmail)
		subject := "Verifikasi Email Akun Majadigi Anda"
		plain := fmt.Sprintf("Halo %s, klik link berikut untuk verifikasi email: %s", firstName, verifyURL)
		html := fmt.Sprintf(
			`<h3>Halo %s,</h3>
<p>Terima kasih telah mendaftar di <b>Majadigi</b>.</p>
<p>Klik tombol di bawah untuk memverifikasi akun Anda:</p>
<p><a href="%s" style="background:#1d4ed8;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;">Verifikasi Email</a></p>
<p>Link ini berlaku selama 24 jam.</p>
<hr><small>Tim Majadigi Jawa Timur</small>`,
			firstName, verifyURL,
		)

		msg := mail.NewSingleEmail(from, subject, to, plain, html)
		client := sendgrid.NewSendClient(apiKey)
		resp, err := client.Send(msg)
		switch {
		case err != nil:
			slog.Error("email.send_failed", "error", err.Error(), "to", toEmail)
		case resp.StatusCode >= 400:
			slog.Error("email.rejected", "status", resp.StatusCode, "body", resp.Body)
		default:
			slog.Info("email.sent", "to", toEmail)
		}
	}()
}

func SendPasswordResetEmailAsync(toEmail, firstName, resetURL string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("email.panic", "error", r)
			}
		}()

		apiKey := os.Getenv("SENDGRID_API_KEY")
		senderEmail := os.Getenv("SENDGRID_SENDER_EMAIL")
		if apiKey == "" || senderEmail == "" {
			slog.Warn("email.config_missing", "note", "SendGrid env tidak diset, skip kirim email")
			return
		}

		from := mail.NewEmail("Majadigi", senderEmail)
		to := mail.NewEmail(firstName, toEmail)
		subject := "Reset Password Akun Majadigi Anda"
		plain := fmt.Sprintf("Halo %s, klik link berikut untuk reset password: %s (berlaku 15 menit)", firstName, resetURL)
		html := fmt.Sprintf(
			`<h3>Halo %s,</h3>
<p>Kami menerima permintaan untuk mereset password akun <b>Majadigi</b> Anda.</p>
<p>Klik tombol di bawah untuk membuat password baru:</p>
<p><a href="%s" clicktracking="off" style="background:#dc2626;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;">Reset Password</a></p>
<p><b>Link ini hanya berlaku selama 15 menit.</b></p>
<p>Jika Anda tidak meminta reset password, abaikan email ini.</p>
<hr><small>Tim Majadigi Jawa Timur</small>`,
			firstName, resetURL,
		)

		msg := mail.NewSingleEmail(from, subject, to, plain, html)
		client := sendgrid.NewSendClient(apiKey)
		resp, err := client.Send(msg)
		switch {
		case err != nil:
			slog.Error("email.send_failed", "error", err.Error(), "to", toEmail)
		case resp.StatusCode >= 400:
			slog.Error("email.rejected", "status", resp.StatusCode, "body", resp.Body)
		default:
			slog.Info("email.sent", "to", toEmail)
		}
	}()
}
