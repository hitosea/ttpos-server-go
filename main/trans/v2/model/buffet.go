package model

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	v1 "ttpos-server-go/trans/v1"
)

func NewBuffet(buffet *v1.Buffet) (*model.BuffetPackage, error) {
	languageName, err := NewMultiLanguageName(buffet.Name)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMultiLanguageName failed")
	}

	buffetCustomerTypePrices := make([]model.BuffetCustomerTypePrice, 0)
	for _, buffetCustomer := range buffet.BuffetCustomers {
		buffetCustomerTypePrices = append(buffetCustomerTypePrices, model.BuffetCustomerTypePrice{
			BaseModel: model.BaseModel{
				Uuid:       uint64(buffetCustomer.ID),
				CreateTime: buffetCustomer.CreateTime,
				UpdateTime: buffetCustomer.UpdateTime,
			},
			BuffetPackageUuid: uint64(buffet.ID),
			CustomerTypeUuid:  uint64(buffetCustomer.CustomerTypeID),
			Price:             buffetCustomer.Price,
		})
	}

	buffetProducts := make([]model.BuffetProduct, 0)
	for _, buffetProduct := range buffet.BuffetProducts {
		buffetProducts = append(buffetProducts, model.BuffetProduct{
			BaseModel: model.BaseModel{
				Uuid:       uint64(buffetProduct.ID),
				CreateTime: buffetProduct.CreateTime,
				UpdateTime: buffetProduct.UpdateTime,
			},
			ProductPackageUuid: uint64(buffetProduct.ProductID),
			BuffetPackageUuid:  uint64(buffet.ID),
			IsShowCashier:      uint(buffetProduct.IsShowCashier),
			IsShowTablet:       uint(buffetProduct.IsShowTablet),
			IsShowKitchen:      uint(buffetProduct.IsShowKitchen),
			IsShowAssistant:    uint(buffetProduct.IsShowAssistant),
			Limit:              uint(buffetProduct.LimitNum),
		})
	}

	buffetPackage := &model.BuffetPackage{
		BaseModel: model.BaseModel{
			Uuid:       uint64(buffet.ID),
			CreateTime: buffet.CreateTime,
			UpdateTime: buffet.UpdateTime,
			DeleteTime: buffet.DeleteTime,
		},
		Name:                     languageName.ToJson(),
		MultiLanguageNameUuid:    uint64(languageName.Uuid),
		Sort:                     uint(buffet.Sort),
		TaxUuid:                  buffet.GetBuffetTaxUuid(),
		IsLimitTime:              buffet.GetIsTimeLimit(),
		LimitTime:                uint(buffet.TimeLimit),
		CanCombined:              uint(buffet.IsComb),
		NonOrderingTime:          uint(buffet.RemainContinueTime),
		ReminderOrderTime:        uint(buffet.RemainContinueNoticeTime),
		ActualSaleNum:            float64(buffet.SaleNum),
		MultiLanguageName:        *languageName,
		BuffetCustomerTypePrices: buffetCustomerTypePrices,
		BuffetProducts:           buffetProducts,
	}
	return buffetPackage, nil
}
