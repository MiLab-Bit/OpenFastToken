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
