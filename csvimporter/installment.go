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

type Installment struct {
	gorm.Model
	Order           int       `json:"order"`
	Value           float64   `json:"value"`
	IsPaid          bool      `json:"is_paid"`
	PaymentForecast time.Time `json:"payment_forecast"`
	PaymentDate     time.Time `json:"payment_date"`
	PaymentType     string    `json:"payment_type"`
	DebtID          int       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"debt_id"`
}

func (i InstallmentDB) GetAll() ([]Installment, error) {

	var installments []Installment
	err := i.DB.Model(&Client{}).Preload("Debts").Find(&installments).Error
	return installments, err

}

type InstallmentDB struct {
	*gorm.DB
}

//creates a new category db
func NewInstallmentDB(db *gorm.DB) *InstallmentDB {
	return &InstallmentDB{
		DB: db,
	}
}

func ParseMoneyValue(value string) (float64, error) {
	value = strings.ReplaceAll(value, "R$", "")
	value = strings.ReplaceAll(value, ",", ".")
	value = strings.ReplaceAll(value, " ", "")

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return floatValue, nil
}

func ParseMoneyValueUS(value string) (float64, error) {
	value = strings.ReplaceAll(value, "$", "")
	value = strings.ReplaceAll(value, ",", ".")
	value = strings.ReplaceAll(value, " ", "")

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return floatValue, nil
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (d *InstallmentDB) CreateDB(srv *sheets.Service) error {
	d.DB.AutoMigrate(&Installment{})

	spreadsheetId := "1xAcWfbWLulOTIcPHxvKsBlLeIaQam5OQf7BMMzhm71Y"
	readRange := "parcelas!A3:I"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {
		for _, row := range resp.Values {
			//log.Fatal(row[0].(string), row[1].(string), row[2].(string), row[3].(string), row[4].(string), row[5].(string), row[6].(string))
			order, err := strconv.Atoi(row[1].(string))
			if err != nil {
				log.Fatal(err)
			}
			// value, err := strconv.ParseFloat(strings.Replace(row[2].(string), ",", ".", -1), 64)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			//log.Fatal(strings.Replace(row[2].(string), ",", ".", -1))
			value, err := ParseMoneyValue(row[2].(string))
			if err != nil {
				log.Fatal(err)
			}

			isPaid := false
			if row[3].(string) == "sim" {
				isPaid = true
			}

			model := `02/01/2006` // Este é o DD/MM/YYYY
			paymentForecast, err := time.Parse(model, row[4].(string))
			if err != nil {
				log.Fatal(err)
			}

			var paymentDate time.Time
			paymentDate, err = time.Parse(model, row[5].(string))
			if err != nil {
				paymentDate = time.Time{}
			}

			paymentType := row[6].(string)

			debtID, err := strconv.Atoi(row[7].(string))
			if err != nil {
				log.Fatal(err)
			}

			modelInstallment := &Installment{
				Order:           order,
				Value:           value,
				IsPaid:          isPaid,
				PaymentForecast: paymentForecast,
				PaymentDate:     paymentDate,
				PaymentType:     paymentType,
				DebtID:          debtID,
			}

			d.DB.Create(modelInstallment)
		}
	}

	return nil

}
