package model

import (
	"database/sql"
	"fmt"
	"log"
	"my-gauss-app/db"
	"strings"
)

// hashRoomID 计算 room_id 的简单 hash，用于决定落在哪个分片（0 或 1）
func hashRoomID(roomID string) int {
	if roomID == "" {
		return 0
	}
	h := 0
	for i := 0; i < len(roomID); i++ {
		h = (h + int(roomID[i])) % 2
	}
	return h
}

// getRoomShard 根据 room_id 和逻辑表名（document/permission/content）返回对应的 DB 和物理表名
func getRoomShard(baseTable string, roomID string) (*sql.DB, string, error) {
	shard := hashRoomID(roomID)
	switch shard {
	case 0:
		return db.DBOg1, fmt.Sprintf("%s_0", baseTable), nil
	case 1:
		return db.DBOg2, fmt.Sprintf("%s_1", baseTable), nil
	default:
		return nil, "", fmt.Errorf("invalid shard for room_id %s", roomID)
	}
}

// ReadDataset 主键查询，根据主键查询整行数据或特定字段
// dataset_name: 表名 (user, document, permission, content)
// main_key: 主键值，可以是单个值或元组 (room_id, user_id)
// goal_key: 目标字段名，如果是 "*" 则返回整行数据
func queryAllRows(db *sql.DB, query string, datasetName string, goalKey string) (interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []map[string]interface{}
	for rows.Next() {
		row, err := scanRowToMap(rows, datasetName)
		if err != nil {
			return nil, err
		}
		all = append(all, row)
	}
	return all, nil
}

func ReadDataset(datasetName string, mainKey interface{}, goalKey string) (interface{}, error) {
	var targetDB *sql.DB
	var table string
	var query string
	var args []interface{}

	if mainKey == "*" {
		switch datasetName {

		case "user":
			// 用户表不分片
			targetDB = db.DBOg1
			table = "\"user\""

			if goalKey == "*" {
				query = fmt.Sprintf("SELECT id, user_name, email, password FROM %s", table)
			} else {
				query = fmt.Sprintf("SELECT %s FROM %s", goalKey, table)
			}

			return queryAllRows(targetDB, query, datasetName, goalKey)

		case "permission":
			// 遍历所有分片
			shards := []struct {
				db    *sql.DB
				table string
			}{
				{db.DBOg1, "permission_0"},
				{db.DBOg2, "permission_1"},
			}

			var all []map[string]interface{}
			for _, s := range shards {
				var q string
				if goalKey == "*" {
					q = fmt.Sprintf("SELECT room_id, user_id, permission FROM %s", s.table)
				} else {
					q = fmt.Sprintf("SELECT %s FROM %s", goalKey, s.table)
				}

				rows, err := s.db.Query(q)
				if err != nil {
					return nil, err
				}
				defer rows.Close()

				for rows.Next() {
					row, err := scanRowToMap(rows, datasetName)
					if err != nil {
						return nil, err
					}
					all = append(all, row)
				}
			}
			return all, nil

		case "document":
			fallthrough
		case "content":
			// document/content 两份分片
			shards := []struct {
				db    *sql.DB
				table string
			}{
				{db.DBOg1, datasetName + "_0"},
				{db.DBOg2, datasetName + "_1"},
			}

			var all []map[string]interface{}
			for _, s := range shards {
				var q string
				if goalKey == "*" {
					if datasetName == "document" {
						q = fmt.Sprintf("SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM %s", s.table)
					} else {
						q = fmt.Sprintf("SELECT room_id, content FROM %s", s.table)
					}
				} else {
					q = fmt.Sprintf("SELECT %s FROM %s", goalKey, s.table)
				}

				rows, err := s.db.Query(q)
				if err != nil {
					return nil, err
				}
				defer rows.Close()

				for rows.Next() {
					row, err := scanRowToMap(rows, datasetName)
					if err != nil {
						return nil, err
					}
					all = append(all, row)
				}
			}
			return all, nil

		default:
			return nil, fmt.Errorf("unknown dataset: %s", datasetName)
		}
	}

	// -----------------------------
	// 下面是你原来的逻辑 (mainKey != "*")
	// -----------------------------

	if strings.HasPrefix(datasetName, "user") {
		// ...（此处完全保留你的原代码）
		keyStr, ok := mainKey.(string)
		if !ok {
			return nil, fmt.Errorf("user table requires string key")
		}
		targetDB = db.DBOg1
		table = "\"user\""

		if goalKey == "*" {
			query = fmt.Sprintf("SELECT id, user_name, email, password FROM %s WHERE id = $1", table)
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", goalKey, table)
		}
		args = []interface{}{keyStr}

	} else if datasetName == "permission" {
		// ...（完全保留你的原代码）
		// 略，为节省篇幅

	} else {
		// document, content 处理
		// ...（完全保留你的原代码）
	}

	// 执行查询
	rows, err := targetDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	if goalKey == "*" {
		return scanRowToMap(rows, datasetName)
	} else {
		var result interface{}
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		return result, nil
	}
}

