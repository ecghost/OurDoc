// db/init_tables.go
package db

import (
	"database/sql"
	"fmt"
	"log"
)

// InitTables 创建用户、房间(document)、权限和内容表
func InitTables() {
	// 用户表：不再分片，统一使用 DBOg1.user
	userSQL := `
    CREATE TABLE IF NOT EXISTS "user" (
        id VARCHAR(64) PRIMARY KEY,
        user_name VARCHAR(64),
        email VARCHAR(100),
        password VARCHAR(256)
    );`
	if _, err := DBOg1.Exec(userSQL); err != nil {
		log.Fatalf("Create user table failed: %v", err)
	}

	// room(document)、permission、content 表：
	// 以 hash(room_id) 在两个实例上分片，这里采用两个分片：_0 落在 og1，_1 落在 og2。
	type shardTable struct {
		db     *sql.DB
		suffix string
	}

	roomShards := []shardTable{
		{DBOg1, "0"},
		{DBOg2, "1"},
	}

	// document_<shard>
	for _, s := range roomShards {
		sqlStr := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS document_%s (
            room_id VARCHAR(64) PRIMARY KEY,
            room_name VARCHAR(128),
            create_time TIMESTAMP,
            overall_permission INT,
			owner_user_id VARCHAR(64)
        );`, s.suffix)
		if _, err := s.db.Exec(sqlStr); err != nil {
			log.Fatalf("Create table document_%s failed: %v", s.suffix, err)
		}
	}

	// permission_<shard>
	for _, s := range roomShards {
		sqlStr := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS permission_%s (
            room_id VARCHAR(64),
            user_id VARCHAR(64),
            permission INT,
            PRIMARY KEY(room_id, user_id)
        );`, s.suffix)
		if _, err := s.db.Exec(sqlStr); err != nil {
			log.Fatalf("Create table permission_%s failed: %v", s.suffix, err)
		}
	}

	// content_<shard>
	for _, s := range roomShards {
		sqlStr := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS content_%s (
            room_id VARCHAR(64) PRIMARY KEY,
            content TEXT
        );`, s.suffix)
		if _, err := s.db.Exec(sqlStr); err != nil {
			log.Fatalf("Create table content_%s failed: %v", s.suffix, err)
		}
	}

	log.Println("All tables (user + sharded room/permission/content) created successfully")
}
