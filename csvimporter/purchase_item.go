package csvimporter

import (
	"errors"
	"log"
	"strconv"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type PurchaseItem struct {
	gorm.Model

	ProductID int     `gorm:"constraint:OnUpdate:CASCADE,Ondelete:SET NULL;" json:"product_id"`
	Product   Product `json:"product"`

	PurchaseID int      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"purchase_id"`
	Purchase   Purchase `json:"purchase"`

	Quantity int `json:"quantity"`

	PersonalizationID *int            `gorm:"null,constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"personalization_id"`
	Personalization   Personalization `json:"personalization"`

	Size string `json:"size"`
	Bug  string `json:"bug"`
}

func (p PurchaseItemDB) GetAll() ([]Purchase, error) {

	var purchases []Purchase
	err := p.DB.Model(&Client{}).Preload("Purchase").Find(&purchases).Error
	return purchases, err
}

type PurchaseItemDB struct {
	*gorm.DB
}

//creates a new category db
func NewPurchaseItemDB(db *gorm.DB) *PurchaseItemDB {
	return &PurchaseItemDB{
		DB: db,
	}
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (p *PurchaseItemDB) CreateDB(srv *sheets.Service) error {
	p.DB.AutoMigrate(&PurchaseItem{})

	spreadsheetId := "1xAcWfbWLulOTIcPHxvKsBlLeIaQam5OQf7BMMzhm71Y"
	readRange := "itens de compra!A3:G"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {

		for _, row := range resp.Values {
			productID, err := strconv.ParseInt(row[1].(string), 10, 64)
			if err != nil {
				return err
			}

			purchaseID, err := strconv.ParseInt(row[2].(string), 10, 64)
			if err != nil {
				return err
			}

			quantity, err := strconv.Atoi(row[3].(string))
			if err != nil {
				log.Fatal(err)
			}

			var finalPersonalization *int

			if row[4].(string) != "0" {
				personalizationID, _ := strconv.ParseInt(row[4].(string), 10, 64)
				inteiro := int(personalizationID)
				finalPersonalization = &inteiro
			}

			purchaseItem := &PurchaseItem{
				ProductID:         int(productID),
				PurchaseID:        int(purchaseID),
				Quantity:          quantity,
				PersonalizationID: finalPersonalization,
				Size:              row[5].(string),
				Bug:               row[6].(string),
			}

			p.DB.Create(purchaseItem)
		}
	}

	return nil

}
