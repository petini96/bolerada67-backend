package csvimporter

import (
	"errors"
	"log"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Personalization struct {
	gorm.Model
	Name          string `json:"name"`
	Number        string `json:"number"`
	IsAllSponsors bool   `json:"is_all_sponsors"`
	HasPatch      bool   `json:"has_patch"`
}

// func GetShorts(db *gorm.DB) ([]Product, error) {

// }
func (p PersonalizationDB) GetAll() ([]Product, error) {
	var products []Product
	p.DB.Joins("Category").Joins("Type").Find(&products)
	//err := db.Model(&Product{}).Preload("Category").Find(&products).Error
	// log.Fatal(products[0].Category)

	// log.Fatal(err, products)
	return products, nil
}

type PersonalizationDB struct {
	*gorm.DB
	*Product
}

func NewPersonalizationDB(db *gorm.DB) *PersonalizationDB {
	return &PersonalizationDB{
		DB: db,
	}
}

func ParseToBool(str string) bool {
	if str == "VERDADEIRO" {
		return true
	}
	return false
}

func (p PersonalizationDB) CreateDB(srv *sheets.Service) error {
	p.DB.AutoMigrate(&Personalization{})

	spreadsheetId := "1xAcWfbWLulOTIcPHxvKsBlLeIaQam5OQf7BMMzhm71Y"
	readRange := "personalizacao!A3:E"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {
		for _, row := range resp.Values {

			isAllSponsors := ParseToBool(row[2].(string))
			hasPatch := ParseToBool(row[3].(string))

			modelPersonalization := &Personalization{
				Name:          row[0].(string),
				Number:        row[1].(string),
				IsAllSponsors: isAllSponsors,
				HasPatch:      hasPatch,
			}

			p.DB.Create(modelPersonalization)
		}

	}
	return nil

}
