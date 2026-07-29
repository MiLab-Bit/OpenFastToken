package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	// 自动检测数据库路径
	execDir, _ := os.Executable()
	dbPath := filepath.Join(filepath.Dir(execDir), "FastToken.db")
	
	// 兼容旧版本
	if _, err := os.Stat("one-api.db"); err == nil {
		dbPath = "one-api.db"
	} else if _, err := os.Stat("FastToken.db"); err == nil {
		dbPath = "FastToken.db"
	}
	
	// 支持环境变量
	if envPath := os.Getenv("FASTTOKEN_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	fmt.Println("============================================================")
	fmt.Println("FastToken OAuth 精简 - 数据库迁移脚本")
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Printf("📂 数据库路径: %s\n\n", dbPath)

	// 检查数据库文件
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ 错误: 数据库文件不存在: %s\n", dbPath)
		os.Exit(1)
	}

	// 连接数据库
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// ... 其余迁移逻辑保持不变 ...
}
