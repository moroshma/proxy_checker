package models

type ProxyCheckApiModelRes struct {
	ProxyAddress []string `json:"proxy_address"`
}
type ProxyCheckServiceReq struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type ProxyCheckServiceResponse struct {
	CheckID       string `json:"check_id"`
	ProxyMetricID string `json:"proxy_metric"`
}
