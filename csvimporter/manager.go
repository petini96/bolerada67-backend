package csvimporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"

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

//Get's the user info from google.
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

//Makes the directory according with default.
func createDIR(imageDir string, searchFolder string) error {
	imageDir = getUploadFilesDIR(searchFolder)
	return os.MkdirAll(imageDir, os.ModePerm)
}

// Creates a directory to store images if it doesn't exist.
func getUploadFilesDIR(searchFolder string) string {
	return "./images/" + searchFolder
}

//Gets the default name to file creation.
func getCreateFilePath(imageDir string, imageID string) string {
	return fmt.Sprintf("%s/%s.jpg", imageDir, imageID)
}

// Create a file to save the image
func createOSFile(imagePath string) (*os.File, error) {
	return os.Create(imagePath)
}

// Fetch the image content from Google Drive
func fetchDriveFileByID(srv *drive.Service, imageID string) (*http.Response, error) {
	return srv.Files.Get(imageID).Download()
}

// Copy the image content to the file on your server
func copyDriveImageToFile(file *os.File, imageResp *http.Response) (int64, error) {
	return io.Copy(file, imageResp.Body)
}

// Return the URL of the saved image
func getDriveWebImageURL(searchFolder string, imageID string) string {
	return "http://localhost:5000/images/" + searchFolder + "/" + imageID + ".jpg"
}

