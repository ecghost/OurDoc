// db/init_tables.go
package db

import (
	"database/sql"
	"fmt"
	"log"
)

// InitTables 创建用户、文档、权限和内容表
func InitTables() {
	// 用户表分片（简单: id % 2）
	userTables := []struct {
		db    *sql.DB
		table string
	}{
		{DBOg1, "user_0"},
		{DBOg2, "user_1"},
	}

	for _, t := range userTables {
		sqlStr := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id VARCHAR(64) PRIMARY KEY,
            user_name VARCHAR(64),
            email VARCHAR(100),
            password VARCHAR(256)
        );`, t.table)
		if _, err := t.db.Exec(sqlStr); err != nil {
			log.Fatalf("Create table %s failed: %v", t.table, err)
		}
	}

	// 文档表（单表）
	docSQL := `
    CREATE TABLE IF NOT EXISTS document (
        room_id VARCHAR(64) PRIMARY KEY,
        room_name VARCHAR(128),
        create_time TIMESTAMP,
        overall_permission INT,
		owner_user_id VARCHAR(64)
    );`
	if _, err := DBOg1.Exec(docSQL); err != nil {
		log.Fatalf("Create document table failed: %v", err)
	}

	// 权限表
	permSQL := `
    CREATE TABLE IF NOT EXISTS permission (
        room_id VARCHAR(64),
        user_id VARCHAR(64),
        permission INT,
        PRIMARY KEY(room_id, user_id)
    );`
	if _, err := DBOg1.Exec(permSQL); err != nil {
		log.Fatalf("Create permission table failed: %v", err)
	}

	// 内容表
	contentSQL := `
    CREATE TABLE IF NOT EXISTS content (
        room_id VARCHAR(64) PRIMARY KEY,
        content TEXT
    );`
	if _, err := DBOg1.Exec(contentSQL); err != nil {
		log.Fatalf("Create content table failed: %v", err)
	}

	log.Println("All tables created successfully")
}
