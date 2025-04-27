package email

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/resend/resend-go/v2"
)

func SendEndEmail(to string, name string, email string, token string) (*resend.SendEmailResponse, error) {
	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
    emilerHost := os.Getenv("EMAILER_HOST")

    // Make request to verify token
    resp, err := http.Post(emilerHost, "application/json", bytes.NewBuffer([]byte(`{"token":"`+token+`"}`)))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    params := &resend.SendEmailRequest{
        From:    "ctf@autonoma.app",
        To:      []string{to},
        Subject: "Autonoma CTF",
        Html:    string(body),
    }

    sent, err := client.Emails.Send(params)

    if err != nil {
        return nil, err
    }

    return sent, nil
}