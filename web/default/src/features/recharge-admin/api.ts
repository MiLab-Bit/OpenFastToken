import { api } from '@/lib/api'
import type { AdminTopUpsResponse } from './types'

export interface GetAdminTopUpsParams {
  page?: number
  page_size?: number
  user_id?: string
  keyword?: string
}

/**
 * 管理员获取充值订单列表，支持按用户ID、订单号关键字筛选
 */
export async function getAdminTopUps(
  params: GetAdminTopUpsParams = {}
): Promise<AdminTopUpsResponse> {
  const query = new URLSearchParams()
  query.set('p', String(params.page ?? 1))
  query.set('page_size', String(params.page_size ?? 20))
  if (params.user_id) query.set('user_id', params.user_id)
  if (params.keyword) query.set('keyword', params.keyword)
  const res = await api.get(`/api/user/topup?${query.toString()}`)
  return res.data
}
