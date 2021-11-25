package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var random string = "random"

var config = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

func main() {
	godotenv.Load(".env")
	config.ClientID = os.Getenv("CLIENT")
	config.ClientSecret = os.Getenv("SECRET")
	http.HandleFunc("/", Home)
	http.HandleFunc("/google", Login)
	http.HandleFunc("/callback", CallBack)
	fmt.Println("Starting server at port :8080.")
	log.Fatal(http.ListenAndServe(":8080", nil).Error())
}

func Home(response http.ResponseWriter, request *http.Request) {
	temp := template.Must(template.ParseFiles("index.html"))
	if err := temp.Execute(response, nil); err != nil {
		log.Fatalln(err.Error())
	}
}

func Login(response http.ResponseWriter, request *http.Request) {
	url := config.AuthCodeURL(random)
	http.Redirect(response, request, url, http.StatusTemporaryRedirect)
}

func CallBack(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	if request.Form.Get("state") != random {
		log.Println("Invalid state.")
		http.Redirect(response, request, "/", http.StatusTemporaryRedirect)
		return
	}
	token, err := config.Exchange(context.Background(), request.Form.Get("code"))
	if err != nil {
		log.Println("Token state.")
		http.Redirect(response, request, "/", http.StatusTemporaryRedirect)
		return
	}
	resp, _ := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	var data map[string]interface{}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err.Error())
	}
	json.Unmarshal(bytes, &data)
	entries := make([]map[string]interface{}, 0)
	previous, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Fatalln(err.Error())
	}
	json.Unmarshal(previous, &entries)
	entries = append(entries, data) 
	file, err := os.Create("data.json") 
	if err != nil {
		log.Fatalln(err.Error())
	}
	json.NewEncoder(file).Encode(entries)
	fmt.Fprintln(response, "Success!")
}
