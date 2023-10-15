package csvimporter

import (
	"errors"
	"log"
	"strconv"
	"time"

	"google.golang.org/api/sheets/v4"
	"gorm.io/gorm"
)

type Purchase struct {
	gorm.Model
	SolicitationDate time.Time `json:"solicitation_date"`
	AliOrderNumber   string    `json:"ali_order_number"`
	PackageNick      int       `json:"package_nick"`
	TrackingCode     string    `json:"tracking_code"`

	PaymentDate      time.Time `json:"payment_date"`
	AmountPaied      float64   `json:"amount_date"`
	DollarAliexpress float64   `json:"dolar_aliexpress"`
	AmountPaiedUS    float64   `json:"amount_date_us"`
	TaxedAmount      float64   `json:"taxed_amount"`

	PostDate     time.Time `json:"post_date"`
	DeliveryDate time.Time `json:"delivery_date"`

	Observation string `json:"observation"`

	SupplierID int      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"supplier_id"`
	Supplier   Supplier `json:"supplier"`
}

func (p PurchaseDB) GetAll() ([]Purchase, error) {

	var purchases []Purchase
	err := p.DB.Model(&Client{}).Preload("Purchase").Find(&purchases).Error
	return purchases, err
}

type PurchaseDB struct {
	*gorm.DB
}

//creates a new category db
func NewPurchaseDB(db *gorm.DB) *PurchaseDB {
	return &PurchaseDB{
		DB: db,
	}
}

func DefineSheetStrValue(value string) string {
	if len(value) == 0 {
		return "n/a"
	} else {
		return value
	}
}

func DefineSheetFloatValue(value string) float64 {
	return 0.0
}

//Gets category data from google sheets especific tab. If the proccess fail will thorw an error.
func (p *PurchaseDB) CreateDB(srv *sheets.Service) error {
	p.DB.AutoMigrate(&Purchase{})

	spreadsheetId := "1xAcWfbWLulOTIcPHxvKsBlLeIaQam5OQf7BMMzhm71Y"
	readRange := "compras!A3:N"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	} else {

		for _, row := range resp.Values {

			formato := `02/01/2006` // Este é o DD/MM/YYYY
			solicitationDate, err := time.Parse(formato, row[1].(string))
			if err != nil {
				log.Fatal(err)
			}

			supplierID, err := strconv.ParseInt(row[2].(string), 10, 64)
			if err != nil {
				return err
			}

			var aliOrderNumber string
			aliOrderNumber = DefineSheetStrValue(row[3].(string))

			packageNick, err := strconv.Atoi(row[4].(string))
			if err != nil {
				log.Fatal(err)
			}

			var trackingCode string
			trackingCode = DefineSheetStrValue(row[5].(string))

			var paymentDate time.Time
			paymentDate, err = time.Parse(formato, row[6].(string))
			if err != nil {
				paymentDate = time.Time{}
			}

			amountPaied, err := ParseMoneyValue(row[7].(string))
			if err != nil {
				log.Fatal(err)
			}

			dollarAliexpress, _ := ParseMoneyValue(row[8].(string))
			// if err != nil {
			// 	log.Fatal(err)
			// }

			amountPaiedUS, _ := ParseMoneyValueUS(row[9].(string))
			// if err != nil {
			// 	log.Fatal(err)
			// }

			taxedAmount, _ := ParseMoneyValue(row[10].(string))
			// if err != nil {
			// 	log.Fatal(err)
			// }

			var postDate time.Time
			postDate, err = time.Parse(formato, row[11].(string))
			if err != nil {
				postDate = time.Time{}
			}

			var deliveryDate time.Time
			deliveryDate, err = time.Parse(formato, row[12].(string))
			if err != nil {
				deliveryDate = time.Time{}
			}

			obs := row[13].(string)

			modelPurchase := &Purchase{
				SolicitationDate: solicitationDate,
				SupplierID:       int(supplierID),
				AliOrderNumber:   aliOrderNumber,
				PackageNick:      packageNick,
				TrackingCode:     trackingCode,
				PaymentDate:      paymentDate,
				AmountPaied:      amountPaied,
				DollarAliexpress: dollarAliexpress,
				AmountPaiedUS:    amountPaiedUS,
				TaxedAmount:      taxedAmount,
				PostDate:         postDate,
				DeliveryDate:     deliveryDate,
				Observation:      obs,
			}

			p.DB.Create(modelPurchase)
		}
	}

	return nil

}
