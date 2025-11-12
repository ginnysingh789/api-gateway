package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api-gateway/internal/circuit"
	"api-gateway/internal/service"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	registry       *service.Registry
	loadBalancer   *service.LoadBalancer
	breakerManager *circuit.BreakerManager
	logger         *logger.Logger
}

func NewProxyHandler(
	registry *service.Registry,
	lb *service.LoadBalancer,
	bm *circuit.BreakerManager,
	log *logger.Logger,
) *ProxyHandler {
	return &ProxyHandler{
		registry:       registry,
		loadBalancer:   lb,
		breakerManager: bm,
		logger:         log,
	}
}

func (p *ProxyHandler) ProxyRequest(c *gin.Context) {
	// Extract service name from path
	path := c.Param("path")
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)

	if len(parts) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request path")
		return
	}

	serviceName := parts[0]
	remainingPath := ""
	if len(parts) > 1 {
		remainingPath = "/" + parts[1]
	}

	// Get service from registry
	svc, err := p.registry.Get(serviceName)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Service not found")
		return
	}

	// Get target URL using load balancer
	targetURL, err := p.loadBalancer.RoundRobin(svc)
	if err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "No available instances")
		return
	}

	// Execute request through circuit breaker
	breaker := p.breakerManager.GetBreaker(serviceName)
	result, err := breaker.Execute(func() (interface{}, error) {
		return p.forwardRequest(c, targetURL, remainingPath)
	})

	if err != nil {
		p.logger.Errorw("Circuit breaker error",
			"service", serviceName,
			"error", err,
		)
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Service temporarily unavailable")
		return
	}

	response := result.(*ProxyResponse)

	// Copy headers
	for key, values := range response.Headers {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Send response
	c.Data(response.StatusCode, response.ContentType, response.Body)
}

type ProxyResponse struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	ContentType string
}

func (p *ProxyHandler) forwardRequest(c *gin.Context, targetURL, path string) (*ProxyResponse, error) {
	// Build target URL
	fullURL, err := url.Parse(targetURL + path)
	if err != nil {
		return nil, err
	}

	// Copy query parameters
	fullURL.RawQuery = c.Request.URL.RawQuery

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	// Create new request
	req, err := http.NewRequest(c.Request.Method, fullURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Copy headers (exclude hop-by-hop headers)
	for key, values := range c.Request.Header {
		if !isHopByHopHeader(key) {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// Add forwarding headers
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", c.Request.Proto)
	req.Header.Set("X-Forwarded-Host", c.Request.Host)

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ProxyResponse{
		StatusCode:  resp.StatusCode,
		Headers:     resp.Header,
		Body:        respBody,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, h := range hopByHopHeaders {
		if strings.EqualFold(header, h) {
			return true
		}
	}
	return false
}

func (p *ProxyHandler) ListServices(c *gin.Context) {
	services := p.registry.List()
	utils.SuccessResponse(c, http.StatusOK, "Services retrieved successfully", services)
}

func (p *ProxyHandler) RegisterService(c *gin.Context) {
	var req struct {
		Name      string   `json:"name" binding:"required"`
		URLs      []string `json:"urls" binding:"required"`
		HealthURL string   `json:"health_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	p.registry.Register(req.Name, req.URLs, req.HealthURL)
	utils.SuccessResponse(c, http.StatusCreated, "Service registered successfully", nil)
}

func (p *ProxyHandler) UnregisterService(c *gin.Context) {
	name := c.Param("name")

	if err := p.registry.Unregister(name); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Service unregistered successfully", nil)
}
