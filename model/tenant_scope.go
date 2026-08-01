package model

import "gorm.io/gorm"

// TenantScope returns a GORM Scope that filters by tenant_id.
// If tenantId is nil or 0, no filter is applied (admin/global query).
func TenantScope(tenantId *int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantId == nil || *tenantId == 0 {
			return db
		}
		return db.Where("tenant_id = ?", *tenantId)
	}
}

// TenantWhereClause generates a WHERE clause for tenant_id filtering.
func TenantWhereClause(tenantId *int64) string {
	if tenantId == nil || *tenantId == 0 {
		return ""
	}
	return "tenant_id = ?"
}

func TenantWhereArgs(tenantId *int64) []interface{} {
	if tenantId == nil || *tenantId == 0 {
		return nil
	}
	return []interface{}{*tenantId}
}

// OwnerOrTenantWhere 返回 Phase 1 租户可见性谓词。
// 语义：本人拥有的资源 OR 所属企业的共享资源。
// 个人用户 entId=0 时后半恒假，等价于 user_id = ?（向后兼容）。
// prefix 用于多表查询时限定列名，如 "logs."；单表传 ""。
//
// 注意：返回值自带最外层括号。与其他条件用 AND 组合时不要拆括号，
// 否则 SQL 运算符优先级会让 AND 先与右侧结合，造成跨租户越权放行。
// 占位符顺序固定为 (userId, entId)。
func OwnerOrTenantWhere(prefix string) string {
	return "(" + prefix + "user_id = ? OR (" + prefix + "tenant_id = ? AND " + prefix + "tenant_id != 0))"
}
