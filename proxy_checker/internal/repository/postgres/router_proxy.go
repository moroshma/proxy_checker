package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/models"
)

type ProxyRepository struct {
	db *pgxpool.Pool
}

func NewProxyRepository(db *pgxpool.Pool) *ProxyRepository {
	return &ProxyRepository{db: db}
}

func (p *ProxyRepository) CreateTaskProxy(ctx context.Context, proxy []models.ProxyCheckServiceReq) (models.ProxyCheckServiceResponse, error) {
	var idTask string
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return models.ProxyCheckServiceResponse{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, createTaskInTableId).Scan(&idTask)
	if err != nil {
		return models.ProxyCheckServiceResponse{}, err
	}

	for _, prx := range proxy {
		var proxyID string
		err := tx.QueryRow(ctx, createTaskInProxy, idTask, prx.IP, prx.Port).Scan(&proxyID)
		if err != nil {
			return models.ProxyCheckServiceResponse{}, err
		}

		for _, proxyType := range []string{"SOCKS5", "HTTP"} {
			_, err = tx.Exec(ctx, createTaskInProxyMetric, idTask, proxyID, proxyType)
			if err != nil {
				return models.ProxyCheckServiceResponse{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.ProxyCheckServiceResponse{}, err
	}

	return models.ProxyCheckServiceResponse{
		CheckID: idTask,
	}, nil
}

func (p *ProxyRepository) GetStatusProxy(ctx context.Context, checkID string) ([]models.ProxyResultServiceResponse, error) {
	rows, err := p.db.Query(ctx, getStatusProxy, checkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ProxyResultServiceResponse

	for rows.Next() {
		var res models.ProxyResultServiceResponse
		err := rows.Scan(&res.CheckID, &res.IP, &res.Port, &res.City, &res.RealIP, &res.Type, &res.IsWork, &res.Speed, &res.Status)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *ProxyRepository) SelectWork(ctx context.Context) ([]models.Proxy, error) {
	rows, err := p.db.Query(ctx, selectTaskInWork)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Proxy

	for rows.Next() {
		var res models.Proxy
		err := rows.Scan(&res.ProxyID, &res.CheckID, &res.IP, &res.Port, &res.ProxyMetricID, &res.Type)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *ProxyRepository) GetHistory(ctx context.Context) ([]models.HistoryItem, error) {
	rows, err := p.db.Query(ctx, getHistory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.HistoryItem
	for rows.Next() {
		var res models.HistoryItem
		err := rows.Scan(&res.CheckID, &res.CreateAt, &res.ProxyCount)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *ProxyRepository) UpdateProxyMetric(ctx context.Context, proxyMetric models.ProxyMetric) error {
	_, err := p.db.Exec(ctx, updateProxyMetric, proxyMetric.Type, proxyMetric.IsWork, proxyMetric.Speed, proxyMetric.ProxyMetricID)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProxyRepository) UpdateProxy(ctx context.Context, proxyMetric models.Proxy) error {
	_, err := p.db.Exec(ctx, updateProxy, proxyMetric.City, proxyMetric.RealIP, proxyMetric.ProxyID)
	if err != nil {
		return err
	}
	return nil
}
