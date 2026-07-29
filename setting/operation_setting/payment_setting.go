package operation_setting

import (
	"encoding/json"
	"github.com/MiLab-Bit/OpenFastToken/setting"
	"github.com/MiLab-Bit/OpenFastToken/setting/config"
)

// 支付相关全局变量
var Price float64 = 1.0           // 价格（美元/单位）
var MinTopUp int = 1              // 最小充值金额（元），自定义输入最低1元
var PayAddress string = ""         // 支付地址
var CustomCallbackAddress string = "" // 自定义回调地址
var PayMethods []map[string]string = []map[string]string{} // 支付方式列表

// 支付方式定义
type PayMethod struct {
	Name   string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func PayMethods2JsonString() string {
	payMethods := []PayMethod{
		{Name: "Alipay", Enabled: setting.AlipayEnabled},
		{Name: "WeChat Pay", Enabled: setting.WechatEnabled},
	}
	data, _ := json.Marshal(payMethods)
	return string(data)
}

func UpdatePayMethodsByJsonString(jsonStr string) error {
	var payMethods []PayMethod
	err := json.Unmarshal([]byte(jsonStr), &payMethods)
	if err != nil {
		return err
	}
	return nil
}

type PaymentSetting struct {
	AmountOptions  []int           `json:"amount_options"`
	AmountDiscount map[int]float64 `json:"amount_discount"`

	ComplianceConfirmed    bool   `json:"compliance_confirmed"`
	ComplianceTermsVersion string `json:"compliance_terms_version"`
	ComplianceConfirmedAt  int64  `json:"compliance_confirmed_at"`
	ComplianceConfirmedBy  int    `json:"compliance_confirmed_by"`
	ComplianceConfirmedIP  string `json:"compliance_confirmed_ip"`
}

const CurrentComplianceTermsVersion = "v1"

var paymentSetting = PaymentSetting{
	AmountOptions: []int{100, 200, 500, 1000},
	AmountDiscount: map[int]float64{},
}

func init() {
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}

func IsPaymentComplianceConfirmed() bool {
	return paymentSetting.ComplianceConfirmed &&
		paymentSetting.ComplianceTermsVersion == CurrentComplianceTermsVersion
}