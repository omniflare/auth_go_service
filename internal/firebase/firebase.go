package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

func Init() {
	 ctx := context.Background()
    
    opt := option.WithCredentialsFile("serviceAccountKey.json")
    app, err := firebase.NewApp(ctx, nil, opt)
    if err != nil {
        log.Fatalf("error initializing firebase app: %v", err)
    }

	AuthClient, err = app.Auth(ctx)
    if err != nil {
        log.Fatalf("error getting Auth client: %v", err)
    }

    log.Println("Firebase initialized successfully")
}

func VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token , err := AuthClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	return token, nil 
}