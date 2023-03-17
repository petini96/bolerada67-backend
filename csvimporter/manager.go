package csvimporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func init() {

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:5000/oauth2/callback",
		ClientID:     "474096237238-tdag3gfl9l2tvqjlje5mdb5j6ssmkk10.apps.googleusercontent.com",
		ClientSecret: "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",
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets.readonly",
			// "https://www.googleapis.com/auth/userinfo.profile",
			// "https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

}

var (
	googleOauthConfig *oauth2.Config
	// TODO: randomize it
	oauthStateString = "pseudo-random"
	adminCallback    string
	clientCallback   string
)

type loginOauth2UserResponse struct {
	AccessToken string `json:"access_token"`
}
type UserGoogleResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	Name          string `json:"name"`
	Locale        string `json:"locale"`
}

//USER INFO
func getUserInfo(state string, code string) (UserGoogleResponse, error) {
	if state != oauthStateString {
		return UserGoogleResponse{}, fmt.Errorf("invalid oauth state")
	}
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return UserGoogleResponse{}, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return UserGoogleResponse{}, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return UserGoogleResponse{}, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	var userGoogle UserGoogleResponse
	if err := json.Unmarshal(contents, &userGoogle); err != nil {
		panic(err)
	}
	return userGoogle, nil
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

type DBCreator interface {
	CreateDB(srv *sheets.Service) error
}

func PermissionGoogle(ctx *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func Import(entityInterface DBCreator, client *http.Client) error {

	ctx := context.Background()
	// b, err := os.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// // If modifying these scopes, delete your previously saved token.json.
	// config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	// if err != nil {
	// 	log.Fatalf("Unable to parse client secret file to config: %v", err)
	// }

	// client := getClient(config)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	err = entityInterface.CreateDB(srv)
	if err != nil {
		return err
	}
	return nil
}