//Saves the image and get the URL according to the file attributes
func SaveImageAndGetURL(searchFolder string, imageID string, imageName string, srv *drive.Service) (string, error) {
	imageDir := getUploadFilesDIR(searchFolder)
	if err := createDIR(imageDir, searchFolder); err != nil {
		return "", err
	}
	imagePath := getCreateFilePath(imageDir, imageID)
	file, err := createOSFile(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	imageResp, err := fetchDriveFileByID(srv, imageID)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	_, err = copyDriveImageToFile(file, imageResp)
	if err != nil {
		return "", err
	}

	imageURL := getDriveWebImageURL(searchFolder, imageID)
	return imageURL, nil
}

//Gets the parameters from context.
func getPageSizeAndFolderIDFromRequest(ctx *gin.Context) (string, string) {
	pageSize := ctx.Query("pageSize")
	searchFolder := ctx.Param("id")
	return pageSize, searchFolder
}

//Gets the keys from redis.
func getKeysFromRedis(ctx *gin.Context, redisClient *redis.Client, searchFolder string) ([]string, error) {
	return redisClient.Keys(ctx, searchFolder+"*").Result()
}

//Converts pageSize from string to int64.
func pageSizeToInt64(pageSize string) (int64, error) {
	return strconv.ParseInt(pageSize, 10, 64)
}

// Verifies if the key is greather than zero.
func hasRedisKeys(keys []string) bool {
	if len(keys) <= 0 {
		return false
	}
	return true
}

//Defines the default pageToken.
func setEmptyPageToken(ctx *gin.Context) string {
	return ctx.DefaultQuery("pageToken", "")
}

//Perform file query in google api.
//Defines the folder ID and defines the total number of items to return.
func getFilesFromDrive(srv *drive.Service, searchFolder string, pageToken string, num int64) (*drive.FileList, error) {
	q := srv.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed = false", searchFolder)).PageSize(num).PageToken(pageToken)
	return q.Do()
}

//Add a response object with the file information
func addDrivePhotoResponse(f *drive.File, imageURL string, response []DrivePhotoResponse) []DrivePhotoResponse {
	result := append(response, DrivePhotoResponse{
		ID:   f.Id,
		Name: f.Name,
		URL:  imageURL,
	})
	return result
}

// adiciona na resposta final o item do cache
func addDrivePhotoResponseByObj(response []DrivePhotoResponse, drivePhotoResponse DrivePhotoResponse) []DrivePhotoResponse {
	result := append(response, drivePhotoResponse)
	return result
}

// faz serialização json do objeto de response criado
func marshalFile(f *drive.File, imageURL string) ([]byte, error) {
	return json.Marshal(DrivePhotoResponse{
		ID:   f.Id,
		Name: f.Name,
		URL:  imageURL,
	})
}

// busca pelo nome(concatenação pastaGoogle + file.ID) padrão
func getCacheKey(f *drive.File, searchFolder string) string {
	return fmt.Sprintf("%s_%s", searchFolder, f.Id)
}

// faz criação de chave no redis com valor
func setCacheKey(redisClient *redis.Client, cacheKey string, ctx *gin.Context, responseJSON string) error {
	return redisClient.Set(ctx, cacheKey, responseJSON, 0).Err()
}

// verifica se pageToken é nulo, caso-contrário adiciona o campo ao header
func nextPageTokenIsEmpty(r *drive.FileList) bool {
	return r.NextPageToken != ""
}

func getNextPageToken(r *drive.FileList) string {
	return r.NextPageToken
}

//busca por chave no redis
func getKeyFromRedis(redisClient *redis.Client, ctx *gin.Context, key string) ([]byte, error) {
	return redisClient.Get(ctx, key).Bytes()
}

//faz unmarshalling de JSON para o objeto de resposta
func unmarshalDrivePhotoResponse(cachedImage []byte, cachedResponse *DrivePhotoResponse) error {
	return json.Unmarshal(cachedImage, &cachedResponse)
}

//Return products obtained from cache or drive api paginated. The process download the file and put into cache.
func DriveProducts(ctx *gin.Context, srv *drive.Service, redisClient *redis.Client) []DrivePhotoResponse {
	var err error
	response := []DrivePhotoResponse{}
	var nexPageTokenStr string

	pageSize, searchFolder := getPageSizeAndFolderIDFromRequest(ctx)

	keys, err := getKeysFromRedis(ctx, redisClient, searchFolder)
	if err != nil {
		log.Fatal("Error:", err)
	}
	if !hasRedisKeys(keys) {
		num, err := pageSizeToInt64(pageSize)
		if err != nil {
			log.Fatal(err)
		}

		pageToken := setEmptyPageToken(ctx)
		r, err := getFilesFromDrive(srv, searchFolder, pageToken, num)
		if err != nil {
			log.Fatal("fail my man...", err)
		}

		for _, f := range r.Files {
			imageURL, err := SaveImageAndGetURL(searchFolder, f.Id, f.Name, srv)
			if err != nil {
				log.Printf("Error saving image: %v", err)
			}
			response = addDrivePhotoResponse(f, imageURL, response)
			responseJSON, err := marshalFile(f, imageURL)
			if err != nil {
				log.Printf("Error to marshal to JSON: %v", err)
			}
			cacheKey := getCacheKey(f, searchFolder)
			err = setCacheKey(redisClient, cacheKey, ctx, string(responseJSON))
			if err != nil {
				log.Printf("Erro to remain on cache: %v", err)
			}
			fmt.Println("The image has been added to the cache")
		}
		if nextPageTokenIsEmpty(r) {
			nexPageTokenStr = r.NextPageToken
		}
		ctx.Header("nextPageToken", nexPageTokenStr)
		return response
	}

	for _, key := range keys {
		cachedImage, err := getKeyFromRedis(redisClient, ctx, key)
		if err == nil {
			fmt.Println("The image was found on cache!")
			var cachedResponse DrivePhotoResponse
			err := unmarshalDrivePhotoResponse(cachedImage, &cachedResponse)
			if err != nil {
				log.Printf("Error to unmarshall cache: %v", err)
			}
			response = addDrivePhotoResponseByObj(response, cachedResponse)
		} else { //not found on cache
			num, err := pageSizeToInt64(pageSize)
			if err != nil {
				log.Fatal(err)
			}

			pageToken := setEmptyPageToken(ctx)
			r, err := getFilesFromDrive(srv, searchFolder, pageToken, num)
			if err != nil {
				log.Fatal("fail guys. My blame.", err)
			}

			for _, f := range r.Files {
				imageURL, err := SaveImageAndGetURL(searchFolder, f.Id, f.Name, srv)
				if err != nil {
					log.Printf("Error saving image: %v", err)
				}
				response = addDrivePhotoResponse(f, imageURL, response)
				responseJSON, err := marshalFile(f, imageURL)
				if err != nil {
					log.Printf("Erro ao serializar resposta para JSON: %v", err)
				}
				cacheKey := getCacheKey(f, searchFolder)
				err = setCacheKey(redisClient, cacheKey, ctx, string(responseJSON))
				if err != nil {
					log.Printf("Erro ao armazenar resposta no cache: %v", err)
				}
				fmt.Println("Imagem carregada e armazenada em cache!")
			}
			if nextPageTokenIsEmpty(r) {
				nexPageTokenStr = getNextPageToken(r)
			}
			ctx.Header("nextPageToken", nexPageTokenStr)
			return response
		}
	}
	ctx.Header("nextPageToken", nexPageTokenStr)
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
