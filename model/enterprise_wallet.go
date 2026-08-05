package model

import (
	"errors"

	"gorm.io/gorm"
)

// EnterpriseWallet 企业主钱包：企业账户维度的资金池，由平台授信或企业自助充值入账，
// 经企业管理员派发到成员 enterprise_user.quota 后供消费。
type EnterpriseWallet struct {
	Id            int   `json:"id" gorm:"primaryKey;autoIncrement"`
	EnterpriseId  int   `json:"enterprise_id" gorm:"not null;uniqueIndex"`
	Balance       int   `json:"balance" gorm:"not null;default:0"`
	TotalGranted  int   `json:"total_granted" gorm:"not null;default:0"`
	TotalRecycled int   `json:"total_recycled" gorm:"not null;default:0"`
	CreatedAt     int64 `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     int64 `json:"updated_at" gorm:"autoUpdateTime"`
}

func (EnterpriseWallet) TableName() string { return "enterprise_wallet" }

// GetOrCreateEnterpriseWallet 获取或创建企业主钱包（首次访问时建一行）
func GetOrCreateEnterpriseWallet(enterpriseId int) (*EnterpriseWallet, error) {
	if enterpriseId <= 0 {
		return nil, errors.New("invalid enterprise id")
	}
	var w EnterpriseWallet
	err := DB.Where("enterprise_id = ?", enterpriseId).First(&w).Error
	if err == nil {
		return &w, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	w = EnterpriseWallet{EnterpriseId: enterpriseId}
	if err := DB.Create(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

// Recharge 平台授信或企业自助充值：增加主钱包余额与累计授予额，并写流水（单事务）。
func (w *EnterpriseWallet) Recharge(amount int, tradeNo string) error {
	if amount <= 0 {
		return errors.New("recharge amount must be positive")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&EnterpriseWallet{}).
			Where("id = ?", w.Id).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&EnterpriseWallet{}).
			Where("id = ?", w.Id).
			Update("total_granted", gorm.Expr("total_granted + ?", amount)).Error; err != nil {
			return err
		}
		txn := &EnterpriseWalletTxn{
			EnterpriseId:  w.EnterpriseId,
			Type:         WalletTxnTypeRecharge,
			Amount:       amount,
			BalanceAfter: w.Balance + amount,
			OperatorId:   0,
			TradeNo:      tradeNo,
		}
		return tx.Create(txn).Error
	})
}

// GrantToMember 从企业主钱包向成员派发额度：扣主钱包、增成员余额、写流水（单事务）。
func (w *EnterpriseWallet) GrantToMember(member *EnterpriseUser, amount int, operatorId int, tradeNo string) error {
	if amount <= 0 {
		return errors.New("grant amount must be positive")
	}
	if member == nil || member.Id <= 0 {
		return errors.New("invalid member")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&EnterpriseWallet{}).
			Where("id = ? AND balance >= ?", w.Id, amount).
			Update("balance", gorm.Expr("balance - ?", amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("enterprise wallet balance insufficient")
		}
		res2 := tx.Model(&EnterpriseUser{}).
			Where("id = ? AND quota + ? >= 0", member.Id, amount).
			Update("quota", gorm.Expr("quota + ?", amount))
		if res2.Error != nil {
			return res2.Error
		}
		if res2.RowsAffected == 0 {
			return errors.New("member quota update failed")
		}
		txn := &EnterpriseWalletTxn{
			EnterpriseId:  w.EnterpriseId,
			UserId:       member.UserId,
			Type:         WalletTxnTypeGrant,
			Amount:       amount,
			BalanceAfter: w.Balance - amount,
			OperatorId:   operatorId,
			TradeNo:      tradeNo,
		}
		return tx.Create(txn).Error
	})
}

// RecycleFromMember 从成员回收额度回企业主钱包：扣成员余额、增主钱包、写流水（单事务）。
func (w *EnterpriseWallet) RecycleFromMember(member *EnterpriseUser, amount int, operatorId int, tradeNo string) error {
	if amount <= 0 {
		return errors.New("recycle amount must be positive")
	}
	if member == nil || member.Id <= 0 {
		return errors.New("invalid member")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&EnterpriseUser{}).
			Where("id = ? AND quota >= ?", member.Id, amount).
			Update("quota", gorm.Expr("quota - ?", amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("member quota insufficient to recycle")
		}
		if err := tx.Model(&EnterpriseWallet{}).
			Where("id = ?", w.Id).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&EnterpriseWallet{}).
			Where("id = ?", w.Id).
			Update("total_recycled", gorm.Expr("total_recycled + ?", amount)).Error; err != nil {
			return err
		}
		txn := &EnterpriseWalletTxn{
			EnterpriseId:  w.EnterpriseId,
			UserId:       member.UserId,
			Type:         WalletTxnTypeRecycle,
			Amount:       amount,
			BalanceAfter: w.Balance + amount,
			OperatorId:   operatorId,
			TradeNo:      tradeNo,
		}
		return tx.Create(txn).Error
	})
}
