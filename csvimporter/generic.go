package csvimporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type SimpleTable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DrivePhotoResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	URL  string `json:"url"`
}

func (d DrivePhotoResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *DrivePhotoResponse) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, d)
}

func Callback(ctx *gin.Context, db *gorm.DB) error {

	state := ctx.Request.FormValue("state")
	code := ctx.Request.FormValue("code")

	if state != oauthStateString {
		return errors.New("fail")
	}

	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return fmt.Errorf("code exchange failed: %s", err.Error())
	}

	client := googleOauthConfig.Client(ctx, token)

	categoryDB := NewCategoryDB(db)
	err = Import(categoryDB, client)
	if err != nil {
		fmt.Println(err)
	}

	typeDB := NewTypeDB(db)
	err = Import(typeDB, client)
	if err != nil {
		fmt.Println(err)
	}

	productDB := NewProductDB(db)
	err = Import(productDB, client)
	if err != nil {
		fmt.Println(err)
	}

	supplierDB := NewSupplierDB(db)
	err = Import(supplierDB, client)
	if err != nil {
		fmt.Println(err)
	}

	clientDB := NewClientDB(db)
	err = Import(clientDB, client)
	if err != nil {
		fmt.Println(err)
	}

	debtDB := NewDebtDB(db)
	err = Import(debtDB, client)
	if err != nil {
		fmt.Println(err)
	}

	installmentDB := NewInstallmentDB(db)
	err = Import(installmentDB, client)
	if err != nil {
		fmt.Println(err)
	}

	purchaseDB := NewPurchaseDB(db)
	err = Import(purchaseDB, client)
	if err != nil {
		fmt.Println(err)
	}

	personalizationDB := NewPersonalizationDB(db)
	err = Import(personalizationDB, client)
	if err != nil {
		fmt.Println(err)
	}

	purchaseItemDB := NewPurchaseItemDB(db)
	err = Import(purchaseItemDB, client)
	if err != nil {
		fmt.Println(err)
	}

	order := NewOrderDB(db)
	err = Import(order, client)
	if err != nil {
		fmt.Println(err)
	}
	// Configurar o cliente da API do Google Drive
	// srv, err := drive.NewService(ctx, option.WithCredentials(creds))
	// if err != nil {
	// 	log.Fatalf("Falha ao criar o cliente da API do Google Drive: %v", err)
	// }
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// ID da pasta compartilhada
	folderID := "1WURcRzgC8mktrb4elvrmA-jtrjkxKCBj"

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
	// Listar os arquivos na pasta compartilhada
	files, err := srv.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).Fields("files(id, name)").Do()
	if err != nil {
		log.Fatalf("Falha ao listar os arquivos na pasta compartilhada: %v", err)
	}

	// Imprimir os nomes e IDs dos arquivos na pasta compartilhada
	for _, f := range files.Files {

		// ID do arquivo da imagem
		// fileID := f.Id

		response = append(response, DrivePhotoResponse{
			ID:   f.Id,
			Name: f.Name,
			URL:  f.Name,
		})
		// // Acessar as informações do arquivo
		// file, err := srv.Files.Get(fileID).Fields("webViewLink, webContentLink").Do()
		// if err != nil {
		// 	log.Fatalf("Não foi possível obter as informações do arquivo: %v", err)
		// }

		// // Exibir a imagem em uma tag de imagem em seu site
		// fmt.Printf("<img src='%s' alt='imagem' />\n", file.WebViewLink)

		// url := f.WebContentLink
		// fmt.Printf("Nome: %s, ID: %s\n", f.Name, f.Id, url)
	}
	fmt.Print(response)
	return nil

}
