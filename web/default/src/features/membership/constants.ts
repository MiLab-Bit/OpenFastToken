/*
Copyright (C) 2023-2026 OpenFastToken
*/
import type { MembershipLevelConfig, MembershipLevel } from './types'

/** Membership level configs for UI display */
export const MEMBERSHIP_LEVEL_CONFIGS: Record<
  MembershipLevel,
  MembershipLevelConfig
> = {
  silver: {
    key: 'silver',
    label: 'Silver',
    labelZh: '白银会员',
    discountRate: 0.98,
    discountLabel: '9.8折',
    color: 'text-muted-foreground',
    bgColor: 'bg-muted border-border',
    icon: '🥈',
    description: '享受标准会员折扣，API调用费用9.8折优惠',
  },
  gold: {
    key: 'gold',
    label: 'Gold',
    labelZh: '黄金会员',
    discountRate: 0.95,
    discountLabel: '9.5折',
    color: 'text-yellow-600',
    bgColor: 'bg-yellow-50 border-yellow-300',
    icon: '🥇',
    description: '企业认证会员，API调用费用9.5折优惠',
  },
  platinum: {
    key: 'platinum',
    label: 'Platinum',
    labelZh: '铂金会员',
    discountRate: 0.9,
    discountLabel: '9折',
    color: 'text-purple-600',
    bgColor: 'bg-purple-50 border-purple-300',
    icon: '💎',
    description: '顶级会员，API调用费用9折优惠',
  },
}

/** Enterprise status labels */
export const ENTERPRISE_STATUS_MAP: Record<string, { label: string; labelZh: string; color: string }> = {
  pending: { label: 'Pending', labelZh: '待审核', color: 'text-yellow-600 bg-yellow-50' },
  approved: { label: 'Approved', labelZh: '已通过', color: 'text-green-600 bg-green-50' },
  rejected: { label: 'Rejected', labelZh: '已拒绝', color: 'text-red-600 bg-red-50' },
}

/** Format timestamp to readable date */
export function formatExpireTime(timestamp: number): string {
  if (timestamp === 0) return '永久有效'
  const date = new Date(timestamp * 1000)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

/** Format discount rate to 折 (handles decimals like 0.98 -> 9.8折) */
export function formatDiscountRate(rate: number): string {
  const zhe = Math.round(rate * 100) / 10
  return `${zhe}折`
}