package handler

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

// ResetNotifier achemine le lien de réinitialisation vers l'utilisateur.
// Le jeton ne doit JAMAIS être renvoyé dans la réponse HTTP.
type ResetNotifier interface {
	SendResetLink(email, link string) error
}

// LogNotifier écrit le lien dans les journaux du serveur. À réserver au
// développement : n'importe qui lisant les journaux pourrait prendre le compte.
type LogNotifier struct{}

func (LogNotifier) SendResetLink(email, link string) error {
	slog.Warn("[DEV] lien de réinitialisation (aucun SMTP configuré)", "email", email, "link", link)
	return nil
}

// SMTPNotifier envoie le lien par e-mail via net/smtp.
type SMTPNotifier struct {
	Host, Port, User, Pass, From string
}

func (n SMTPNotifier) SendResetLink(email, link string) error {
	// Les adresses proviennent du magasin d'utilisateurs, mais on refuse tout
	// retour chariot par précaution (injection d'en-têtes).
	if strings.ContainsAny(email+n.From, "\r\n") {
		return fmt.Errorf("adresse invalide")
	}
	body := "Bonjour,\r\n\r\n" +
		"Une réinitialisation de mot de passe a été demandée pour votre compte.\r\n" +
		"Ce lien est valable 15 minutes et ne peut être utilisé qu'une fois :\r\n\r\n" +
		link + "\r\n\r\n" +
		"Si vous n'êtes pas à l'origine de cette demande, ignorez ce message.\r\n"
	msg := "From: " + n.From + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: =?UTF-8?Q?R=C3=A9initialisation_de_votre_mot_de_passe?=\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body

	var a smtp.Auth
	if n.User != "" {
		a = smtp.PlainAuth("", n.User, n.Pass, n.Host)
	}
	return smtp.SendMail(n.Host+":"+n.Port, a, n.From, []string{email}, []byte(msg))
}
