package service

import (
	"fmt"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

// ---------------------------------------------------------------------------
// FundingSource — 资金来源接口（个人钱包 / 企业钱包）
// ---------------------------------------------------------------------------

// FundingSource 抽象了预扣费的资金来源。
// 实现者必须保证 PreConsume / Settle / Refund / Reserve 之间的额度守恒。
type FundingSource interface {
	// Source 返回资金来源标识："wallet" | "enterprise"
	Source() string
	// PreConsume 从该资金来源预扣 amount 额度
	PreConsume(amount int) error
	// Settle 根据差额调整资金来源（正数补扣，负数退还）
	Settle(delta int) error
	// Refund 退还所有预扣费
	Refund() error
	// Reserve 在已有预扣基础上追加预扣 delta（流式请求中途补扣）
	Reserve(delta int) error
	// RollbackReserve 回滚一次失败的 Reserve
	RollbackReserve(delta int)
}

// ---------------------------------------------------------------------------
// WalletFunding — 个人钱包资金来源（users.quota）
// ---------------------------------------------------------------------------

type WalletFunding struct {
	userId   int
	consumed int // 实际预扣的用户额度
}

func (w *WalletFunding) Source() string { return BillingSourceWallet }

func (w *WalletFunding) PreConsume(amount int) error {
	if amount <= 0 {
		return nil
	}
	if err := model.DecreaseUserQuota(w.userId, amount, false); err != nil {
		return err
	}
	w.consumed = amount
	return nil
}

func (w *WalletFunding) Settle(delta int) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return model.DecreaseUserQuota(w.userId, delta, false)
	}
	return model.IncreaseUserQuota(w.userId, -delta, false)
}

func (w *WalletFunding) Refund() error {
	if w.consumed <= 0 {
		return nil
	}
	return model.IncreaseUserQuota(w.userId, w.consumed, false)
}

func (w *WalletFunding) Reserve(delta int) error {
	if delta <= 0 {
		return nil
	}
	if err := model.DecreaseUserQuota(w.userId, delta, false); err != nil {
		return err
	}
	w.consumed += delta
	return nil
}

func (w *WalletFunding) RollbackReserve(delta int) {
	if delta <= 0 {
		return
	}
	if err := model.IncreaseUserQuota(w.userId, delta, false); err != nil {
		common.SysLog("error rolling back wallet funding reserve: " + err.Error())
		return
	}
	w.consumed -= delta
}

// ---------------------------------------------------------------------------
// EnterpriseFunding — 企业钱包资金来源（enterprise_user.quota）
// ---------------------------------------------------------------------------

// EnterpriseFunding 从企业派发给成员的余额中扣费。
// enterprise_user.quota 为成员真实可用余额，used_quota 由 model 层同步维护。
type EnterpriseFunding struct {
	euId         int // enterprise_user 主键
	enterpriseId int
	userId       int
	consumed     int
}

func (e *EnterpriseFunding) Source() string { return BillingSourceEnterprise }

func (e *EnterpriseFunding) PreConsume(amount int) error {
	if amount <= 0 {
		return nil
	}
	if err := model.ConsumeEUQuota(e.euId, amount); err != nil {
		return err
	}
	e.consumed = amount
	return nil
}

func (e *EnterpriseFunding) Settle(delta int) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return model.ConsumeEUQuota(e.euId, delta)
	}
	return model.RefundEUQuota(e.euId, -delta)
}

func (e *EnterpriseFunding) Refund() error {
	if e.consumed <= 0 {
		return nil
	}
	return model.RefundEUQuota(e.euId, e.consumed)
}

func (e *EnterpriseFunding) Reserve(delta int) error {
	if delta <= 0 {
		return nil
	}
	if err := model.ConsumeEUQuota(e.euId, delta); err != nil {
		return err
	}
	e.consumed += delta
	return nil
}

func (e *EnterpriseFunding) RollbackReserve(delta int) {
	if delta <= 0 {
		return
	}
	if err := model.RefundEUQuota(e.euId, delta); err != nil {
		common.SysLog("error rolling back enterprise funding reserve: " + err.Error())
		return
	}
	e.consumed -= delta
}

