package csvimporter

import (
	"errors"
	"log"
	"strconv"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	SimpleTable
}

type CategoryDB struct {
	*gorm.DB
}

//creates a new category db
func NewCategoryDB(db *gorm.DB) *CategoryDB {
	return &CategoryDB{
		DB: db,
	}
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (c *CategoryDB) CreateDB(srv *sheets.Service) error {
	c.DB.AutoMigrate(&Category{})

	spreadsheetId := "1_zvgbpYAljsiIxCCIkBDGsFHj12j9TKbl5I4O94y-38"
	readRange := "categorias!A2:C"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {
		for _, row := range resp.Values {
			_, err := strconv.ParseInt(row[0].(string), 10, 64) //id value
			if err != nil {
				return err
			}

			modelCategory := &Category{
				SimpleTable: SimpleTable{
					Name:        row[1].(string),
					Description: row[2].(string),
				},
			}

			c.DB.Create(modelCategory)
		}
	}

	return nil

}
