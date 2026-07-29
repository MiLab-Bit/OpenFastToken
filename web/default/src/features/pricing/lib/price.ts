/*
 * 定价计算 - 人民币为核心（元）
 *
 * 设计：
 * - model_ratio = 输入价格（元/MTokens），直接来自后端配置
 * - completion_ratio = 输出价格 / 输入价格（倍率）
 * - 计算结果显示为 "数字+元"，不做任何系统单位转换
 * - showWithRecharge=true: 充值视图，应用会员专属 group_ratio 折扣
 * - showWithRecharge=false: 标准视图，显示基础价格 (group_ratio=1)
 */
import { QUOTA_TYPE_VALUES, TOKEN_UNIT_DIVISORS } from '../constants'
import type { PricingModel, TokenUnit, PriceType } from '../types'

/**
 * 去掉末尾多余的 0（如 2.14000 → 2.14）
 */
export function stripTrailingZeros(formatted: string): string {
  return formatted.replace(/\.0+$/, '').replace(/(\.[0-9]*?)0+$/, '')
}

/**
 * 获取用户组最低倍率
 */
function getMinGroupRatio(
  enableGroups: string[],
  groupRatio: Record<string, number>
): number {
  if (enableGroups.length === 0) return 1
  let minRatio = Number.POSITIVE_INFINITY
  for (const group of enableGroups) {
    const ratio = groupRatio[group]
    if (ratio !== undefined && ratio < minRatio) minRatio = ratio
  }
  return minRatio === Number.POSITIVE_INFINITY ? 1 : minRatio
}

/**
 * 计算模型价格（单位：元/MTokens）
 *
 * model_ratio 直接等于 元/MTokens，不需要任何转换
 *   deepseek-v3: model_ratio=2 → 输入 2元/MTokens
 *   deepseek-v4-pro: model_ratio=12 → 输入 12元/MTokens
 */
function calculateTokenPrice(
  model: PricingModel,
  type: PriceType,
  groupRatioValue: number
): number {
  // groupRatioValue: 用户组倍率（1=原价，0.8=8折）
  const inputPrice = model.model_ratio * groupRatioValue  // 元/MTokens

  switch (type) {
    case 'input':
      return inputPrice
    case 'output':
      return inputPrice * (model.completion_ratio || 1)
    case 'cache':
      return model.cache_ratio != null ? inputPrice * Number(model.cache_ratio) : NaN
    case 'create_cache':
      return model.create_cache_ratio != null ? inputPrice * Number(model.create_cache_ratio) : NaN
    case 'image':
      return model.image_ratio != null ? inputPrice * Number(model.image_ratio) : NaN
    case 'audio_input':
      return model.audio_ratio != null ? inputPrice * Number(model.audio_ratio) : NaN
    case 'audio_output':
      return model.audio_ratio != null && model.audio_completion_ratio != null
        ? inputPrice * Number(model.audio_ratio) * Number(model.audio_completion_ratio)
        : NaN
  }
}

/**
 * 格式化数字（带千分位，固定小数位）
 */
function formatNumber(num: number, digits: number): string {
  if (isNaN(num)) return '-'
  // 去掉末尾多余的0
  const fixed = num.toFixed(digits)
  const cleaned = fixed.replace(/\.?0+$/, '')
  // 加千分位
  const parts = cleaned.split('.')
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return parts.join('.')
}

/**
 * 获取模型的会员折扣倍率
 * 返回用于显示的最小 group_ratio（实际折扣）
 * 同时返回 membershipDiscount 用于标签显示
 */
export function getModelDiscountInfo(
  model: PricingModel
): { minRatio: number; hasDiscount: boolean } {
  const enableGroups = Array.isArray(model.enable_groups) ? model.enable_groups : []
  const groupRatio = model.group_ratio || {}
  const minRatio = getMinGroupRatio(enableGroups, groupRatio)
  return {
    minRatio,
    hasDiscount: minRatio > 0 && minRatio < 1,
  }
}

/**
 * 格式化价格显示
 *
 * @param tokenUnit - 显示单位：'M' = 元/MTokens, 'K' = 元/KTokens
 * @param showWithRecharge - 充值视图模式：true=应用会员 group_ratio 折扣，false=显示标准基础价
 * @param priceRate - 充值折扣倍率（如 0.95 = 95折），仅在 showWithRecharge=true 时额外生效
 * @param membershipDiscountRate - 会员等级折扣（silver=0.98, gold=0.95, platinum=0.9），充值视图生效
 */
export function formatPrice(
  model: PricingModel,
  type: PriceType,
  tokenUnit: TokenUnit,
  showWithRecharge = false,
  priceRate = 1,
  _usdExchangeRate?: number,
  membershipDiscountRate = 1
): string {
  if (model.quota_type === QUOTA_TYPE_VALUES.REQUEST) return '-'

  // 标准视图: 显示基础价格 (group_ratio=1)
  // 充值视图: 显示会员专属 group_ratio 折扣价
  const enableGroups = Array.isArray(model.enable_groups) ? model.enable_groups : []
  const groupRatio = model.group_ratio || {}
  const minRatio = showWithRecharge
    ? getMinGroupRatio(enableGroups, groupRatio)
    : 1

  let price = calculateTokenPrice(model, type, minRatio)

  // 充值倍率额外折扣
  if (showWithRecharge && priceRate !== 1) {
    price = price * priceRate
  }

  // 会员等级折扣（充值视图生效：silver=0.98, gold=0.95, platinum=0.9）
  if (showWithRecharge && membershipDiscountRate !== 1) {
    price = price * membershipDiscountRate
  }

  // 转换为指定单位（M: 不除, K: 除以1000）
  const unitPrice = price / TOKEN_UNIT_DIVISORS[tokenUnit]

  if (isNaN(unitPrice)) return '-'

  // 根据数值大小决定小数位
  const digits = Math.abs(unitPrice) >= 1 ? 4 : 6
  return formatNumber(unitPrice, digits) + '元'
}

