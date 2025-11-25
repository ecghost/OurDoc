package model

import (
	"database/sql"
	"fmt"
	"log"
	"my-gauss-app/db"
	"strings"
)

// ReadDataset 主键查询，根据主键查询整行数据或特定字段
// dataset_name: 表名 (user, document, permission, content)
// main_key: 主键值，可以是单个值或元组 (room_id, user_id)
// goal_key: 目标字段名，如果是 "*" 则返回整行数据
func ReadDataset(datasetName string, mainKey interface{}, goalKey string) (interface{}, error) {
	var targetDB *sql.DB
	var table string
	var query string
	var args []interface{}

	// 设置目标表和数据库
	switch datasetName {
	case "user":
		if keyStr, ok := mainKey.(string); ok && keyStr != "*" {
			// 用户分片逻辑
			lastChar := keyStr[len(keyStr)-1]
			if lastChar >= '0' && lastChar <= '9' {
				lastDigit := int(lastChar - '0')
				if lastDigit%2 == 0 {
					targetDB = db.DBOg2
					table = "user_1"
				} else {
					targetDB = db.DBOg1
					table = "user_0"
				}
			} else {
				targetDB = db.DBOg1
				table = "user_0"
			}
		} else {
			// mainKey == "*" 或空，读取所有分片
			targetDB = db.DBOg1 // 可以先从 DBOg1 读取 user_0
			table = "user_0"
		}

		if goalKey == "*" {
			query = fmt.Sprintf("SELECT id, user_name, email, password FROM %s", table)
		} else {
			query = fmt.Sprintf("SELECT %s FROM %s", goalKey, table)
		}
		if mainKey != "*" {
			args = []interface{}{mainKey}
			query += " WHERE id = $1"
		}

	case "document", "content":
		targetDB = db.DBOg1
		table = datasetName
		var primaryKey string
		if datasetName == "document" {
			primaryKey = "room_id"
		} else {
			primaryKey = "room_id"
		}

		if mainKey == "*" {
			if goalKey == "*" {
				if datasetName == "document" {
					query = fmt.Sprintf("SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM %s", table)
				} else {
					query = fmt.Sprintf("SELECT room_id, content FROM %s", table)
				}
			} else {
				query = fmt.Sprintf("SELECT %s FROM %s", goalKey, table)
			}
		} else {
			if goalKey == "*" {
				if datasetName == "document" {
					query = fmt.Sprintf("SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM %s WHERE %s = $1", table, primaryKey)
				} else {
					query = fmt.Sprintf("SELECT room_id, content FROM %s WHERE %s = $1", table, primaryKey)
				}
			} else {
				query = fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", goalKey, table, primaryKey)
			}
			args = []interface{}{mainKey}
		}

	case "permission":
		targetDB = db.DBOg1
		table = "permission"

		if mainKey == "*" {
			if goalKey == "*" {
				query = fmt.Sprintf("SELECT room_id, user_id, permission FROM %s", table)
			} else {
				query = fmt.Sprintf("SELECT %s FROM %s", goalKey, table)
			}
		} else if keyTuple, ok := mainKey.([]interface{}); ok && len(keyTuple) == 2 {
			if goalKey == "*" {
				query = fmt.Sprintf("SELECT room_id, user_id, permission FROM %s WHERE room_id = $1 AND user_id = $2", table)
			} else {
				query = fmt.Sprintf("SELECT %s FROM %s WHERE room_id = $1 AND user_id = $2", goalKey, table)
			}
			args = []interface{}{keyTuple[0], keyTuple[1]}
		} else {
			return nil, fmt.Errorf("invalid key for permission table")
		}

	default:
		return nil, fmt.Errorf("unknown dataset: %s", datasetName)
	}

	rows, err := targetDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		rowMap, err := scanRowToMap(rows, datasetName)
		if err != nil {
			return nil, err
		}
		results = append(results, rowMap)
	}

	if len(results) == 0 {
		return nil, nil
	}

	// 如果 mainKey 是单个值并只查询单列，返回单值
	if goalKey != "*" && mainKey != "*" {
		return results[0][goalKey], nil
	}

	return results, nil
}