// ReadDatasetCondition 条件查询，根据某个字段的值查询
func ReadDatasetCondition(datasetName string, keyName string, keyValue interface{}, goalKey string) (interface{}, error) {
	// 用户表：不分片，直接在 user 上查询
	if strings.HasPrefix(datasetName, "user") {
		table := "\"user\""
		var query string
		if goalKey == "*" {
			query = fmt.Sprintf("SELECT id, user_name, email, password FROM %s WHERE %s = $1", table, keyName)
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", goalKey, table, keyName)
		}

		rows, err := db.DBOg1.Query(query, keyValue)
		if err != nil {
			return nil, fmt.Errorf("query user failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, nil
		}

		if goalKey == "*" {
			return scanRowToMap(rows, "user")
		}

		var result interface{}
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		return result, nil
	}

	// 其他表（document(room)、permission、content）：按 room_id hash 分片
	if datasetName != "document" && datasetName != "permission" && datasetName != "content" {
		return nil, fmt.Errorf("unknown dataset: %s", datasetName)
	}

	// 如果按 room_id 作为条件，可以根据 room_id 精确定位分片
	if keyName == "room_id" {
		roomID, ok := keyValue.(string)
		if !ok {
			return nil, fmt.Errorf("room_id must be string")
		}
		targetDB, table, err := getRoomShard(datasetName, roomID)
		if err != nil {
			return nil, err
		}

		var query string
		if goalKey == "*" {
			if datasetName == "document" {
				query = fmt.Sprintf("SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM %s WHERE %s = $1", table, keyName)
			} else if datasetName == "permission" {
				query = fmt.Sprintf("SELECT room_id, user_id, permission FROM %s WHERE %s = $1", table, keyName)
			} else if datasetName == "content" {
				query = fmt.Sprintf("SELECT room_id, content FROM %s WHERE %s = $1", table, keyName)
			}
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", goalKey, table, keyName)
		}

		rows, err := targetDB.Query(query, keyValue)
		if err != nil {
			return nil, fmt.Errorf("query failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, nil
		}

		if goalKey == "*" {
			return scanRowToMap(rows, datasetName)
		}

		var result interface{}
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		return result, nil
	}

	// 其他条件（如 permission.user_id 等），需要扫描所有分片
	shards := []struct {
		db    *sql.DB
		table string
	}{
		{db.DBOg1, fmt.Sprintf("%s_0", datasetName)},
		{db.DBOg2, fmt.Sprintf("%s_1", datasetName)},
	}

	for _, s := range shards {
		var query string
		if goalKey == "*" {
			if datasetName == "document" {
				query = fmt.Sprintf("SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM %s WHERE %s = $1", s.table, keyName)
			} else if datasetName == "permission" {
				query = fmt.Sprintf("SELECT room_id, user_id, permission FROM %s WHERE %s = $1", s.table, keyName)
			} else if datasetName == "content" {
				query = fmt.Sprintf("SELECT room_id, content FROM %s WHERE %s = $1", s.table, keyName)
			}
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", goalKey, s.table, keyName)
		}

		rows, err := s.db.Query(query, keyValue)
		if err != nil {
			return nil, fmt.Errorf("query %s failed: %v", s.table, err)
		}
		defer rows.Close()

		if rows.Next() {
			if goalKey == "*" {
				return scanRowToMap(rows, datasetName)
			}
			var result interface{}
			if err := rows.Scan(&result); err != nil {
				return nil, fmt.Errorf("scan failed: %v", err)
			}
			return result, nil
		}
	}

	return nil, nil
}

// InsertDataIntoDataset 插入整行数据
func InsertDataIntoDataset(datasetName string, data map[string]interface{}) error {
	var targetDB *sql.DB
	var table string
	var columns []string
	var placeholders []string
	var values []interface{}

	// 确定目标数据库和表名
	if strings.HasPrefix(datasetName, "user") {
		// 用户表不分片
		targetDB = db.DBOg1
		table = "\"user\""

		columns = []string{"id", "user_name", "email", "password"}
		for i, col := range columns {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			val, ok := data[col]
			if !ok {
				val = ""
			}
			values = append(values, val)
		}

	} else {
		// room(document)、permission、content：按 room_id hash 分片
		roomIDVal, ok := data["room_id"]
		if !ok {
			return fmt.Errorf("%s requires 'room_id' field", datasetName)
		}
		roomID, ok := roomIDVal.(string)
		if !ok {
			return fmt.Errorf("room_id must be string")
		}

		var err error
		targetDB, table, err = getRoomShard(datasetName, roomID)
		if err != nil {
			return err
		}

		if datasetName == "document" {
			columns = []string{"room_id", "room_name", "create_time", "overall_permission", "owner_user_id"}
		} else if datasetName == "permission" {
			columns = []string{"room_id", "user_id", "permission"}
		} else if datasetName == "content" {
			columns = []string{"room_id", "content"}
		} else {
			return fmt.Errorf("unknown dataset: %s", datasetName)
		}

		for i, col := range columns {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			val, ok := data[col]
			if !ok {
				val = nil
			}
			values = append(values, val)
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	_, err := targetDB.Exec(query, values...)
	if err != nil {
		log.Printf("Insert into %s failed: %v", table, err)
		return fmt.Errorf("insert failed: %v", err)
	}

	return nil
}

// ModifyDatasetCondition 根据条件修改某个字段的值
func ModifyDatasetCondition(datasetName string, keyName string, keyValue interface{}, goalKey string, goalValue interface{}) (bool, error) {
	// 用户表：单表 user
	if strings.HasPrefix(datasetName, "user") {
		table := "\"user\""
		query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", table, goalKey, keyName)
		result, err := db.DBOg1.Exec(query, goalValue, keyValue)
		if err != nil {
			return false, fmt.Errorf("update failed: %v", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return false, fmt.Errorf("get rows affected failed: %v", err)
		}

		return rowsAffected > 0, nil
	}

	// 其他表（document, permission, content）：按 room_id hash 分片
	if datasetName != "document" && datasetName != "permission" && datasetName != "content" {
		return false, fmt.Errorf("unknown dataset: %s", datasetName)
	}

	// 如果按 room_id 更新，则可以定位到单个分片
	if keyName == "room_id" {
		roomID, ok := keyValue.(string)
		if !ok {
			return false, fmt.Errorf("room_id must be string")
		}
		targetDB, table, err := getRoomShard(datasetName, roomID)
		if err != nil {
			return false, err
		}

		query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", table, goalKey, keyName)
		result, err := targetDB.Exec(query, goalValue, keyValue)
		if err != nil {
			return false, fmt.Errorf("update failed: %v", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return false, fmt.Errorf("get rows affected failed: %v", err)
		}

		return rowsAffected > 0, nil
	}

	// 其他条件需要更新所有分片
	shards := []struct {
		db    *sql.DB
		table string
	}{
		{db.DBOg1, fmt.Sprintf("%s_0", datasetName)},
		{db.DBOg2, fmt.Sprintf("%s_1", datasetName)},
	}

	totalRows := int64(0)
	for _, s := range shards {
		query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", s.table, goalKey, keyName)
		result, err := s.db.Exec(query, goalValue, keyValue)
		if err != nil {
			return false, fmt.Errorf("update %s failed: %v", s.table, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return false, fmt.Errorf("get rows affected failed: %v", err)
		}
		totalRows += rowsAffected
	}

	return totalRows > 0, nil
}

// scanRowToMap 将查询结果扫描到 map
func scanRowToMap(rows *sql.Rows, datasetName string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if datasetName == "user" || strings.HasPrefix(datasetName, "user") {
		var id, userName, email, password string
		if err := rows.Scan(&id, &userName, &email, &password); err != nil {
			return nil, err
		}
		result["id"] = id
		result["user_name"] = userName
		result["email"] = email
		result["password"] = password
	} else if datasetName == "document" {
		var roomID, roomName, owner_user_id string
		var createTime sql.NullTime
		var overallPermission sql.NullInt64
		if err := rows.Scan(&roomID, &roomName, &createTime, &overallPermission, &owner_user_id); err != nil {
			return nil, err
		}
		result["owner_user_id"] = owner_user_id
		result["room_id"] = roomID
		result["room_name"] = roomName
		if createTime.Valid {
			result["create_time"] = createTime.Time
		}
		if overallPermission.Valid {
			result["overall_permission"] = overallPermission.Int64
		}
	} else if datasetName == "permission" {
		var roomID, userID string
		var permission sql.NullInt64
		if err := rows.Scan(&roomID, &userID, &permission); err != nil {
			return nil, err
		}
		result["room_id"] = roomID
		result["user_id"] = userID
		if permission.Valid {
			result["permission"] = permission.Int64
		}
	} else if datasetName == "content" {
		var roomID, content string
		if err := rows.Scan(&roomID, &content); err != nil {
			return nil, err
		}
		result["room_id"] = roomID
		result["content"] = content
	}

	return result, nil
}

// ReadJSON 读取整个数据集（表）的所有数据
// dataset_name 支持：
//   - "user" 或 "user_table" -> 用户表（单表）
//   - "document" 或 "user_room_table" -> document 分片表
//   - "permission" 或 "room_permission_table" -> permission 分片表
//   - "content" 或 "room_content_table" -> content 分片表
func ReadJSON(datasetName string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 映射数据集名称到实际表名
	if datasetName == "user_table" || datasetName == "user" {
		// 用户表单表
		query := "SELECT id, user_name, email, password FROM \"user\""
		rows, err := db.DBOg1.Query(query)
		if err != nil {
			return nil, fmt.Errorf("query user failed: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id, userName, email, password string
			if err := rows.Scan(&id, &userName, &email, &password); err != nil {
				return nil, fmt.Errorf("scan failed: %v", err)
			}
			results = append(results, map[string]interface{}{
				"id":        id,
				"user_name": userName,
				"email":     email,
				"password":  password,
			})
		}
		return results, nil

	} else if datasetName == "user_room_table" || datasetName == "document" {
		// document 分片表
		shards := []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "document_0"},
			{db.DBOg2, "document_1"},
		}

		for _, s := range shards {
			query := fmt.Sprintf("SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM %s", s.table)
			rows, err := s.db.Query(query)
			if err != nil {
				return nil, fmt.Errorf("query %s failed: %v", s.table, err)
			}
			defer rows.Close()

			for rows.Next() {
				var roomID, roomName, owner_user_id string
				var createTime sql.NullTime
				var overallPermission sql.NullInt64
				if err := rows.Scan(&roomID, &roomName, &createTime, &overallPermission); err != nil {
					return nil, fmt.Errorf("scan failed: %v", err)
				}
				row := map[string]interface{}{
					"room_id":       roomID,
					"room_name":     roomName,
					"owner_user_id": owner_user_id,
				}
				if createTime.Valid {
					row["create_time"] = createTime.Time
				}
				if overallPermission.Valid {
					row["overall_permission"] = overallPermission.Int64
				}
				results = append(results, row)
			}
		}
		return results, nil

	} else if datasetName == "room_permission_table" || datasetName == "permission" {
		// permission 分片表
		shards := []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "permission_0"},
			{db.DBOg2, "permission_1"},
		}

		for _, s := range shards {
			query := fmt.Sprintf("SELECT room_id, user_id, permission FROM %s", s.table)
			rows, err := s.db.Query(query)
			if err != nil {
				return nil, fmt.Errorf("query %s failed: %v", s.table, err)
			}
			defer rows.Close()

			for rows.Next() {
				var roomID, userID string
				var permission sql.NullInt64
				if err := rows.Scan(&roomID, &userID, &permission); err != nil {
					return nil, fmt.Errorf("scan failed: %v", err)
				}
				row := map[string]interface{}{
					"room_id": roomID,
					"user_id": userID,
				}
				if permission.Valid {
					row["permission"] = permission.Int64
				}
				results = append(results, row)
			}
		}
		return results, nil

	} else if datasetName == "room_content_table" || datasetName == "content" {
		// content 分片表
		shards := []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "content_0"},
			{db.DBOg2, "content_1"},
		}

		for _, s := range shards {
			query := fmt.Sprintf("SELECT room_id, content FROM %s", s.table)
			rows, err := s.db.Query(query)
			if err != nil {
				return nil, fmt.Errorf("query %s failed: %v", s.table, err)
			}
			defer rows.Close()

			for rows.Next() {
				var roomID, content string
				if err := rows.Scan(&roomID, &content); err != nil {
					return nil, fmt.Errorf("scan failed: %v", err)
				}
				results = append(results, map[string]interface{}{
					"room_id": roomID,
					"content": content,
				})
			}
		}
		return results, nil

	} else {
		return nil, fmt.Errorf("unknown dataset: %s", datasetName)
	}
}

// WriteJSON 写入整个数据集（表）的数据
// 先清空表，然后插入新数据
// dataset_name 支持同 ReadJSON
func WriteJSON(datasetName string, data []map[string]interface{}) error {
	// 开始事务
	var targetDB *sql.DB
	var table string
	var columns []string

	// 确定目标数据库和表名
	if datasetName == "user_table" || datasetName == "user" {
		// 用户表单表，直接清空 user 然后写入
		targetDB = db.DBOg1
		table = "\"user\""
		columns = []string{"id", "user_name", "email", "password"}

		if _, err := targetDB.Exec("TRUNCATE TABLE \"user\""); err != nil {
			return fmt.Errorf("truncate user failed: %v", err)
		}

		for _, row := range data {
			values := []interface{}{
				row["id"],
				row["user_name"],
				row["email"],
				row["password"],
			}
			placeholders := []string{"$1", "$2", "$3", "$4"}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				table,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "))

			if _, err := targetDB.Exec(query, values...); err != nil {
				return fmt.Errorf("insert into %s failed: %v", table, err)
			}
		}
		return nil

	} else if datasetName == "user_room_table" || datasetName == "document" {
		// document 分片表，按 room_id hash 分片写入
		columns = []string{"room_id", "room_name", "create_time", "overall_permission", "owner_user_id"}

		// 先清空所有分片
		for _, t := range []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "document_0"},
			{db.DBOg2, "document_1"},
		} {
			if _, err := t.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", t.table)); err != nil {
				return fmt.Errorf("truncate %s failed: %v", t.table, err)
			}
		}
		columns = []string{"room_id", "room_name", "create_time", "overall_permission", "owner_user_id"}

		// 插入数据
		for _, row := range data {
			roomID, ok := row["room_id"].(string)
			if !ok {
				log.Printf("Skip row without valid room_id: %v", row)
				continue
			}

			shardDB, shardTable, err := getRoomShard("document", roomID)
			if err != nil {
				return err
			}

			values := []interface{}{
				row["room_id"],
				row["room_name"],
				row["create_time"],
				row["overall_permission"],
				row["owner_user_id"],
			}
			placeholders := []string{"$1", "$2", "$3", "$4"}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				shardTable,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "))

			if _, err := shardDB.Exec(query, values...); err != nil {
				return fmt.Errorf("insert into %s failed: %v", shardTable, err)
			}
		}
		return nil

	} else if datasetName == "room_permission_table" || datasetName == "permission" {
		columns = []string{"room_id", "user_id", "permission"}

		// 清空所有分片
		for _, t := range []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "permission_0"},
			{db.DBOg2, "permission_1"},
		} {
			if _, err := t.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", t.table)); err != nil {
				return fmt.Errorf("truncate %s failed: %v", t.table, err)
			}
		}

		// 插入数据（按 room_id 分片）
		for _, row := range data {
			roomID, ok := row["room_id"].(string)
			if !ok {
				log.Printf("Skip row without valid room_id: %v", row)
				continue
			}

			shardDB, shardTable, err := getRoomShard("permission", roomID)
			if err != nil {
				return err
			}

			values := []interface{}{
				row["room_id"],
				row["user_id"],
				row["permission"],
			}
			placeholders := []string{"$1", "$2", "$3"}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				shardTable,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "))

			if _, err := shardDB.Exec(query, values...); err != nil {
				return fmt.Errorf("insert into %s failed: %v", shardTable, err)
			}
		}
		return nil

	} else if datasetName == "room_content_table" || datasetName == "content" {
		columns = []string{"room_id", "content"}

		// 清空所有分片
		for _, t := range []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "content_0"},
			{db.DBOg2, "content_1"},
		} {
			if _, err := t.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", t.table)); err != nil {
				return fmt.Errorf("truncate %s failed: %v", t.table, err)
			}
		}

		// 插入数据（按 room_id 分片）
		for _, row := range data {
			roomID, ok := row["room_id"].(string)
			if !ok {
				log.Printf("Skip row without valid room_id: %v", row)
				continue
			}

			shardDB, shardTable, err := getRoomShard("content", roomID)
			if err != nil {
				return err
			}

			values := []interface{}{
				row["room_id"],
				row["content"],
			}
			placeholders := []string{"$1", "$2"}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				shardTable,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "))

			if _, err := shardDB.Exec(query, values...); err != nil {
				return fmt.Errorf("insert into %s failed: %v", shardTable, err)
			}
		}
		return nil

	} else {
		return fmt.Errorf("unknown dataset: %s", datasetName)
	}
}

