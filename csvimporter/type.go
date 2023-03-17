package csvimporter

import (
	"errors"
	"log"
	"strconv"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Type struct {
	gorm.Model
	SimpleTable `gorm:"embedded"`
}

type TypeDB struct {
	*gorm.DB
}

// type Category struct {
// 	gorm.Model
// 	SimpleTable
// }
//creates a new category db
func NewTypeDB(db *gorm.DB) *TypeDB {
	return &TypeDB{
		DB: db,
	}
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (t *TypeDB) CreateDB(srv *sheets.Service) error {
	t.DB.AutoMigrate(&Type{})

	spreadsheetId := "1_zvgbpYAljsiIxCCIkBDGsFHj12j9TKbl5I4O94y-38"
	readRange := "tipos!A2:C"

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

			modelType := &Type{
				SimpleTable: SimpleTable{
					Name:        row[1].(string),
					Description: row[2].(string),
				},
			}

			t.DB.Create(modelType)
		}
	}

	return nil

}
