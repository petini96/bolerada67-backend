package csvimporter

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Debt struct {
	gorm.Model
	Data                time.Time `json:"date"`
	TotalValue          float64   `json:"total_value"`
	QuantityInstallment int       `json:"quantity_installment"`
	IsDone              bool      `json:"is_done"`
	ClientID            int       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"client_id"`
	Installments        []Installment
}

func (d DebtDB) GetAll() ([]Client, error) {

	var clients []Client
	err := d.DB.Model(&Client{}).Preload("Debts").Find(&clients).Error
	return clients, err

	// var clients []Client
	// d.DB.Joins("JOIN debts deb ON deb.client_id = deb.id").Find(&clients)

	//var clients []Client
	//d.Model(&Client{}).Association("debt").Find(&clients)
	//d.DB.Joins("Debt").Find(&clients)
	//d.DB.Model(&Debt{}).Preload("Client").Find(&debts)

	//d.DB.First(&client, 2)
	// qtd := d.DB.Model(&client).Association("Debt").Count()

	// d.DB.Model(&debts).Association("Client").find()
	// err := d.DB.Model(&Debt{}).Preload("Client").Find(&debts).Error
	return clients, nil

	// log.Fatal(qtd)
	// return client, nil
	// // d.DB.First(&client)

	// fmt.Println(qtd)
	// return client, nil
}

type DebtDB struct {
	*gorm.DB
}

//creates a new category db
func NewDebtDB(db *gorm.DB) *DebtDB {
	return &DebtDB{
		DB: db,
	}
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (d *DebtDB) CreateDB(srv *sheets.Service) error {
	d.DB.AutoMigrate(&Debt{})

	spreadsheetId := "1xAcWfbWLulOTIcPHxvKsBlLeIaQam5OQf7BMMzhm71Y"
	readRange := "parcelamentos!A3:F"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {

		for _, row := range resp.Values {

			formato := `02/01/2006` // Este é o DD/MM/YYYY
			data, err := time.Parse(formato, row[1].(string))
			if err != nil {
				log.Fatal(err)
			}

			totalValue, err := strconv.ParseFloat(strings.Replace(row[2].(string), ",", ".", -1), 64)
			if err != nil {
				log.Fatal(err)
			}

			quantityInstallment, err := strconv.Atoi(row[3].(string))
			if err != nil {
				log.Fatal(err)
			}

			isDone := false
			if row[4].(string) == "sim" {
				isDone = true
			}

			clientID, err := strconv.Atoi(row[5].(string))
			if err != nil {
				log.Fatal(err)
			}

			modelDebt := &Debt{
				Data:                data,
				TotalValue:          totalValue,
				QuantityInstallment: quantityInstallment,
				IsDone:              isDone,
				ClientID:            clientID,
			}

			d.DB.Create(modelDebt)
		}
	}

	return nil

}
