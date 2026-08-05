package model

import (
	"errors"
)

// 企业钱包流水类型
const (
	WalletTxnTypeRecharge = "recharge"
	WalletTxnTypeGrant    = "grant"
	WalletTxnTypeRecycle  = "recycle"
	WalletTxnTypeConsume  = "consume"
	WalletTxnTypeRefund   = "refund"
)

// EnterpriseWalletTxn 企业资金流水（主钱包维度）：充值/派发/回收/消费/退款。
type EnterpriseWalletTxn struct {
	Id           int   `json:"id" gorm:"primaryKey;autoIncrement"`
	EnterpriseId int   `json:"enterprise_id" gorm:"not null;index"`
	UserId       int   `json:"user_id" gorm:"not null;default:0;index"`
	Type         string `json:"type" gorm:"type:varchar(20);not null"`
	Amount       int   `json:"amount" gorm:"not null"`
	BalanceAfter int   `json:"balance_after" gorm:"not null;default:0"`
	OperatorId   int   `json:"operator_id" gorm:"not null;default:0"`
	TradeNo      string `json:"trade_no" gorm:"type:varchar(64);not null;default:'';index"`
	CreatedAt    int64  `json:"created_at" gorm:"autoCreateTime"`
}

func (EnterpriseWalletTxn) TableName() string { return "enterprise_wallet_txn" }

// GetEnterpriseWalletTxns 分页查询企业钱包流水
func GetEnterpriseWalletTxns(enterpriseId int, page int, pageSize int) ([]EnterpriseWalletTxn, int64, error) {
	if enterpriseId <= 0 {
		return nil, 0, errors.New("invalid enterprise id")
	}
	var txns []EnterpriseWalletTxn
	var total int64
	q := DB.Model(&EnterpriseWalletTxn{}).Where("enterprise_id = ?", enterpriseId)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	err := q.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&txns).Error
	return txns, total, err
}
