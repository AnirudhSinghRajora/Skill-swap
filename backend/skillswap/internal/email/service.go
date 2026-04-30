package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
)

// Service handles sending emails via the Resend API.
type Service struct {
	apiKey      string
	fromEmail   string
	frontendURL string
}

// NewService creates a new email service from config.
func NewService(cfg config.Config) *Service {
	return &Service{
		apiKey:      cfg.ResendAPIKey,
		fromEmail:   cfg.FromEmail,
		frontendURL: cfg.FrontendURL,
	}
}

// IsConfigured returns true if the Resend API key is set.
func (s *Service) IsConfigured() bool {
	return s.apiKey != ""
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *Service) send(to, subject, html string) error {
	if !s.IsConfigured() {
		return nil
	}

	payload := resendPayload{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SendWelcome sends a welcome email to a newly registered user.
func (s *Service) SendWelcome(to, name string) error {
	subject := "Welcome to SkillSwap!"
	html := fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:0 auto">
<h1 style="color:#4f46e5">Welcome to SkillSwap!</h1>
<p>Hi %s,</p>
<p>Thanks for joining SkillSwap. You're now part of a community of people who love to learn and teach.</p>
<p>Get started by browsing skills and connecting with other learners.</p>
<a href="%s/browse" style="display:inline-block;background:#4f46e5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;margin-top:16px">Browse Skills</a>
</div>`, name, s.frontendURL)
	return s.send(to, subject, html)
}

// SendEmailVerification sends an email with a verification link. The link
// points at the frontend /auth/verify page (not the backend API), so the
// user lands on a styled page that calls the API and shows success/error
// feedback.
func (s *Service) SendEmailVerification(to, name, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", s.frontendURL, token)
	subject := "Verify your SkillSwap email"
	html := fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:0 auto">
<h1 style="color:#4f46e5">Verify Your Email</h1>
<p>Hi %s,</p>
<p>Please verify your email address by clicking the button below.</p>
<a href="%s" style="display:inline-block;background:#4f46e5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;margin-top:16px">Verify Email</a>
<p style="color:#71717a;font-size:14px;margin-top:24px">This link expires in 24 hours.</p>
</div>`, name, verifyURL)
	return s.send(to, subject, html)
}

// SendPasswordReset sends a password reset email.
func (s *Service) SendPasswordReset(to, name, token string) error {
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.frontendURL, token)
	subject := "Reset your SkillSwap password"
	html := fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:0 auto">
<h1 style="color:#4f46e5">Reset Your Password</h1>
<p>Hi %s,</p>
<p>We received a request to reset your password. Click the button below to create a new password.</p>
<a href="%s" style="display:inline-block;background:#4f46e5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;margin-top:16px">Reset Password</a>
<p style="color:#71717a;font-size:14px;margin-top:24px">This link expires in 1 hour. If you didn't request this, you can safely ignore this email.</p>
</div>`, name, resetURL)
	return s.send(to, subject, html)
}

// SendSwapRequest notifies a user about a new swap request.
func (s *Service) SendSwapRequest(to, name, requesterName, skillName string) error {
	subject := "New skill swap request from " + requesterName
	html := fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:0 auto">
<h1 style="color:#4f46e5">New Swap Request</h1>
<p>Hi %s,</p>
<p><strong>%s</strong> wants to swap skills with you for <strong>%s</strong>.</p>
<a href="%s/swaps" style="display:inline-block;background:#4f46e5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;margin-top:16px">View Request</a>
</div>`, name, requesterName, skillName, s.frontendURL)
	return s.send(to, subject, html)
}

// SendSwapStatusChange notifies a user about a swap status change.
func (s *Service) SendSwapStatusChange(to, name, otherName, status string) error {
	subject := fmt.Sprintf("Swap %s — SkillSwap", status)
	html := fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:0 auto">
<h1 style="color:#4f46e5">Swap %s</h1>
<p>Hi %s,</p>
<p>Your skill swap with <strong>%s</strong> has been <strong>%s</strong>.</p>
<a href="%s/swaps" style="display:inline-block;background:#4f46e5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;margin-top:16px">View Swaps</a>
</div>`, status, name, otherName, status, s.frontendURL)
	return s.send(to, subject, html)
}

// SendNewMessage notifies an offline user about a new message.
func (s *Service) SendNewMessage(to, name, senderName string) error {
	subject := "New message from " + senderName
	html := fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:0 auto">
<h1 style="color:#4f46e5">New Message</h1>
<p>Hi %s,</p>
<p>You have a new message from <strong>%s</strong>.</p>
<a href="%s/messages" style="display:inline-block;background:#4f46e5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;margin-top:16px">Read Message</a>
</div>`, name, senderName, s.frontendURL)
	return s.send(to, subject, html)
}