// ReadDatasetCondition 条件查询，根据某个字段的值查询
func ReadDatasetCondition(datasetName string, keyName string, keyValue interface{}, goalKey string) (interface{}, error) {
	// 用户表需要查询所有分片
	if strings.HasPrefix(datasetName, "user") {
		var results []map[string]interface{}

		// 查询所有分片
		for _, t := range []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "user_0"},
			{db.DBOg2, "user_1"},
		} {
			var query string
			if goalKey == "*" {
				query = fmt.Sprintf("SELECT id, user_name, email, password FROM %s WHERE %s = $1", t.table, keyName)
			} else {
				query = fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", goalKey, t.table, keyName)
			}

			rows, err := t.db.Query(query, keyValue)
			if err != nil {
				return nil, fmt.Errorf("query %s failed: %v", t.table, err)
			}
			defer rows.Close()

			for rows.Next() {
				if goalKey == "*" {
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
				} else {
					var result interface{}
					if err := rows.Scan(&result); err != nil {
						return nil, fmt.Errorf("scan failed: %v", err)
					}
					// 如果只查询单个字段，返回第一个匹配的值
					if len(results) == 0 {
						return result, nil
					}
				}
			}
		}

		if len(results) == 0 {
			return nil, nil
		}
		if goalKey == "*" {
			return results[0], nil // 返回第一个匹配的结果
		}
		return nil, nil
	}

	// 其他表（document, permission, content）
	var targetDB *sql.DB
	var table string
	targetDB = db.DBOg1
	table = datasetName

	var query string
	if goalKey == "*" {
		if datasetName == "document" {
			query = fmt.Sprintf("SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM %s WHERE %s = $1", table, keyName)
		} else if datasetName == "permission" {
			query = fmt.Sprintf("SELECT room_id, user_id, permission FROM %s WHERE %s = $1", table, keyName)
		} else if datasetName == "content" {
			query = fmt.Sprintf("SELECT room_id, content FROM %s WHERE %s = $1", table, keyName)
		} else {
			return nil, fmt.Errorf("unknown dataset: %s", datasetName)
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
		return nil, nil // 没有找到数据
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

// InsertDataIntoDataset 插入整行数据
func InsertDataIntoDataset(datasetName string, data map[string]interface{}) error {
	var targetDB *sql.DB
	var table string
	var columns []string
	var placeholders []string
	var values []interface{}

	// 确定目标数据库和表名
	if strings.HasPrefix(datasetName, "user") {
		// 用户表需要分片
		id, ok := data["id"].(string)
		if !ok {
			return fmt.Errorf("user table requires 'id' field")
		}

		if len(id) > 0 {
			lastChar := id[len(id)-1]
			if lastChar >= '0' && lastChar <= '9' {
				lastDigit := int(lastChar - '0')
				if lastDigit%2 == 0 {
					targetDB = db.DBOg2
					table = "user_1"
				} else {
					targetDB = db.DBOg1
					table = "user_0"
				}
			} else {
				targetDB = db.DBOg1
				table = "user_0"
			}
		} else {
			targetDB = db.DBOg1
			table = "user_0"
		}

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
		targetDB = db.DBOg1
		table = datasetName

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

// RemoveDatasetMainKey 删除整行数据
func RemoveDatasetMainKey(datasetName string, mainKey interface{}, mainValue interface{}) error {
	var targetDB *sql.DB
	var table string
	var whereClause string
	var values []interface{}

	// 处理用户表分片
	if strings.HasPrefix(datasetName, "user") {
		idVal := ""
		if v, ok := mainValue.(string); ok {
			idVal = v
		} else {
			return fmt.Errorf("user table requires string id for deletion")
		}

		lastChar := idVal[len(idVal)-1]
		if lastChar >= '0' && lastChar <= '9' {
			lastDigit := int(lastChar - '0')
			if lastDigit%2 == 0 {
				targetDB = db.DBOg2
				table = "user_1"
			} else {
				targetDB = db.DBOg1
				table = "user_0"
			}
		} else {
			targetDB = db.DBOg1
			table = "user_0"
		}

		whereClause = "id = $1"
		values = append(values, idVal)

	} else {
		// 其他表
		targetDB = db.DBOg1
		table = datasetName

		// 支持单主键或复合主键
		switch k := mainKey.(type) {
		case string:
			whereClause = fmt.Sprintf("%s = $1", k)
			values = append(values, mainValue)
		case []interface{}: // 改成 []interface{} 而不是 []string
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
		log.Printf("Delete from %s failed: %v", table, err)
		return fmt.Errorf("delete failed: %v", err)
	}

	return nil
}

// ModifyDatasetCondition 根据条件修改某个字段的值
func ModifyDatasetCondition(datasetName string, keyName string, keyValue interface{}, goalKey string, goalValue interface{}) (bool, error) {
	// 用户表需要处理分片
	if strings.HasPrefix(datasetName, "user") {
		// 如果 keyName 是 "id"，可以根据 id 确定分片
		if keyName == "id" {
			keyStr, ok := keyValue.(string)
			if !ok {
				return false, fmt.Errorf("user id must be string")
			}

			// 确定分片
			var targetDB *sql.DB
			var table string
			if len(keyStr) > 0 {
				lastChar := keyStr[len(keyStr)-1]
				if lastChar >= '0' && lastChar <= '9' {
					lastDigit := int(lastChar - '0')
					if lastDigit%2 == 0 {
						targetDB = db.DBOg2
						table = "user_1"
					} else {
						targetDB = db.DBOg1
						table = "user_0"
					}
				} else {
					targetDB = db.DBOg1
					table = "user_0"
				}
			} else {
				targetDB = db.DBOg1
				table = "user_0"
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
		} else {
			// 如果 keyName 不是 "id"，需要在所有分片中查找并更新
			totalRowsAffected := int64(0)
			for _, t := range []struct {
				db    *sql.DB
				table string
			}{
				{db.DBOg1, "user_0"},
				{db.DBOg2, "user_1"},
			} {
				query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", t.table, goalKey, keyName)
				result, err := t.db.Exec(query, goalValue, keyValue)
				if err != nil {
					return false, fmt.Errorf("update %s failed: %v", t.table, err)
				}

				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return false, fmt.Errorf("get rows affected failed: %v", err)
				}
				totalRowsAffected += rowsAffected
			}

			return totalRowsAffected > 0, nil
		}
	}

	// 其他表（document, permission, content）
	var targetDB *sql.DB
	var table string
	targetDB = db.DBOg1
	table = datasetName

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
		result["room_id"] = roomID
		result["room_name"] = roomName
		result["owner_user_id"] = owner_user_id
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
//   - "user" 或 "user_table" -> 用户表（合并所有分片）
//   - "document" 或 "user_room_table" -> document 表
//   - "permission" 或 "room_permission_table" -> permission 表
//   - "content" 或 "room_content_table" -> content 表
func ReadJSON(datasetName string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 映射数据集名称到实际表名
	if datasetName == "user_table" || datasetName == "user" {
		// 用户表需要查询所有分片并合并
		for _, t := range []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "user_0"},
			{db.DBOg2, "user_1"},
		} {
			query := fmt.Sprintf("SELECT id, user_name, email, password FROM %s", t.table)
			rows, err := t.db.Query(query)
			if err != nil {
				return nil, fmt.Errorf("query %s failed: %v", t.table, err)
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
		}
		return results, nil

	} else if datasetName == "user_room_table" || datasetName == "document" {
		// document 表
		query := "SELECT room_id, room_name, create_time, overall_permission, owner_user_id FROM document"
		rows, err := db.DBOg1.Query(query)
		if err != nil {
			return nil, fmt.Errorf("query document failed: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var roomID, roomName, owner_user_id string
			var createTime sql.NullTime
			var overallPermission sql.NullInt64
			if err := rows.Scan(&roomID, &roomName, &createTime, &overallPermission, &owner_user_id); err != nil {
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
		return results, nil

	} else if datasetName == "room_permission_table" || datasetName == "permission" {
		// permission 表
		query := "SELECT room_id, user_id, permission FROM permission"
		rows, err := db.DBOg1.Query(query)
		if err != nil {
			return nil, fmt.Errorf("query permission failed: %v", err)
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
		return results, nil

	} else if datasetName == "room_content_table" || datasetName == "content" {
		// content 表
		query := "SELECT room_id, content FROM content"
		rows, err := db.DBOg1.Query(query)
		if err != nil {
			return nil, fmt.Errorf("query content failed: %v", err)
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
		// 用户表需要分片处理，先清空所有分片
		for _, t := range []struct {
			db    *sql.DB
			table string
		}{
			{db.DBOg1, "user_0"},
			{db.DBOg2, "user_1"},
		} {
			if _, err := t.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", t.table)); err != nil {
				return fmt.Errorf("truncate %s failed: %v", t.table, err)
			}
		}

		// 插入数据到对应的分片
		columns = []string{"id", "user_name", "email", "password"}
		for _, row := range data {
			id, ok := row["id"].(string)
			if !ok {
				log.Printf("Skip row without valid id: %v", row)
				continue
			}

			// 确定分片
			var targetDB *sql.DB
			var table string
			if len(id) > 0 {
				lastChar := id[len(id)-1]
				if lastChar >= '0' && lastChar <= '9' {
					lastDigit := int(lastChar - '0')
					if lastDigit%2 == 0 {
						targetDB = db.DBOg2
						table = "user_1"
					} else {
						targetDB = db.DBOg1
						table = "user_0"
					}
				} else {
					targetDB = db.DBOg1
					table = "user_0"
				}
			} else {
				targetDB = db.DBOg1
				table = "user_0"
			}

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
		targetDB = db.DBOg1
		table = "document"
		columns = []string{"room_id", "room_name", "create_time", "overall_permission", "owner_user_id"}

		// 清空表
		if _, err := targetDB.Exec("TRUNCATE TABLE document"); err != nil {
			return fmt.Errorf("truncate document failed: %v", err)
		}

		// 插入数据
		for _, row := range data {
			values := []interface{}{
				row["room_id"],
				row["room_name"],
				row["create_time"],
				row["overall_permission"],
				row["owner_user_id"],
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

	} else if datasetName == "room_permission_table" || datasetName == "permission" {
		targetDB = db.DBOg1
		table = "permission"
		columns = []string{"room_id", "user_id", "permission"}

		// 清空表
		if _, err := targetDB.Exec("TRUNCATE TABLE permission"); err != nil {
			return fmt.Errorf("truncate permission failed: %v", err)
		}

		// 插入数据
		for _, row := range data {
			values := []interface{}{
				row["room_id"],
				row["user_id"],
				row["permission"],
			}
			placeholders := []string{"$1", "$2", "$3"}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				table,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "))

			if _, err := targetDB.Exec(query, values...); err != nil {
				return fmt.Errorf("insert into %s failed: %v", table, err)
			}
		}
		return nil

	} else if datasetName == "room_content_table" || datasetName == "content" {
		targetDB = db.DBOg1
		table = "content"
		columns = []string{"room_id", "content"}

		// 清空表
		if _, err := targetDB.Exec("TRUNCATE TABLE content"); err != nil {
			return fmt.Errorf("truncate content failed: %v", err)
		}

		// 插入数据
		for _, row := range data {
			values := []interface{}{
				row["room_id"],
				row["content"],
			}
			placeholders := []string{"$1", "$2"}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				table,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "))

			if _, err := targetDB.Exec(query, values...); err != nil {
				return fmt.Errorf("insert into %s failed: %v", table, err)
			}
		}
		return nil

	} else {
		return fmt.Errorf("unknown dataset: %s", datasetName)
	}
}
