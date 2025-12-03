package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"my-gauss-app/model"
)

// HandleReadDataset 处理主键查询请求
// GET /api/dataset/read?dataset_name=user&main_key=123&goal_key=*
func HandleReadDataset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	datasetName := query.Get("dataset_name")
	mainKeyStr := query.Get("main_key")
	goalKey := query.Get("goal_key")

	if datasetName == "" || mainKeyStr == "" {
		http.Error(w, "Missing required parameters: dataset_name, main_key", http.StatusBadRequest)
		return
	}

	if goalKey == "" {
		goalKey = "*"
	}

	// 解析 main_key，可能是单个值或元组
	var mainKey interface{}

	// 尝试解析为 JSON 数组（元组）
	var keyArray []interface{}
	if err := json.Unmarshal([]byte(mainKeyStr), &keyArray); err == nil && len(keyArray) > 0 {
		mainKey = keyArray
	} else {
		// 否则作为字符串处理
		mainKey = mainKeyStr
	}

	result, err := model.ReadDataset(datasetName, mainKey, goalKey)
	if err != nil {
		log.Printf("ReadDataset failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"result": nil})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
}

// HandleReadDatasetCondition 处理条件查询请求
// GET /api/dataset/read_condition?dataset_name=user&key_name=email&key_value=test@example.com&goal_key=*
func HandleReadDatasetCondition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	datasetName := query.Get("dataset_name")
	keyName := query.Get("key_name")
	keyValueStr := query.Get("key_value")
	goalKey := query.Get("goal_key")
	fmt.Printf("Received request - dataset_name: %s, key_name: %s, key_value: %s, goal_key: %s\n",
		datasetName, keyName, keyValueStr, goalKey)

	if datasetName == "" || keyName == "" || keyValueStr == "" {
		http.Error(w, "Missing required parameters: dataset_name, key_name, key_value", http.StatusBadRequest)
		return
	}

	if goalKey == "" {
		goalKey = "*"
	}

	// URL 解码
	keyValue, err := url.QueryUnescape(keyValueStr)
	if err != nil {
		keyValue = keyValueStr
	}

	result, err := model.ReadDatasetCondition(datasetName, keyName, keyValue, goalKey)
	if err != nil {
		log.Printf("ReadDatasetCondition failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": nil,
			"debug": map[string]string{
				"dataset_name": datasetName,
				"key_name":     keyName,
				"key_value":    keyValueStr,
				"goal_key":     goalKey,
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
}

// HandleRemoveDatasetMainKey 删除某行数据集请求
// GET /api/dataset/read_json?dataset_name=user_table
func HandleRemoveDatasetMainKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DatasetName string      `json:"dataset_name"`
		MainKey     interface{} `json:"main_key"`
		MainValue   interface{} `json:"main_value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.DatasetName == "" || req.MainKey == nil || req.MainValue == nil {
		http.Error(w, "Missing required parameters: dataset_name, main_key, main_value", http.StatusBadRequest)
		return
	}

	err := model.RemoveDatasetMainKey(req.DatasetName, req.MainKey, req.MainValue)
	if err != nil {
		log.Printf("RemoveDatasetMainKey failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"msg":     "删除成功",
	})
}

// HandleInsertDataIntoDataset 处理插入数据请求
// POST /api/dataset/insert
// Body: {"dataset_name": "user", "data": {"id": "123", "user_name": "test", "email": "test@example.com", "password": "pass"}}
func HandleInsertDataIntoDataset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DatasetName string                 `json:"dataset_name"`
		Data        map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("Invalid JSON: %v", err)
		return
	}

	if req.DatasetName == "" {
		http.Error(w, "Missing dataset_name", http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		http.Error(w, "Missing data", http.StatusBadRequest)
		return
	}

	if err := model.InsertDataIntoDataset(req.DatasetName, req.Data); err != nil {
		log.Printf("InsertDataIntoDataset failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data inserted successfully"})
}

// HandleModifyDatasetCondition 处理修改数据请求
// POST /api/dataset/modify
// Body: {"dataset_name": "user", "key_name": "id", "key_value": "123", "goal_key": "email", "goal_value": "new@example.com"}
func HandleModifyDatasetCondition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DatasetName string      `json:"dataset_name"`
		KeyName     string      `json:"key_name"`
		KeyValue    interface{} `json:"key_value"`
		GoalKey     string      `json:"goal_key"`
		GoalValue   interface{} `json:"goal_value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("Invalid JSON: %v", err)
		return
	}

	if req.DatasetName == "" || req.KeyName == "" || req.GoalKey == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	modified, err := model.ModifyDatasetCondition(req.DatasetName, req.KeyName, req.KeyValue, req.GoalKey, req.GoalValue)
	if err != nil {
		log.Printf("ModifyDatasetCondition failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !modified {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"modified": false, "message": "No rows matched"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"modified": true, "message": "Data modified successfully"})
}

// HandleReadJSON 处理读取整个数据集请求
// GET /api/dataset/read_json?dataset_name=user_table
func HandleReadJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	datasetName := query.Get("dataset_name")

	if datasetName == "" {
		http.Error(w, "Missing required parameter: dataset_name", http.StatusBadRequest)
		return
	}

	data, err := model.ReadJSON(datasetName)
	if err != nil {
		log.Printf("ReadJSON failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleWriteJSON 处理写入整个数据集请求
// POST /api/dataset/write_json
// Body: {"dataset_name": "user_table", "data": [{"id": "1", "user_name": "test", ...}, ...]}
func HandleWriteJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DatasetName string                   `json:"dataset_name"`
		Data        []map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("Invalid JSON: %v", err)
		return
	}

	if req.DatasetName == "" {
		http.Error(w, "Missing dataset_name", http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		req.Data = []map[string]interface{}{} // 允许空数组
	}

	if err := model.WriteJSON(req.DatasetName, req.Data); err != nil {
		log.Printf("WriteJSON failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data written successfully"})
}
