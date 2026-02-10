package service

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/moroshma/proxy_checker/proxy_checker/internal/models"
)

type ProxyApiRepositoryI interface {
	CreateTaskProxy(ctx context.Context, proxy []models.ProxyCheckServiceReq) (models.ProxyCheckServiceResponse, error)
	GetStatusProxy(ctx context.Context, checkID string) ([]models.ProxyResultServiceResponse, error)
	GetHistory(ctx context.Context) ([]models.HistoryItem, error)
	SelectWork(ctx context.Context) ([]models.Proxy, error)
	UpdateProxyMetric(ctx context.Context, proxyMetric models.ProxyMetric) error
	UpdateProxy(ctx context.Context, proxyMetric models.Proxy) error
}

type ProxyService struct {
	repo ProxyApiRepositoryI
}

func NewResumeService(repo ProxyApiRepositoryI) *ProxyService {
	return &ProxyService{
		repo: repo,
	}
}

func (r *ProxyService) CreateTaskProxy(ctx context.Context, proxy models.ProxyCheckApiModelRes) (models.ProxyCheckServiceResponse, error) {
	pr := make([]models.ProxyCheckServiceReq, 0, len(proxy.ProxyAddress))
	for _, v := range proxy.ProxyAddress {
		host, port, err := net.SplitHostPort(v)
		if err != nil {
			return models.ProxyCheckServiceResponse{}, fmt.Errorf("incorrect format ip:port - %v", err)
		}
		if net.ParseIP(host) == nil {
			return models.ProxyCheckServiceResponse{}, fmt.Errorf("incorrect IP-address: %s", host)
		}
		p, err := strconv.Atoi(port)
		if err != nil || p < 0 || p > 65535 {
			return models.ProxyCheckServiceResponse{}, fmt.Errorf("incorrect port: %s", port)
		}

		pr = append(pr, models.ProxyCheckServiceReq{
			IP:   host,
			Port: p,
		})
	}

	id, err := r.repo.CreateTaskProxy(ctx, pr)
	if err != nil {
		return models.ProxyCheckServiceResponse{}, err
	}

	return models.ProxyCheckServiceResponse{
		CheckID: id.CheckID,
	}, nil
}

func (r *ProxyService) GetHistory(ctx context.Context) ([]models.HistoryItem, error) {
	return r.repo.GetHistory(ctx)
}

func (r *ProxyService) GetStatusProxy(ctx context.Context, proxy models.ProxyResultServiceReq) ([]models.ProxyResultServiceResponse, error) {
	proxyList, err := r.repo.GetStatusProxy(ctx, proxy.TaskUUID)
	if err != nil {
		return nil, err
	}

	return proxyList, nil
}
