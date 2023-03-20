package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/petini96/go-sheets/csvimporter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
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
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

//Client method to import data.
func Import(db *gorm.DB) {

}
func Drive(ctx *gin.Context) {

	// Configurar as credenciais de autenticação
	creds, err := google.FindDefaultCredentials(ctx, drive.DriveScope)
	if err != nil {
		log.Fatalf("Falha ao encontrar credenciais padrão: %v", err)
	}

	// Configurar o cliente da API do Google Drive
	srv, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("Falha ao criar o cliente da API do Google Drive: %v", err)
	}

	// ID da pasta compartilhada
	folderID := "1WURcRzgC8mktrb4elvrmA-jtrjkxKCBj"

	// Listar os arquivos na pasta compartilhada
	files, err := srv.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).Fields("files(id, name)").Do()
	if err != nil {
		log.Fatalf("Falha ao listar os arquivos na pasta compartilhada: %v", err)
	}

	// Imprimir os nomes e IDs dos arquivos na pasta compartilhada
	for _, f := range files.Files {
		log.Fatal(fmt.Printf("Nome: %s, ID: %s\n", f.Name, f.Id))
	}
}
func Download(imageUrl string) {
	// Faça o download da imagem usando o Go
	imageResponse, err := http.Get(imageUrl)
	if err != nil {
		fmt.Println("Erro ao fazer o download da imagem:", err)
		return
	}
	defer imageResponse.Body.Close()
	// uuid := uuid.New().String()
	// Crie um arquivo para salvar a imagem
	file, err := os.Create("./assets/image.jpg")
	if err != nil {
		fmt.Println("Erro ao criar o arquivo:", err)
		return
	}
	defer file.Close()

	// Salve a imagem no arquivo
	_, err = io.Copy(file, imageResponse.Body)
	if err != nil {
		fmt.Println("Erro ao salvar a imagem no arquivo:", err)
		return
	}
}
func Yuppo() []string {
	// URL da página que contém as imagens
	url := "https://minkang.x.yupoo.com/categories/3855959"

	// Faz a solicitação HTTP para a página
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Analisa a página HTML usando a biblioteca goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var links []string
	// Seleciona todos os elementos HTML com a classe "thumb"
	doc.Find(".album__absolute.album__img.autocover").Each(func(i int, s *goquery.Selection) {

		// Extrai o valor do atributo "src" de cada elemento
		imgSrc, exists := s.Attr("data-src")
		if exists {
			link := fmt.Sprintf("https:" + imgSrc)
			Download(link)
			links = append(links, link)
			// Imprime o URL completo da imagem

		}
	})
	return links
}
func main() {
	var srv *drive.Service

	dsn := "host=172.21.16.1 user=postgres password=secret dbname=bolerada67 port=5432 sslmode=disable TimeZone=America/Sao_Paulo"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("wow man! It's a little pork...")
	}

	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/start", func(c *gin.Context) {
		ctx := context.Background()
		srv, err = drive.NewService(ctx, option.WithAPIKey("AIzaSyCvj50k6p4pvPF5YmADMkZwkAkV3r1r8Kk"))
		if err != nil {
			log.Fatalf("Unable to retrieve Sheets client: %v", err)
		}
	})
	r.GET("/products/folder/:id", func(c *gin.Context) {
		// c.Writer.Header().Set("Nextpagetoken", "fs")

		c.Header("Access-Control-Expose-Headers", "Nextpagetoken")

		c.JSON(http.StatusOK, gin.H{
			"message":  "pong",
			"products": csvimporter.DriveProducts(c, srv),
		})

		// csvimporter.PermissionGoogle(c)
	})
	r.GET("xaxaxa", func(c *gin.Context) {

		csvimporter.PermissionGoogleDrive(c, c.Query("id"))
		c.JSON(http.StatusOK, gin.H{
			"message":  "pong",
			"products": Yuppo(),
		})
	})
	r.GET("/import", func(c *gin.Context) {
		csvimporter.PermissionGoogle(c)

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.JSON(http.StatusOK, gin.H{
			"message": "imported!",
		})
	})
	r.GET("/products-yuppo", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"message":  "pong",
			"products": Yuppo(),
		})
	})

	r.GET("/storage", func(c *gin.Context) {
		// URL da página que contém as imagens
		url := "https://minkang.x.yupoo.com/categories/3855959"

		// Faz a solicitação HTTP para a página
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// Analisa a página HTML usando a biblioteca goquery
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var links []string
		// Seleciona todos os elementos HTML com a classe "thumb"
		doc.Find(".album__absolute.album__img.autocover").Each(func(i int, s *goquery.Selection) {

			// Extrai o valor do atributo "src" de cada elemento
			imgSrc, exists := s.Attr("data-src")
			if exists {
				link := fmt.Sprintf("https:" + imgSrc)
				links = append(links, link)
				// Imprime o URL completo da imagem

			}
		})
		resp, err = http.Get(links[0])
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer resp.Body.Close()

		// Escreva a imagem no corpo da resposta
		imgBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, "image/jpeg", imgBytes)
		// c.File("path/to/image.jpg")
	})

	r.GET("/products", func(c *gin.Context) {
		productDB := csvimporter.NewProductDB(db)

		products, err := productDB.GetAll()
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "pong",
			"products": products,
		})
	})
	r.GET("/clients", func(c *gin.Context) {
		clientDB := csvimporter.NewClientDB(db)

		clients, err := clientDB.GetAll()
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"clients": clients,
		})
	})
	r.GET("/debts", func(c *gin.Context) {
		debtDB := csvimporter.NewDebtDB(db)

		debts, err := debtDB.GetAll()
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"debts":   debts,
		})
	})
	r.GET("/oauth2/callback", func(c *gin.Context) {
		err := csvimporter.Callback(c, db)
		if err != nil {
			log.Fatal(err)
		}
	})

	r.GET("/suppliers", func(c *gin.Context) {
		suppliersDB := csvimporter.NewSupplierDB(db)
		suppliers, err := suppliersDB.GetAll()
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "pong",
			"suppliers": suppliers,
		})
	})
	r.Run("0.0.0.0:5000")
}
