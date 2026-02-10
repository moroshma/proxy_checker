package models

import "time"

type ProxyResultServiceReq struct {
	TaskUUID string `json:"task_uuid"`
}

type ProxyResultServiceResponse struct {
	CheckID string `json:"check_id"`
	Type    string `json:"type"`
	IsWork  bool   `json:"is_work"`
	Speed   int    `json:"speed"`
	Status  string `json:"status"`
	City    string `json:"city"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	RealIP  string `json:"real_ip"`
}

type HistoryItem struct {
	CheckID    string    `json:"check_id"`
	CreateAt   time.Time `json:"create_at"`
	ProxyCount int       `json:"proxy_count"`
}