export function formatGroupPrice(
  model: PricingModel,
  group: string,
  type: PriceType,
  tokenUnit: TokenUnit,
  showWithRecharge = false,
  priceRate = 1,
  _usdExchangeRate?: number,
  groupRatio?: Record<string, number>,
  membershipDiscountRate = 1
): string {
  if (model.quota_type === QUOTA_TYPE_VALUES.REQUEST) return '-'

  const ratio = (groupRatio || {})[group] || 1
  let price = calculateTokenPrice(model, type, ratio)

  if (showWithRecharge && priceRate !== 1) {
    price = price * priceRate
  }

  if (showWithRecharge && membershipDiscountRate !== 1) {
    price = price * membershipDiscountRate
  }

  const unitPrice = price / TOKEN_UNIT_DIVISORS[tokenUnit]
  if (isNaN(unitPrice)) return '-'

  const digits = Math.abs(unitPrice) >= 1 ? 4 : 6
  return formatNumber(unitPrice, digits) + '元'
}

/**
 * 格式化固定价格（按次计费模型）
 */
export function formatFixedPrice(
  model: PricingModel,
  group: string,
  showWithRecharge = false,
  priceRate = 1,
  _usdExchangeRate?: number,
  groupRatio?: Record<string, number>,
  membershipDiscountRate = 1
): string {
  if (model.quota_type !== QUOTA_TYPE_VALUES.REQUEST) return '-'

  const ratio = (groupRatio || {})[group] || 1
  let price = (model.model_price || 0) * ratio

  if (showWithRecharge && priceRate !== 1) {
    price = price * priceRate
  }

  if (showWithRecharge && membershipDiscountRate !== 1) {
    price = price * membershipDiscountRate
  }

  if (isNaN(price)) return '-'
  return formatNumber(price, 4) + '元'
}

/**
 * 格式化请求计费价格（按次）
 */
export function formatRequestPrice(
  model: PricingModel,
  showWithRecharge = false,
  priceRate = 1,
  _usdExchangeRate?: number,
  membershipDiscountRate = 1
): string {
  if (model.quota_type !== QUOTA_TYPE_VALUES.REQUEST) return '-'

  const enableGroups = Array.isArray(model.enable_groups) ? model.enable_groups : []
  const groupRatio = model.group_ratio || {}
  const minRatio = showWithRecharge
    ? getMinGroupRatio(enableGroups, groupRatio)
    : 1

  let price = (model.model_price || 0) * minRatio

  if (showWithRecharge && priceRate !== 1) {
    price = price * priceRate
  }

  if (showWithRecharge && membershipDiscountRate !== 1) {
    price = price * membershipDiscountRate
  }

  if (isNaN(price)) return '-'
  return formatNumber(price, 4) + '元'
}
/**
 * 视频分段计费模型名称前缀匹配
 */
const VIDEO_TIERED_MODEL_PREFIXES = [
  'happyhorse-1.0-t2v',
  'happyhorse-1.0-i2v',
  'happyhorse-1.0-r2v',
  'happyhorse-1.0-video-edit',
]

/**
 * 检查模型是否为视频分段计费模型
 */
export function isVideoTieredModel(model: PricingModel): boolean {
  if (model.quota_type !== QUOTA_TYPE_VALUES.REQUEST) return false
  const name = model.model_name || ''
  return VIDEO_TIERED_MODEL_PREFIXES.some((p) => name.startsWith(p))
}

/**
 * 视频分段定价信息（元/秒）
 * 基准: 720P = model_ratio 元/秒, 1080P = 按比例上浮
 */
export interface VideoTierPriceInfo {
  price720p: number   // 720P 单价（元/秒）
  price1080p: number  // 1080P 单价（元/秒）
  unit: string        // '元/秒'
}

/**
 * 获取视频模型的分段定价信息
 */
export function getVideoTierPrice(model: PricingModel): VideoTierPriceInfo | null {
  if (!isVideoTieredModel(model)) return null
  // 视频模型价格固定（720P=0.9元/秒, 1080P=1.6元/秒），不依赖 ratio 表、会员/API 折扣均不生效
  return {
    price720p: 0.9,
    price1080p: 1.6,
    unit: '元/秒',
  }
}

/**
 * 格式化视频分段计费价格显示
 * 返回格式: "1.2元/秒(720P) / 2.0元/秒(1080P)"
 */
export function formatVideoTieredPrice(
  model: PricingModel,
  _showWithRecharge = false,
  _priceRate = 1,
  _usdExchangeRate?: number,
  _membershipDiscountRate = 1
): string | null {
  // 视频模型价格固定，会员折扣与 API 折扣计费均不生效
  const info = getVideoTierPrice(model)
  if (!info) return null

  const p720 = info.price720p
  const p1080 = info.price1080p

  return `${formatNumber(parseFloat(p720.toFixed(2)), 2)}元/秒(720P) / ${formatNumber(parseFloat(p1080.toFixed(2)), 2)}元/秒(1080P)`
}
