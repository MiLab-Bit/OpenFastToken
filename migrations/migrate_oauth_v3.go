package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

func main() {
	// 自动检测数据库路径
	execDir, _ := os.Executable()
	dbPath := filepath.Join(filepath.Dir(execDir), "FastToken.db")
	
	// 兼容旧版本数据库文件名
	if _, err := os.Stat("one-api.db"); err == nil {
		dbPath = "one-api.db"
	} else if _, err := os.Stat("FastToken.db"); err == nil {
		dbPath = "FastToken.db"
	}
	
	// 支持环境变量指定
	if envPath := os.Getenv("FASTTOKEN_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	fmt.Println("============================================================")
	fmt.Println("FastToken OAuth 精简 - 数据库迁移脚本")
	fmt.Println("============================================================")
	fmt.Println()

	// 检查数据库文件
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ 错误: 数据库文件不存在: %s\n", dbPath)
		fmt.Println("提示: 可以通过环境变量 FASTTOKEN_DB_PATH 指定数据库路径")
		os.Exit(1)
	}

	fileInfo, _ := os.Stat(dbPath)
	fmt.Printf("✅ 找到数据库文件: %s\n", dbPath)
	fmt.Printf("   文件大小: %d 字节\n", fileInfo.Size())
	fmt.Println()

	// 初始化数据库连接
	common.InitDB()
	defer common.CloseDB()

	fmt.Println("✅ 数据库连接成功")
	fmt.Println()

	// 检查要删除的表是否存在
	fmt.Println("🔍 检查要删除的表...")
	var existingTables []string

	result := common.DB.Raw(`
		SELECT name FROM sqlite_master 
		WHERE type='table' 
		AND name IN ('custom_oauth_providers', 'user_oauth_bindings')
	`).Scan(&existingTables)

	if result.Error != nil {
		fmt.Printf("❌ 查询失败: %v\n", result.Error)
		os.Exit(1)
	}

	if len(existingTables) == 0 {
		fmt.Println("ℹ️  要删除的表不存在，无需迁移")
		fmt.Println("   - custom_oauth_providers: 不存在")
		fmt.Println("   - user_oauth_bindings: 不存在")
		fmt.Println()
		fmt.Println("============================================================")
		fmt.Println("✅ 迁移完成（无需操作）")
		fmt.Println("============================================================")
		return
	}

	fmt.Printf("📋 找到 %d 个要删除的表:\n", len(existingTables))
	for _, table := range existingTables {
		fmt.Printf("   - %s\n", table)
	}
	fmt.Println()

	// 执行删除
	fmt.Println("🗑️  开始删除表...")

	for _, table := range existingTables {
		fmt.Printf("   删除 %s...\n", table)
		result := common.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if result.Error != nil {
			fmt.Printf("❌ 删除 %s 失败: %v\n", table, result.Error)
			continue
		}
		fmt.Printf("   ✅ %s 已删除\n", table)
	}

	fmt.Println()
	fmt.Println("✅ 表删除成功！")
	fmt.Println()

	// 验证结果
	fmt.Println("🔍 验证删除结果...")
	var remaining []string

	common.DB.Raw(`
		SELECT name FROM sqlite_master 
		WHERE type='table' 
		AND name IN ('custom_oauth_providers', 'user_oauth_bindings')
	`).Scan(&remaining)

	if len(remaining) == 0 {
		fmt.Println("✅ 验证成功：表已被完全删除")
	} else {
		fmt.Printf("⚠️  警告：仍有残留表: %v\n", remaining)
	}

	// 显示所有表
	fmt.Println()
	fmt.Println("📊 当前数据库中的所有表:")
	var tables []string
	common.DB.Raw("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").Scan(&tables)

	for i, table := range tables {
		fmt.Printf("   %d. %s\n", i+1, table)
	}

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println("✅ 迁移完成！")
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Println("📝 后续步骤:")
	fmt.Println("   1. 重启 FastToken 服务")
	fmt.Println("   2. 测试登录功能（邮箱、手机号、微信、GitHub）")
	fmt.Println("   3. 检查用户数据是否正常")
	fmt.Println()
}
