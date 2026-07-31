package model

// GetAllTokensAdmin 管理员导出用：返回全量 token（不含已软删除）。调用方需自行脱敏 Key。
func GetAllTokensAdmin(startIdx int, num int) (tokens []*Token, total int64, err error) {
	err = DB.Model(&Token{}).Where("deleted_at IS NULL").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = DB.Where("deleted_at IS NULL").Order("id desc").Limit(num).Offset(startIdx).Find(&tokens).Error
	return
}
