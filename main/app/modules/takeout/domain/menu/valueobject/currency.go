package valueobject

import "ttpos-server-go/app/errors"

// Currency 货币值对象（平台通用）
type Currency struct {
	Code     string // 货币代码，如 SGD, USD, THB
	Symbol   string // 货币符号，如 S$, $, ฿
	Exponent int    // 小数位数，一般为 2
}

// NewCurrency 创建货币值对象
func NewCurrency(code, symbol string, exponent int) (*Currency, error) {
	if code == "" {
		return nil, errors.New("货币代码不能为空")
	}
	if symbol == "" {
		return nil, errors.New("货币符号不能为空")
	}
	if exponent < 0 || exponent > 4 {
		return nil, errors.New("货币小数位数必须在 0-4 之间")
	}

	return &Currency{
		Code:     code,
		Symbol:   symbol,
		Exponent: exponent,
	}, nil
}

// Validate 验证货币值对象
func (c *Currency) Validate() error {
	if c.Code == "" {
		return errors.New("货币代码不能为空")
	}
	if c.Symbol == "" {
		return errors.New("货币符号不能为空")
	}
	if c.Exponent < 0 || c.Exponent > 4 {
		return errors.New("货币小数位数必须在 0-4 之间")
	}
	return nil
}
