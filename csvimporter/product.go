package csvimporter

import (
	"errors"
	"log"
	"strconv"

	"github.com/petini96/go-sheets/helper"
	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Uniform    string   `json:"uniform"`
	ModelName  string   `json:"model_name"`
	Photo      string   `json:"photo"`
	Serie      string   `json:"serie"`
	Gender     string   `json:"gender"`
	TypeID     int      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"type_id"`
	Type       Type     `json:"type"`
	CategoryID int      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"category_id"`
	Category   Category `json:"category"`
}

// func GetShorts(db *gorm.DB) ([]Product, error) {

// }
func (p ProductDB) GetAll() ([]Product, error) {
	var products []Product
	p.DB.Joins("Category").Joins("Type").Find(&products)
	//err := db.Model(&Product{}).Preload("Category").Find(&products).Error
	// log.Fatal(products[0].Category)

	// log.Fatal(err, products)
	return products, nil
}

type ProductDB struct {
	*gorm.DB
	*Product
}

func NewProductDB(db *gorm.DB) *ProductDB {
	return &ProductDB{
		DB: db,
	}
}

func (p ProductDB) CreateDB(srv *sheets.Service) error {
	p.DB.AutoMigrate(&Product{})

	spreadsheetId := "1_zvgbpYAljsiIxCCIkBDGsFHj12j9TKbl5I4O94y-38"
	readRange := "produtos!A3:H"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {
		for _, row := range resp.Values {
			_, err := strconv.ParseInt(row[0].(string), 10, 64)
			typeID, err := strconv.ParseInt(row[6].(string), 10, 64)
			categoryID, err := strconv.ParseInt(row[7].(string), 10, 64)
			if err != nil {
				return err
			}

			modelProduct := &Product{
				Uniform:    row[1].(string),
				ModelName:  row[2].(string),
				Photo:      helper.EmbedPhoto(row[3].(string)),
				Serie:      row[4].(string),
				Gender:     row[5].(string),
				TypeID:     int(typeID),
				CategoryID: int(categoryID),
			}

			p.DB.Create(modelProduct)
		}

	}
	return nil

}
