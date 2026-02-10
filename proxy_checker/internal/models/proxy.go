package models

import (
	"time"

	"github.com/google/uuid"
)

type Proxy struct {
	ProxyID       uuid.UUID
	CheckID       uuid.UUID `db:"id"`
	IP            string    `json:"ip"`
	Port          string    `json:"port"`
	City          string    `json:"city"`
	RealIP        string    `json:"real_ip"`
	ProxyMetricID uuid.UUID
	Type          string
}

type ProxyMetric struct {
	ProxyMetricID uuid.UUID
	CheckID       uuid.UUID `db:"id"`
	Type          string    `json:"type"`
	IsWork        bool      `json:"is_work"`
	Speed         int       `json:"speed"`
	Status        string    `json:"status"`
}

type CheckTable struct {
	CheckID  uuid.UUID `json:"check_id"`
	CreateAt time.Time `json:"create_at"`
}
