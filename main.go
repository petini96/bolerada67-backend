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
	"github.com/petini96/go-sheets/util"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//Open and create a file through own id. the path gived then uses the token and file created to encode.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

//Gets web token through own id. user agreements.
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

//Get token through own id. a file.
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

//Gets the http client thougt the config. Uses the json format.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

//Apply the cors policies.
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

//Generic function to conect a folder through own id.
func Drive(ctx *gin.Context) {
	creds, err := google.FindDefaultCredentials(ctx, drive.DriveScope)
	if err != nil {
		log.Fatalf("Falha ao encontrar credenciais padrão: %v", err)
	}

	srv, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("Falha ao criar o cliente da API do Google Drive: %v", err)
	}

	folderID := "1WURcRzgC8mktrb4elvrmA-jtrjkxKCBj" //id from folder shared

	files, err := srv.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).Fields("files(id, name)").Do()
	if err != nil {
		log.Fatalf("Falha ao listar os arquivos na pasta compartilhada: %v", err)
	}

	for _, f := range files.Files {
		log.Fatal(fmt.Printf("Nome: %s, ID: %s\n", f.Name, f.Id))
	}
}

//Download a image through a URL link.
func Download(imageUrl string) {
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

	_, err = io.Copy(file, imageResponse.Body)
	if err != nil {
		fmt.Println("Erro ao salvar a imagem no arquivo:", err)
		return
	}
}

//Do a scrapping fish on site yuppo.
func Yuppo() []string {
	url := "https://minkang.x.yupoo.com/categories/3855959" //page url

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body) //uses the goquery lib to do a Scraping Fish.
	if err != nil {
		log.Fatal(err)
	}
	var links []string

	doc.Find(".album__absolute.album__img.autocover").Each(func(i int, s *goquery.Selection) { // Seleciona all elements HTML
		imgSrc, exists := s.Attr("data-src") // get only src
		if exists {
			link := fmt.Sprintf("https:" + imgSrc)
			Download(link)
			links = append(links, link)
		}
	})
	return links
}

func main() {
	var srv *drive.Service

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		config.DbHost, config.DbUsername, config.DbPassword, config.DbDatabase, config.DbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Wow man! It's a little pork...")
	}

	redisAdd := fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort)
	cache := redis.NewClient(&redis.Options{
		Addr:     redisAdd,
		Password: config.RedisPassword,
		DB:       config.RedisDatabase,
	})

	r := gin.Default()
	r.Use(CORSMiddleware())

	//Do a simple cache test.
	r.GET("/tests/cache/:id", func(c *gin.Context) {
		ctx := context.Background()
		err = cache.Set(ctx, "key", "value", 0).Err()
		if err != nil {
			panic(err)
		}

		val, err := cache.Get(ctx, "key").Result()
		if err != nil {
			panic(err)
		}
		fmt.Println("key", val)

		val2, err := cache.Get(ctx, "key").Result()
		if err == redis.Nil {
			fmt.Println("key2 does not exist")
		} else if err != nil {
			panic(err)
		} else {
			fmt.Println("key2", val2)
		}
	})
	//Makes a test to test the connection with google.
	r.GET("/tests/drive/conection", func(c *gin.Context) {
		ctx := context.Background()
		srv, err = drive.NewService(ctx, option.WithAPIKey(config.ApiKey))
		if err != nil {
			log.Fatalf("Unable to retrieve Sheets client: %v", err)
		}
	})
	//Makes a download from google drive and save into cache.
	r.GET("/tests/drive/products/folder/:id", func(c *gin.Context) {
		c.Header("Access-Control-Expose-Headers", "Nextpagetoken")
		srv, err = drive.NewService(c, option.WithAPIKey("AIzaSyCvj50k6p4pvPF5YmADMkZwkAkV3r1r8Kk"))
		if err != nil {
			log.Fatalf("Unable to retrieve Sheets client: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "pong",
			"products": csvimporter.DriveProducts(c, srv, cache),
		})
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
		url := "https://minkang.x.yupoo.com/categories/3855959"
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var links []string
		doc.Find(".album__absolute.album__img.autocover").Each(func(i int, s *goquery.Selection) {
			imgSrc, exists := s.Attr("data-src")
			if exists {
				link := fmt.Sprintf("https:" + imgSrc)
				links = append(links, link)
			}
		})
		resp, err = http.Get(links[0])
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer resp.Body.Close()

		imgBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, "image/jpeg", imgBytes)
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
	serverHost := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
	r.Static("/images", "./images")
	r.Run(serverHost)
}
