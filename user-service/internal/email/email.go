package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

type brevoEmail struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HtmlContent string         `json:"htmlContent"`
	TextContent string         `json:"textContent"`
}

type brevoContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func sendBrevo(apiKey, senderEmail string, msg brevoEmail) {
	body, err := json.Marshal(msg)
	if err != nil {
		slog.Error("email.marshal_failed", "error", err.Error())
		return
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		slog.Error("email.request_failed", "error", err.Error())
		return
	}
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("email.send_failed", "error", err.Error(), "to", msg.To[0].Email)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Error("email.rejected", "status", resp.StatusCode, "to", msg.To[0].Email)
		return
	}
	slog.Info("email.sent", "to", msg.To[0].Email)
}

func SendVerificationEmailAsync(toEmail, firstName, verifyURL string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("email.panic", "error", r)
			}
		}()

		apiKey := os.Getenv("BREVO_API_KEY")
		senderEmail := os.Getenv("SENDER_EMAIL")
		if apiKey == "" || senderEmail == "" {
			slog.Warn("email.config_missing", "note", "Brevo env tidak diset, skip kirim email")
			return
		}

		sendBrevo(apiKey, senderEmail, brevoEmail{
			Sender:  brevoContact{Name: "Majadigi", Email: senderEmail},
			To:      []brevoContact{{Name: firstName, Email: toEmail}},
			Subject: "Verifikasi Email Akun Majadigi Anda",
			TextContent: fmt.Sprintf("Halo %s, klik link berikut untuk verifikasi email: %s", firstName, verifyURL),
			HtmlContent: fmt.Sprintf(
				`<h3>Halo %s,</h3>
<p>Terima kasih telah mendaftar di <b>Majadigi</b>.</p>
<p>Klik tombol di bawah untuk memverifikasi akun Anda:</p>
<p><a href="%s" style="background:#1d4ed8;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;">Verifikasi Email</a></p>
<p>Link ini berlaku selama 24 jam.</p>
<hr><small>Tim Majadigi Jawa Timur</small>`,
				firstName, verifyURL,
			),
		})
	}()
}

func SendPasswordResetEmailAsync(toEmail, firstName, resetURL string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("email.panic", "error", r)
			}
		}()

		apiKey := os.Getenv("BREVO_API_KEY")
		senderEmail := os.Getenv("SENDER_EMAIL")
		if apiKey == "" || senderEmail == "" {
			slog.Warn("email.config_missing", "note", "Brevo env tidak diset, skip kirim email")
			return
		}

		sendBrevo(apiKey, senderEmail, brevoEmail{
			Sender:  brevoContact{Name: "Majadigi", Email: senderEmail},
			To:      []brevoContact{{Name: firstName, Email: toEmail}},
			Subject: "Reset Password Akun Majadigi Anda",
			TextContent: fmt.Sprintf("Halo %s, klik link berikut untuk reset password: %s (berlaku 15 menit)", firstName, resetURL),
			HtmlContent: fmt.Sprintf(
				`<h3>Halo %s,</h3>
<p>Kami menerima permintaan untuk mereset password akun <b>Majadigi</b> Anda.</p>
<p>Klik tombol di bawah untuk membuat password baru:</p>
<p><a href="%s" style="background:#dc2626;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;">Reset Password</a></p>
<p><b>Link ini hanya berlaku selama 15 menit.</b></p>
<p>Jika Anda tidak meminta reset password, abaikan email ini.</p>
<hr><small>Tim Majadigi Jawa Timur</small>`,
				firstName, resetURL,
			),
		})
	}()
}
