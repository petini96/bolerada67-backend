package csvimporter

import (
	"errors"
	"log"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model
	Name          string `json:"name"`
	ContactNumber string `json:"contact_number"`
	Photo         string `json:"photo"`
	Debts         []Debt
}

func (c ClientDB) GetAll() ([]Client, error) {
	var clients []Client
	err := c.DB.Model(&Client{}).Preload("Debts").Find(&clients).Error
	return clients, err
}

type ClientDB struct {
	*gorm.DB
}

//creates a new category db
func NewClientDB(db *gorm.DB) *ClientDB {
	return &ClientDB{
		DB: db,
	}
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (c *ClientDB) CreateDB(srv *sheets.Service) error {
	c.DB.AutoMigrate(&Client{})

	spreadsheetId := "1_zvgbpYAljsiIxCCIkBDGsFHj12j9TKbl5I4O94y-38"
	readRange := "clientes!A3:D"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {

		for _, row := range resp.Values {
			strFinal := []string{"", "", "", ""}

			for index, value := range row {

				strFinal[index] = value.(string)

			}

			modelClient := &Client{
				Name:          strFinal[1],
				ContactNumber: strFinal[2],
				Photo:         strFinal[3],
			}

			c.DB.Create(modelClient)
		}
	}

	return nil

}
