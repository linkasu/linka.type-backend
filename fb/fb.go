package fb

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var fb *firebase.App

func init() {
	var err error
	conf := &firebase.Config{
		DatabaseURL: "https://distypepro-android.firebaseio.com",
	}
	//cheking if firebase.json is exists
	if _, err := os.Stat("/app/firebase.json"); os.IsNotExist(err) {
		log.Fatalf("firebase.json file does not exist")
	}
	opt := option.WithCredentialsFile("/app/firebase.json")
	fb, err = firebase.NewApp(context.Background(), conf, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	log.Println("Successfully initialized Firebase app")
}
