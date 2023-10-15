package csvimporter

import (
	"errors"
	"log"
	"strconv"
	"time"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	RequestDate  time.Time `json:"request_date"`
	DeliveryDate time.Time `json:"delivery_date"`
	Local        string    `json:"local"`
	ClientID     int       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"client_id"`
	Status       string    `json:"status"`
}

func (o OrderDB) GetAll() ([]Order, error) {

	var orders []Order
	err := o.DB.Model(&Client{}).Preload("Orders").Find(&orders).Error
	return orders, err
}

type OrderDB struct {
	*gorm.DB
}

//creates a new category db
func NewOrderDB(db *gorm.DB) *OrderDB {
	return &OrderDB{
		DB: db,
	}
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (o *OrderDB) CreateDB(srv *sheets.Service) error {
	o.DB.AutoMigrate(&Order{})

	spreadsheetId := "1xAcWfbWLulOTIcPHxvKsBlLeIaQam5OQf7BMMzhm71Y"
	readRange := "encomendas!A3:H"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {

		for _, row := range resp.Values {

			formato := `02/01/2006` // Este é o DD/MM/YYYY
			requestDate, err := time.Parse(formato, row[1].(string))
			if err != nil {
				log.Fatal(err)
			}

			var deliveryDate time.Time
			deliveryDate, err = time.Parse(formato, row[2].(string))
			if err != nil {
				deliveryDate = time.Time{}
			}

			local := row[3].(string)

			clientID, err := strconv.Atoi(row[4].(string))
			if err != nil {
				log.Fatal(err)
			}
			status := row[5].(string)

			modelOrder := &Order{
				RequestDate:  requestDate,
				DeliveryDate: deliveryDate,
				Local:        local,
				ClientID:     clientID,
				Status:       status,
			}

			o.DB.Create(modelOrder)
		}
	}

	return nil

}
