package model

import (
	"database/sql"
	"fmt"
	"log"
	"my-gauss-app/db"
)

type User struct {
	ID       string `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"` // 建议存 hash
}

// InsertUser 插入用户到分片表
func InsertUser(u User) error {
	var targetDB *sql.DB
	var table string

	if u.ID == "" {
		return fmt.Errorf("empty user ID")
	}

	// 获取 ID 的最后一位数字作为分片依据
	lastChar := u.ID[len(u.ID)-1]
	if lastChar < '0' || lastChar > '9' {
		log.Printf("User ID's last character is not a digit, default to shard 0: %s", u.ID)
		lastChar = '1' // 默认分到 user_0
	}
	lastDigit := int(lastChar - '0')

	if lastDigit%2 == 0 {
		targetDB = db.DBOg2
		table = "user_1"
	} else {
		targetDB = db.DBOg1
		table = "user_0"
	}

	_, err := targetDB.Exec(
		fmt.Sprintf("INSERT INTO %s (id, user_name, email, password) VALUES ($1, $2, $3, $4)", table),
		u.ID, u.UserName, u.Email, u.Password,
	)
	if err != nil {
		log.Printf("Insert user %v failed: %v", u, err)
	}
	return err
}

// QueryAllUsers 查询所有分片表
func QueryAllUsers() ([]User, error) {
	users := []User{}
	for _, t := range []struct {
		db    *sql.DB
		table string
	}{
		{db.DBOg1, "user_0"},
		{db.DBOg2, "user_1"},
	} {
		rows, err := t.db.Query(fmt.Sprintf("SELECT id, user_name, email, password FROM %s", t.table))
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.UserName, &u.Email, &u.Password); err != nil {
				return nil, err
			}
			users = append(users, u)
		}
	}
	return users, nil
}
