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
		// 并发首次创建时唯一索引冲突，回读已存在的行
		var existing EnterpriseWallet
		if reErr := DB.Where("enterprise_id = ?", enterpriseId).First(&existing).Error; reErr == nil {
			return &existing, nil
		}
		return nil, err
	}
	return &w, nil
}

// readBalanceInTx 在事务内回读主钱包余额，用于生成精确的流水快照。
// 数据库行锁保证同一钱包的并发更新被串行化，因此此处读到的即为本次更新后的真实余额。
func readBalanceInTx(tx *gorm.DB, walletId int) (int, error) {
	var balance int
	err := tx.Model(&EnterpriseWallet{}).Where("id = ?", walletId).Select("balance").Scan(&balance).Error
	return balance, err
}

// Recharge 平台授信或企业自助充值：增加主钱包余额与累计授予额，并写流水（单事务）。
// operatorId 为 0 表示系统/支付回调入账，非 0 表示平台管理员手工授信。
func (w *EnterpriseWallet) Recharge(amount int, operatorId int, tradeNo string) error {
	if amount <= 0 {
		return errors.New("recharge amount must be positive")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		return w.RechargeInTx(tx, amount, operatorId, tradeNo)
	})
}

// RechargeInTx 在既有事务内完成企业主钱包充值入账（支付回调使用，避免嵌套事务）。
// 仅要求 w.EnterpriseId 有效；事务内自动创建钱包行（并发首建冲突时回读）。
func (w *EnterpriseWallet) RechargeInTx(tx *gorm.DB, amount int, operatorId int, tradeNo string) error {
	if amount <= 0 {
		return errors.New("recharge amount must be positive")
	}
	if w.EnterpriseId <= 0 {
		return errors.New("invalid enterprise id")
	}
	// 事务内确保钱包行存在
	var wallet EnterpriseWallet
	err := tx.Where("enterprise_id = ?", w.EnterpriseId).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		wallet = EnterpriseWallet{EnterpriseId: w.EnterpriseId}
		if cErr := tx.Create(&wallet).Error; cErr != nil {
			// 并发首次创建唯一索引冲突，回读已存在的行
			if reErr := tx.Where("enterprise_id = ?", w.EnterpriseId).First(&wallet).Error; reErr != nil {
				return cErr
			}
		}
	} else if err != nil {
		return err
	}
	if err := tx.Model(&EnterpriseWallet{}).
		Where("id = ?", wallet.Id).
		Updates(map[string]interface{}{
			"balance":       gorm.Expr("balance + ?", amount),
			"total_granted": gorm.Expr("total_granted + ?", amount),
		}).Error; err != nil {
		return err
	}
	balanceAfter, err := readBalanceInTx(tx, wallet.Id)
	if err != nil {
		return err
	}
	return tx.Create(&EnterpriseWalletTxn{
		EnterpriseId: w.EnterpriseId,
		Type:         WalletTxnTypeRecharge,
		Amount:       amount,
		BalanceAfter: balanceAfter,
		OperatorId:   operatorId,
		TradeNo:      tradeNo,
	}).Error
}

// RechargeByOperator 平台管理员手工授信充值，remark 落到流水的 TradeNo 字段作为备注。
func (w *EnterpriseWallet) RechargeByOperator(amount int, operatorId int, remark string) error {
	return w.Recharge(amount, operatorId, remark)
}

// RefundInTx 在既有事务内冲销企业主钱包余额（管理员退款用），并写流水。
// 与个人钱包退款口径一致：只做平台内配额冲销，不触碰支付网关的实际退款。
func (w *EnterpriseWallet) RefundInTx(tx *gorm.DB, amount int, operatorId int, tradeNo string) error {
	if amount <= 0 {
		return errors.New("refund amount must be positive")
	}
	if w.EnterpriseId <= 0 {
		return errors.New("invalid enterprise id")
	}
	res := tx.Model(&EnterpriseWallet{}).
		Where("enterprise_id = ?", w.EnterpriseId).
		Update("balance", gorm.Expr("CASE WHEN balance - ? < 0 THEN 0 ELSE balance - ? END", amount, amount))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("企业钱包不存在")
	}
	var wallet EnterpriseWallet
	if err := tx.Where("enterprise_id = ?", w.EnterpriseId).First(&wallet).Error; err != nil {
		return err
	}
	balanceAfter, err := readBalanceInTx(tx, wallet.Id)
	if err != nil {
		return err
	}
	return tx.Create(&EnterpriseWalletTxn{
		EnterpriseId: w.EnterpriseId,
		Type:         WalletTxnTypeRefund,
		Amount:       amount,
		BalanceAfter: balanceAfter,
		OperatorId:   operatorId,
		TradeNo:      tradeNo,
	}).Error
}

// GrantToMember 从企业主钱包向成员派发额度：扣主钱包、增成员余额、写流水（单事务）。
// 条件更新保证余额不足时整个事务不产生任何变更。
func (w *EnterpriseWallet) GrantToMember(member *EnterpriseUser, amount int, operatorId int, tradeNo string) error {
	if amount <= 0 {
		return errors.New("grant amount must be positive")
	}
	if member == nil || member.Id <= 0 {
		return errors.New("invalid member")
	}
	if member.EnterpriseId != w.EnterpriseId {
		return errors.New("member does not belong to this enterprise")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&EnterpriseWallet{}).
			Where("id = ? AND balance >= ?", w.Id, amount).
			Update("balance", gorm.Expr("balance - ?", amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("企业钱包余额不足")
		}
		res2 := tx.Model(&EnterpriseUser{}).
			Where("id = ?", member.Id).
			Update("quota", gorm.Expr("quota + ?", amount))
		if res2.Error != nil {
			return res2.Error
		}
		if res2.RowsAffected == 0 {
			return errors.New("成员额度更新失败")
		}
		balanceAfter, err := readBalanceInTx(tx, w.Id)
		if err != nil {
			return err
		}
		return tx.Create(&EnterpriseWalletTxn{
			EnterpriseId: w.EnterpriseId,
			UserId:       member.UserId,
			Type:         WalletTxnTypeGrant,
			Amount:       amount,
			BalanceAfter: balanceAfter,
			OperatorId:   operatorId,
			TradeNo:      tradeNo,
		}).Error
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
	if member.EnterpriseId != w.EnterpriseId {
		return errors.New("member does not belong to this enterprise")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&EnterpriseUser{}).
			Where("id = ? AND quota >= ?", member.Id, amount).
			Update("quota", gorm.Expr("quota - ?", amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("成员可回收余额不足")
		}
		if err := tx.Model(&EnterpriseWallet{}).
			Where("id = ?", w.Id).
			Updates(map[string]interface{}{
				"balance":        gorm.Expr("balance + ?", amount),
				"total_recycled": gorm.Expr("total_recycled + ?", amount),
			}).Error; err != nil {
			return err
		}
		balanceAfter, err := readBalanceInTx(tx, w.Id)
		if err != nil {
			return err
		}
		return tx.Create(&EnterpriseWalletTxn{
			EnterpriseId: w.EnterpriseId,
			UserId:       member.UserId,
			Type:         WalletTxnTypeRecycle,
			Amount:       amount,
			BalanceAfter: balanceAfter,
			OperatorId:   operatorId,
			TradeNo:      tradeNo,
		}).Error
	})
}
