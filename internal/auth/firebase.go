package auth

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go"
	auth "firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

// FirebaseAuthenticator is a authenticator for Firebase.
type FirebaseAuthenticator struct {
	client *auth.Client
}

func NewFirebaseAuthenticator(cred []byte) (*FirebaseAuthenticator, error) {
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsJSON(cred))
	if err != nil {
		return nil, fmt.Errorf("firebase app init failed: %w", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("firebase auth init failed: %w", err)
	}

	return &FirebaseAuthenticator{client}, nil
}

// Authenticate returns a new context with the user ID if the session is valid.
func (f *FirebaseAuthenticator) Authenticate(ctx context.Context, session string) (context.Context, error) {
	tk, err := f.client.VerifySessionCookieAndCheckRevoked(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("verify session cookie: %w", err)
	}

	return SetUserID(ctx, UserID(tk.UID)), nil
}

// VerifyIDToken returns a new session if the ID token is valid.
func (f *FirebaseAuthenticator) SessionCookie(ctx context.Context, idToken string) (string, error) {
	return f.client.SessionCookie(ctx, idToken, time.Hour*24*14)
}

// DeleteUser deletes the user from Firebase.
func (f *FirebaseAuthenticator) DeleteUser(ctx context.Context, uid UserID) error {
	return f.client.DeleteUser(ctx, string(uid))
}

// GetUser returns the user from Firebase.
func (f *FirebaseAuthenticator) GetUser(ctx context.Context, uid UserID) (*auth.UserRecord, error) {
	return f.client.GetUser(ctx, string(uid))
}
