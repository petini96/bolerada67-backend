package csvimporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
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
			"https://www.googleapis.com/auth/drive",
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

func PermissionGoogleDrive(ctx *gin.Context, id string) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func DriveProducts(ctx *gin.Context, srv *drive.Service) []DrivePhotoResponse {
	pageSize := ctx.Query("pageSize")
	//pageToken := ctx.Query("pageToken")
	// Carrega o arquivo JSON das credenciais de serviço
	// credentials, err := ioutil.ReadFile("./credentials.json")
	// if err != nil {
	// 	panic(err)
	// }

	// // Cria o cliente da API do Google Drive

	// config, err := google.ConfigFromJSON(credentials, drive.DriveReadonlyScope)
	// if err != nil {
	// 	panic(err)
	// }
	// client := getClient(config)

	// srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve Sheets client: %v", err)
	// }

	// ID da pasta compartilhada
	// c.Query("id")
	folderID := ctx.Param("id")
	// Listar os arquivos na pasta compartilhada
	// files, err := srv.Files.List().
	// 	Q(fmt.Sprintf("'%s' in parents", folderID)).
	// 	Fields("nextPageToken, files(id, name, webViewLink, webContentLink)").
	// 	Do()
	// if err != nil {
	// 	log.Fatalf("Não foi possível listar os arquivos: %v", err)
	// }

	// // Iterar sobre os arquivos e exibir as URLs das imagens
	// for _, file := range files.Files {
	// 	if file.MimeType == "image/jpeg" || file.MimeType == "image/png" {
	// 		url := file.WebContentLink
	// 		// Exibir a URL da imagem em uma tag de imagem em seu site
	// 		log.Fatal("<img src='%s' alt='%s' />\n", url, file.Name)
	// 	}
	// }
	response := []DrivePhotoResponse{}
	//var files *drive.FileList
	var err error
	num, err := strconv.ParseInt(pageSize, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	pageToken := ctx.DefaultQuery("pageToken", "")
	q := srv.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).PageSize(num).PageToken(pageToken)
	r, err := q.Do()
	if err != nil {
		fmt.Println("moio")
	}

	for _, f := range r.Files {
		response = append(response, DrivePhotoResponse{
			ID:   f.Id,
			Name: f.Name,
		})
	}
	// Check if there are more results
	if r.NextPageToken != "" {
		ctx.Header("nextPageToken", r.NextPageToken)
	}
	// for {
	// 	if pageToken == "" {
	// 		files, err = srv.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).PageSize(num).Fields("files(id, name)").Do()
	// 		if err != nil {
	// 			log.Fatalf("Falha ao listar os arquivos na pasta compartilhada: %v", err)
	// 		}

	// 	} else {
	// 		files, err = srv.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).Fields("files(id, name)").Do()
	// 		if err != nil {
	// 			log.Fatalf("Falha ao listar os arquivos na pasta compartilhada: %v", err)
	// 		}
	// 	}

	// 	// Imprimir os nomes e IDs dos arquivos na pasta compartilhada
	// 	for _, f := range files.Files {

	// 		// ID do arquivo da imagem
	// 		// fileID := f.Id

	// 		response = append(response, DrivePhotoResponse{
	// 			ID:   f.Id,
	// 			Name: f.Name,
	// 		})
	// 		// // Acessar as informações do arquivo
	// 		// file, err := srv.Files.Get(fileID).Fields("webViewLink, webContentLink").Do()
	// 		// if err != nil {
	// 		// 	log.Fatalf("Não foi possível obter as informações do arquivo: %v", err)
	// 		// }

	// 		// // Exibir a imagem em uma tag de imagem em seu site
	// 		// fmt.Printf("<img src='%s' alt='imagem' />\n", file.WebViewLink)

	// 		// url := f.WebContentLink
	// 		// fmt.Printf("Nome: %s, ID: %s\n", f.Name, f.Id, url)
	// 	}
	// 	// verifica se há mais páginas
	// 	if files.NextPageToken == "" {
	// 		break
	// 	}

	// 	// configura o token de página para obter a próxima página
	// 	pageToken = files.NextPageToken
	// }
	return response

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
