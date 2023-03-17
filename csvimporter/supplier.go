package csvimporter

import (
	"errors"
	"log"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Supplier struct {
	gorm.Model
	Company       string `json:"company"`
	ContactName   string `json:"contact_name"`
	CNPJ          string `json:"cnpj"`
	PaymentMethod string `json:"payment_method"`
	Pix           string `json:"pix"`
	ContactNumber string `json:"contact_number"`
	Photo         string `json:"photo"`
}

func (s SupplierDB) GetAll() ([]Supplier, error) {
	var suppliers []Supplier
	s.DB.Find(&suppliers)
	return suppliers, nil
}

type SupplierDB struct {
	*gorm.DB
}

//creates a new category db
func NewSupplierDB(db *gorm.DB) *SupplierDB {
	return &SupplierDB{
		DB: db,
	}
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (s *SupplierDB) CreateDB(srv *sheets.Service) error {
	s.DB.AutoMigrate(&Supplier{})

	spreadsheetId := "1_zvgbpYAljsiIxCCIkBDGsFHj12j9TKbl5I4O94y-38"
	readRange := "fornecedores!A3:H"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {

		for _, row := range resp.Values {

			modelSupplier := &Supplier{
				Company:       row[1].(string),
				ContactName:   row[2].(string),
				CNPJ:          row[3].(string),
				PaymentMethod: row[4].(string),
				Pix:           row[5].(string),
				ContactNumber: row[6].(string),
				Photo:         row[7].(string),
			}

			s.DB.Create(modelSupplier)
		}
	}

	return nil

}
