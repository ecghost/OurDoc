package model

import (
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

// InsertUser 插入用户到单表 user（不再分片）
func InsertUser(u User) error {
	if u.ID == "" {
		return fmt.Errorf("empty user ID")
	}

	// 用户统一写入 DBOg1.user
	targetDB := db.DBOg1
	table := "user"

	_, err := targetDB.Exec(
		fmt.Sprintf("INSERT INTO %s (id, user_name, email, password) VALUES ($1, $2, $3, $4)", table),
		u.ID, u.UserName, u.Email, u.Password,
	)
	if err != nil {
		log.Printf("Insert user %v failed: %v", u, err)
	}
	return err
}

// QueryAllUsers 查询所有用户（单表 user）
func QueryAllUsers() ([]User, error) {
	users := []User{}

	rows, err := db.DBOg1.Query("SELECT id, user_name, email, password FROM \"user\"")
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

	return users, nil
}
