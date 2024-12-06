package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type IpApiResponseBody struct {
	Query   string
	Status  string
	Country string
}

type ApiSuccessResponseData struct {
	Country string `json:"country"`
}

type ApiErrorResponseData struct {
	Message string `json:"message"`
}

type IpCacheTableItem struct {
	id      int
	ip      string
	country string
}

type ApiHandler struct {
	db         *sql.DB
	logger     *logrus.Logger
	httpClient *http.Client
	config     *AppConfig
}

// Declare Prometheus metrics
var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of requests handled by the server",
		},
		[]string{"path", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)
	webserviceErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_webservice_errors_total",
			Help: "Errors inside the webservice when handling an http request",
		},
		[]string{"path", "error"},
	)
)

// Helper functions
func writeSuccessResponse(w http.ResponseWriter, body *ApiSuccessResponseData) {
	message, _ := json.Marshal(body)
	w.Write(message)
}

func writeApiError(w http.ResponseWriter, errorMessage string, errorStatusCode int) {
	message, _ := json.Marshal(ApiErrorResponseData{errorMessage})
	w.WriteHeader(errorStatusCode)
	w.Write(message)
}

func validateIp(ip *string) bool {
	return net.ParseIP(*ip) != nil
}

func (h *ApiHandler) ipLocationHandler(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues(r.URL.Path))
	defer timer.ObserveDuration()
	h.logger.Infof("Received request: Method=%s, URL=%s, Headers=%v", r.Method, r.URL.String(), r.Header)
	// X-Real-IP is the IP of the client
	ip := r.Header.Get("X-Real-IP")
	if !validateIp(&ip) {
		h.logger.Errorf("Bad IP address given, returning error.")
		statusCode := http.StatusBadRequest
		totalRequests.WithLabelValues(r.URL.Path, fmt.Sprint(statusCode)).Inc()
		writeApiError(w, "bad ip address", statusCode)
		return
	}

	h.logger.Infof("checking if the ip is in cache for ip: %s", ip)
	country, err := h.getIPCountryFromCache(ip)
	// If it was in cache, write the response and return
	if err == nil {
		h.logger.Infof("Ip %s was found in cache, returning the result.", ip)
		totalRequests.WithLabelValues(r.URL.Path, fmt.Sprint(http.StatusOK)).Inc()
		writeSuccessResponse(w, &ApiSuccessResponseData{*country})
		return
	}

	h.logger.Infof("Getting the country from web for ip: %s", ip)
	// If data was not in cache, get it from web
	country, err = h.getIPCountryFromWeb(ip)
	// if cannot get it from web, terminate the request and return error
	if err != nil {
		h.logger.Errorf("Cannot get the ip from web, got this error: %s", err)
		statusCode := http.StatusInternalServerError
		webserviceErrors.WithLabelValues(r.URL.Path, "web_fetch_error").Inc()
		totalRequests.WithLabelValues(r.URL.Path, fmt.Sprint(statusCode)).Inc()
		writeApiError(w, "internal error", statusCode)
		return
	}

	h.logger.Infof("Writing the data fetched from web to db.")
	err = h.writeIpCountryToCache(ip, *country)
	if err != nil {
		h.logger.Errorf("Cannot write the data to db, got this error: %s", err)
		webserviceErrors.WithLabelValues(r.URL.Path, "db_write_error").Inc()
	}
	totalRequests.WithLabelValues(r.URL.Path, fmt.Sprint(http.StatusOK)).Inc()
	writeSuccessResponse(w, &ApiSuccessResponseData{*country})
}

// getIpCountryFromCache reads the cache to check whether the requested information exists and if so, it will return that.
func (h *ApiHandler) getIPCountryFromCache(ip string) (*string, error) {
	query := fmt.Sprintf("SELECT id, ip, country FROM %s WHERE ip =$1;", h.config.DBTableName)
	h.logger.Infof("row query: %s", query)
	row := h.db.QueryRow(query, ip)
	var ipCacheTableItem IpCacheTableItem
	err := row.Scan(&ipCacheTableItem.id, &ipCacheTableItem.ip, &ipCacheTableItem.country)
	switch err {
	case sql.ErrNoRows:
		return nil, fmt.Errorf("getIpCountryFromCache %s: no such ip exist in table %s", ip, h.config.DBTableName)
	case nil:
		h.logger.Infof("Found row at db: {'id': %d, 'ip': %s, 'country': %s}", ipCacheTableItem.id, ipCacheTableItem.ip, ipCacheTableItem.country)
		return &ipCacheTableItem.country, nil
	default:
		err := fmt.Errorf("Bad state at database execution, cannot run query: %s", query)
		h.logger.Error(err)
		return nil, err
	}
}

// writeIpCountryToCache writes the data fetched externally to the cache table in the db.
func (h *ApiHandler) writeIpCountryToCache(ip string, country string) error {
	query := fmt.Sprintf("INSERT INTO %s (ip, country) VALUES ('%s', '%s');", h.config.DBTableName, ip, country)
	h.logger.Infof("Running insert query: %s", query)
	_, err := h.db.Exec(query)
	if err != nil {
		return fmt.Errorf("writeIpCountryToCache: %s", err)
	}
	return nil
}

// getIPCountryFromWeb gets the requested information from a public api and returns it to the app.
func (h *ApiHandler) getIPCountryFromWeb(ip string) (*string, error) {
	base_url := "http://ip-api.com"
	// sample request: http://ip-api.com/json/24.48.0.1
	requestUrl := fmt.Sprintf("%s/json/%s", base_url, ip)
	res, err := h.httpClient.Get(requestUrl)
	// Check if request was made ok
	if err != nil {
		log.Printf("Error when getting ip from: %s", base_url)
		return nil, err
	}
	// Check status code if not return error
	if res.StatusCode != http.StatusOK {
		log.Printf("IP service is not responding in expected way, status code received: %d", res.StatusCode)
		return nil, fmt.Errorf("Bad status code error")
	}
	// Close the body buffer
	defer res.Body.Close()
	// Read the buffer and create the response type
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Cannot read the response body.")
		return nil, fmt.Errorf("Response body unreadable error")
	}
	var data IpApiResponseBody
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Cannot unmarshall the response json to the IpApiResponseBody type.")
		return nil, err
	}
	return &data.Country, nil
}

func NewApiHandler(db *sql.DB, logger *logrus.Logger, client *http.Client, config *AppConfig) *ApiHandler {
	prometheus.MustRegister(totalRequests)
	prometheus.MustRegister(webserviceErrors)
	prometheus.MustRegister(requestDuration)
	return &ApiHandler{
		db, logger, client, config,
	}
}