func RemoveDatasetMainKey(datasetName string, mainKey interface{}, mainValue interface{}) error {
	var targetDB *sql.DB
	var table string
	var whereClause string
	var values []interface{}

	if strings.HasPrefix(datasetName, "user") {

		idStr, ok := mainValue.(string)
		if !ok {
			return fmt.Errorf("user table requires string id for deletion")
		}

		targetDB = db.DBOg1
		table = "\"user\""

		whereClause = "id = $1"
		values = append(values, idStr)

	} else if datasetName == "permission" {
		switch k := mainKey.(type) {
		case []interface{}:
			// 输入了两个 key: [room_id, user_id]
			if len(k) != 2 {
				return fmt.Errorf("permission delete requires 2-element key or single room_id")
			}
			vals, ok := mainValue.([]interface{})
			if !ok || len(vals) != 2 {
				return fmt.Errorf("permission delete requires 2-element value for [room_id, user_id]")
			}
			roomID, ok := vals[0].(string)
			if !ok {
				return fmt.Errorf("room_id must be string")
			}
			var err error
			targetDB, table, err = getRoomShard("permission", roomID)
			if err != nil {
				return err
			}
			whereClause = "room_id = $1 AND user_id = $2"
			values = append(values, vals[0], vals[1])

		case string:
			// 输入单个 room_id
			roomID, ok := mainValue.(string)
			if !ok {
				return fmt.Errorf("room_id must be string")
			}
			var err error
			targetDB, table, err = getRoomShard("permission", roomID)
			if err != nil {
				return err
			}
			whereClause = "room_id = $1"
			values = append(values, roomID)

		default:
			return fmt.Errorf("invalid key type for permission table")
		}
	} else if datasetName == "document" || datasetName == "content" {

		roomID, ok := mainValue.(string)
		if !ok {
			return fmt.Errorf("%s requires string room_id", datasetName)
		}

		var err error
		targetDB, table, err = getRoomShard(datasetName, roomID)
		if err != nil {
			return err
		}

		whereClause = "room_id = $1"
		values = append(values, roomID)

	} else {
		targetDB = db.DBOg1
		table = datasetName

		switch k := mainKey.(type) {
		case string:
			whereClause = fmt.Sprintf("%s = $1", k)
			values = append(values, mainValue)

		case []interface{}:
			vals, ok := mainValue.([]interface{})
			if !ok || len(k) != len(vals) {
				return fmt.Errorf("mainKey and mainValue length mismatch")
			}

			conds := make([]string, len(k))
			for i := range k {
				keyStr, ok := k[i].(string)
				if !ok {
					return fmt.Errorf("mainKey element is not string")
				}
				conds[i] = fmt.Sprintf("%s = $%d", keyStr, i+1)
				values = append(values, vals[i])
			}
			whereClause = strings.Join(conds, " AND ")

		default:
			return fmt.Errorf("unsupported mainKey type")
		}
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereClause)
	_, err := targetDB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("delete failed: %v", err)
	}

	return nil
}
