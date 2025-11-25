package main

import (
	"fmt"
	"log"
	"net/http"

	"my-gauss-app/db"
	"my-gauss-app/handler"
)

func main() {
	db.InitDB()
	defer db.DBOg1.Close()
	defer db.DBOg2.Close()

	db.InitTables()

	// 原有的用户 API
	http.HandleFunc("/users", handler.HandleUsers)
	http.HandleFunc("/users/query", handler.HandleQueryUsers)

	// 对应 Python 的函数）
	http.HandleFunc("/api/dataset/read", handler.HandleReadDataset)
	http.HandleFunc("/api/dataset/read_condition", handler.HandleReadDatasetCondition)
	http.HandleFunc("/api/dataset/insert", handler.HandleInsertDataIntoDataset)
	http.HandleFunc("/api/dataset/modify", handler.HandleModifyDatasetCondition)
	http.HandleFunc("/api/dataset/read_json", handler.HandleReadJSON)
	http.HandleFunc("/api/dataset/write_json", handler.HandleWriteJSON)
	http.HandleFunc("/api/dataset/remove", handler.HandleRemoveDatasetMainKey)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
