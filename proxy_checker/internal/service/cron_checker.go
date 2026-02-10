package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/moroshma/proxy_checker/proxy_checker/internal/models"
	"golang.org/x/net/proxy"
)

type ProxyCronRepositoryI interface {
	CreateTaskProxy(ctx context.Context, proxy []models.ProxyCheckServiceReq) (models.ProxyCheckServiceResponse, error)
	GetStatusProxy(ctx context.Context, checkID string) ([]models.ProxyResultServiceResponse, error)
	GetHistory(ctx context.Context) ([]models.HistoryItem, error)
	SelectWork(ctx context.Context) ([]models.Proxy, error)
	UpdateProxyMetric(ctx context.Context, proxyMetric models.ProxyMetric) error
	UpdateProxy(ctx context.Context, proxyMetric models.Proxy) error
}

type CroneChecker struct {
	repo    ProxyCronRepositoryI
	timeout time.Duration
}

func NewCroneChecker(repo ProxyCronRepositoryI, timeout time.Duration) *CroneChecker {
	return &CroneChecker{
		repo:    repo,
		timeout: timeout,
	}
}

func (r *CroneChecker) Run() {
	for {
		proxies, err := r.repo.SelectWork(context.Background())
		if err != nil {
			slog.Error(err.Error())
			time.Sleep(time.Second * 5)
			continue
		}

		if len(proxies) == 0 {
			time.Sleep(time.Second * 5)
			continue
		}

		jobs := make(chan models.Proxy, len(proxies))
		var wg sync.WaitGroup

		workers := 5_000
		if len(proxies) < workers {
			workers = len(proxies)
		}

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for p := range jobs {
					r.checkProxy(p)
				}
			}()
		}

		for _, p := range proxies {
			jobs <- p
		}
		close(jobs)

		wg.Wait()
		time.Sleep(time.Second * 5)
	}
}

// TODO: добавить в конфиг отмену контекста
func (r *CroneChecker) checkProxy(p models.Proxy) {
	ctx := context.Background()
	addr := net.JoinHostPort(p.IP, p.Port)

	var client *http.Client
	var speed int
	var ok bool

	switch p.Type {
	case "SOCKS5":
		client, speed, ok = r.trySocks5(addr)
	case "HTTP":
		client, speed, ok = r.tryHTTP(addr)
	default:
		ok = false
	}

	if !ok {
		slog.Error(fmt.Sprintf("proxy %s check failed for %s", p.Type, addr))
		r.repo.UpdateProxyMetric(ctx, models.ProxyMetric{
			ProxyMetricID: p.ProxyMetricID,
			Type:          p.Type,
			IsWork:        false,
			Speed:         0,
		})
		return
	}

	location, err := r.checkHttpLocation(p.IP)
	if err != nil {
		slog.Error(fmt.Sprintf("location error for %s: %v", addr, err))
	}

	realIP := p.IP
	realIPResp, err := client.Get("http://ip-api.com/json/")
	if err == nil {
		body, _ := io.ReadAll(realIPResp.Body)
		realIPResp.Body.Close()
		var loc models.Location
		if json.Unmarshal(body, &loc) == nil && loc.Query != "" {
			realIP = loc.Query
		}
	}

	city := fmt.Sprintf("%s, %s", location.Country, location.City)

	err = r.repo.UpdateProxy(ctx, models.Proxy{
		ProxyID: p.ProxyID,
		City:    city,
		RealIP:  realIP,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("update proxy error: %v", err))
	}

	err = r.repo.UpdateProxyMetric(ctx, models.ProxyMetric{
		ProxyMetricID: p.ProxyMetricID,
		Type:          p.Type,
		IsWork:        true,
		Speed:         speed,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("update proxy metric error: %v", err))
	}
}

func (r *CroneChecker) trySocks5(addr string) (*http.Client, int, bool) {
	dialer, err := proxy.SOCKS5("tcp", addr, nil, &net.Dialer{Timeout: r.timeout})
	if err != nil {
		return nil, 0, false
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, a string) (net.Conn, error) {
			return dialer.Dial(network, a)
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport, Timeout: r.timeout}

	speed, ok := r.measureSpeed(client)
	if !ok {
		return nil, 0, false
	}

	return client, speed, true
}
func (r *CroneChecker) tryHTTP(addr string) (*http.Client, int, bool) {
	proxyURL := &url.URL{Scheme: "http", Host: addr}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),

		// Важно: таймауты на этапы
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		// Если нужно, но лучше избегать:
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   r.timeout, // общий лимит
	}

	speed, ok := r.measureSpeed(client)
	if !ok {
		return nil, 0, false
	}

	return client, speed, true
}

func (r *CroneChecker) measureSpeed(client *http.Client) (int, bool) {
	start := time.Now()

	req, err := http.NewRequest("GET", "https://ip.pn/", nil)
	if err != nil {
		return 0, false
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)
	return int(time.Since(start).Milliseconds()), true
}

func (r *CroneChecker) checkHttpLocation(ip string) (models.Location, error) {
	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return models.Location{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Location{}, err
	}

	var location models.Location
	err = json.Unmarshal(body, &location)
	if err != nil {
		return models.Location{}, err
	}

	return location, nil
}