// ---------------------------------------------------------------------------
// CompositeFunding — 企业优先 / 个人兜底（单请求不混合）
// ---------------------------------------------------------------------------

// CompositeFunding 在预扣阶段一次性选定资金来源，之后所有操作都锁定在同一来源上。
// 选路规则（用户拍板）：
//  1. 用户是 active 企业成员且企业余额足够本次预扣 → 走企业钱包
//  2. 否则 → 走个人钱包
//
// 单次请求内绝不跨钱包拆分，退款一律原路返回。
type CompositeFunding struct {
	enterprise *EnterpriseFunding // 可能为 nil（非企业成员）
	wallet     *WalletFunding
	active     FundingSource // 选定后不再变更
}

// NewCompositeFunding 构造资金来源。membership 为 nil 表示用户不属于任何企业。
func NewCompositeFunding(userId int, membership *model.EnterpriseMembership) *CompositeFunding {
	cf := &CompositeFunding{
		wallet: &WalletFunding{userId: userId},
	}
	if membership != nil && membership.EnterpriseUserId > 0 {
		cf.enterprise = &EnterpriseFunding{
			euId:         membership.EnterpriseUserId,
			enterpriseId: membership.EnterpriseId,
			userId:       userId,
		}
	}
	return cf
}

// HasEnterprise 返回用户是否具备企业资金来源
func (c *CompositeFunding) HasEnterprise() bool { return c.enterprise != nil }

// EnterpriseId 返回绑定的企业 ID，无企业身份时返回 0
func (c *CompositeFunding) EnterpriseId() int {
	if c.enterprise == nil {
		return 0
	}
	return c.enterprise.enterpriseId
}

func (c *CompositeFunding) Source() string {
	if c.active == nil {
		// 尚未选路（信任旁路下 amount=0 也会预先选路，正常不会走到这里）
		return BillingSourceWallet
	}
	return c.active.Source()
}

// ActiveEnterpriseUserId 返回本次请求实际选中的 enterprise_user.id。
// 未选中企业钱包时返回 0。异步任务据此把退款原路退回企业钱包。
func (c *CompositeFunding) ActiveEnterpriseUserId() int {
	if c.enterprise == nil || c.active != FundingSource(c.enterprise) {
		return 0
	}
	return c.enterprise.euId
}

// selectByBalance 在无需实际扣款时（amount == 0）按余额预选路径
func (c *CompositeFunding) selectByBalance() {
	if c.active != nil {
		return
	}
	if c.enterprise != nil {
		if q, err := model.GetEUQuotaById(c.enterprise.euId); err == nil && q > 0 {
			c.active = c.enterprise
			return
		}
	}
	c.active = c.wallet
}

func (c *CompositeFunding) PreConsume(amount int) error {
	if amount <= 0 {
		c.selectByBalance()
		return nil
	}
	// 企业优先
	if c.enterprise != nil {
		err := c.enterprise.PreConsume(amount)
		if err == nil {
			c.active = c.enterprise
			return nil
		}
		// 企业余额不足 → 回退个人钱包（不混合，全额重试）
		common.SysLog(fmt.Sprintf("enterprise funding unavailable for user %d (eu=%d, amount=%d): %s, falling back to personal wallet",
			c.wallet.userId, c.enterprise.euId, amount, err.Error()))
	}
	if err := c.wallet.PreConsume(amount); err != nil {
		return err
	}
	c.active = c.wallet
	return nil
}

func (c *CompositeFunding) Settle(delta int) error {
	if c.active == nil {
		c.selectByBalance()
	}
	return c.active.Settle(delta)
}

func (c *CompositeFunding) Refund() error {
	if c.active == nil {
		return nil
	}
	return c.active.Refund()
}

func (c *CompositeFunding) Reserve(delta int) error {
	if c.active == nil {
		c.selectByBalance()
	}
	return c.active.Reserve(delta)
}

func (c *CompositeFunding) RollbackReserve(delta int) {
	if c.active == nil {
		return
	}
	c.active.RollbackReserve(delta)
}
