package csvimporter

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type SimpleTable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func Callback(ctx *gin.Context, db *gorm.DB) error {

	state := ctx.Request.FormValue("state")
	code := ctx.Request.FormValue("code")

	if state != oauthStateString {
		return errors.New("fail")
	}
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return fmt.Errorf("code exchange failed: %s", err.Error())
	}

	client := googleOauthConfig.Client(ctx, token)

	categoryDB := NewCategoryDB(db)
	err = Import(categoryDB, client)
	if err != nil {
		fmt.Println(err)
	}

	typeDB := NewTypeDB(db)
	err = Import(typeDB, client)
	if err != nil {
		fmt.Println(err)
	}

	productDB := NewProductDB(db)
	err = Import(productDB, client)
	if err != nil {
		fmt.Println(err)
	}

	supplierDB := NewSupplierDB(db)
	err = Import(supplierDB, client)
	if err != nil {
		fmt.Println(err)
	}

	clientDB := NewClientDB(db)
	err = Import(clientDB, client)
	if err != nil {
		fmt.Println(err)
	}

	debtDB := NewDebtDB(db)
	err = Import(debtDB, client)
	if err != nil {
		fmt.Println(err)
	}

	installmentDB := NewInstallmentDB(db)
	err = Import(installmentDB, client)
	if err != nil {
		fmt.Println(err)
	}

	return nil

}
