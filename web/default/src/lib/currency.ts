/*
Copyright (C) 2023-2026 OpenFastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@example.com
*/

import i18n from 'i18next'
import { getIntlLocale } from './locale'

/**
 * ============================================================================
 * 简化货币格式化库
 * ============================================================================
 * 
 * 只支持人民币（元）显示，去掉所有汇率转换逻辑。
 * 
 * 核心概念：
 * - 系统单位：数据库存储的单位（500000 系统单位 = 1 元）
 * - 元：显示单位（人民币）
 * 
 * 转换公式：
 * 元 = 系统单位 / 500
 */

// 人民币基准：¥1 = 500000 系统单位（QuotaPerUnit）
const RMB = 500000

// 格式化选项
interface CurrencyFormatOptions {
  digitsLarge?: number
  digitsSmall?: number
  abbreviate?: boolean
}

const DEFAULT_OPTIONS = {
  digitsLarge: 2,
  digitsSmall: 4,
  abbreviate: true,
}

/**
 * 将系统单位转换为元
 * @param quota - 系统单位
 * @returns 元（人民币）
 */
function quotaToYuan(quota: number): number {
  return quota / RMB
}

/**
 * 格式化数字（带千分位）
 */
function formatCNY(value: number, digits: number): string {
  return new Intl.NumberFormat(getIntlLocale(), {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 0,
    maximumFractionDigits: digits,
  }).format(value)
}


/**
 * 格式化系统单位为元显示
 * 
 * @param quota - 系统单位
 * @param options - 格式化选项
 * @returns 格式化后的字符串（如：1,440.00元）
 * 
 * @example
 * formatQuota(720000) → 1,440.00元
 */
export function formatCurrencyFromUSD(
  quota: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  if (quota == null || Number.isNaN(quota)) return '-'
  
  const opts = { ...DEFAULT_OPTIONS, ...options }
  const yuan = quotaToYuan(quota)
  const digits = Math.abs(yuan) >= 1 ? opts.digitsLarge : opts.digitsSmall
  
  return formatCNY(yuan, digits)
}

/**
 * 格式化系统单位为元显示（计费专用，永远显示元）
 * 
 * @param quota - 系统单位
 * @param options - 格式化选项
 * @returns 格式化后的字符串（如：1,440.00元）
 */
export function formatBillingCurrencyFromUSD(
  quota: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  // 永远显示为元
  return formatCurrencyFromUSD(quota, options)
}

/**
 * 格式化原始系统单位为元
 * 
 * @param quota - 原始系统单位
 * @param options - 格式化选项
 * @returns 格式化后的字符串（如：1,440.00元）
 */
export function formatQuotaWithCurrency(
  quota: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  return formatCurrencyFromUSD(quota, options)
}

/**
 * 获取货币标签
 * @returns 元
 */
export function getCurrencyLabel(): string {
  const label = i18n.t('currency.cny')
  return label && label !== 'currency.cny' ? label : '元'
}

/**
 * 检查是否启用货币显示
 * @returns true（永远启用）
 */
export function isCurrencyDisplayEnabled(): boolean {
  return true
}

/**
 * 格式化已经是元的金额（用于支付金额）
 * 
 * @param amount - 已经是元的金额
 * @param options - 格式化选项
 * @returns 格式化后的字符串（如：1,440.00元）
 */
export function formatLocalCurrencyAmount(
  amount: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  if (amount == null || Number.isNaN(amount)) return '-'
  
  const opts = { ...DEFAULT_OPTIONS, ...options }
  return formatCNY(amount, opts.digitsLarge)
}

/**
 * 货币显示类型
 */
export type CurrencyDisplayKind = 'currency' | 'tokens' | 'custom'

/**
 * 货币显示配置（简化版）
 */
export interface CurrencyDisplayConfig {
  quotaPerUnit?: number
  kind?: CurrencyDisplayKind
  exchangeRate?: number
  // 允许后端返回额外的扩展字段
  [key: string]: unknown
}

/**
 * 货币显示返回结构
 */
export interface CurrencyDisplay {
  config: { quotaPerUnit: number }
  meta: { kind: CurrencyDisplayKind; exchangeRate: number }
}

/**
 * 获取货币显示配置（简化版）
 *
 * @returns 简化的配置对象
 */
export function getCurrencyDisplay(): CurrencyDisplay {
  return {
    config: {
      quotaPerUnit: 500000, // 1 元 = 500000 tokens
    },
    meta: {
      kind: 'currency',
      exchangeRate: 1,
    },
  }
}

/**
 * 获取货币显示元数据（简化版）
 *
 * @param config - 货币配置
 * @returns 显示元数据
 */
export function getDisplayMeta(_config: CurrencyDisplayConfig): {
  kind: CurrencyDisplayKind
} {
  return {
    kind: 'currency',
  }
}

/**
 * 检查是否为有效的货币显示类型（简化版）
 * 
 * @param value - 要检查的值
 * @returns 是否为有效的货币显示类型
 */
export function isCurrencyDisplayType(value: unknown): value is 'CNY' | 'TOKENS' {
  return value === 'CNY' || value === 'TOKENS'
}

/**
 * 解析货币显示类型（简化版）
 * 
 * @param value - 要解析的值
 * @param fallback - 回退值（默认 'CNY'）
 * @returns 货币显示类型
 */
export function parseCurrencyDisplayType(
  value: unknown,
  fallback: 'CNY' | 'TOKENS' = 'CNY'
): 'CNY' | 'TOKENS' {
  return isCurrencyDisplayType(value) ? value : fallback
}
