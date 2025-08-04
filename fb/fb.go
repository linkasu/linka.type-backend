package fb

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	dotenv "github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var fb *firebase.App

func init() {
	dotenv.Load()
	var err error
	conf := &firebase.Config{
		DatabaseURL: "https://distypepro-android.firebaseio.com",
	}
	opt := option.WithCredentialsFile("./firebase.json")
	fb, err = firebase.NewApp(context.Background(), conf, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
}
