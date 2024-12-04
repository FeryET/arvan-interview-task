package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
)

type ApiConfig struct {
	DBHost         string `env:"DB_HOST, default=localhost"`
	DBPort         int    `env:"DB_PORT, default=5432"`
	DBUser         string `env:"DB_USER, default=postgres"`
	DBPassword     string `env:"DB_PASSWORD, default=postgres"`
	DBName         string `env:"DB_NAME, default=db"`
	DBTableName    string `env:"DB_TABLE_NAME, default=ip_cache"`
	DBMaxOpenConns int    `env:"DB_MAX_OPEN_CONNS, default=1024"`
	DBMaxIdleConns int    `env:"DB_MAX_IDLE_CONNS, default=512"`
	DBMaxLifeTime  int    `env:"DB_MAX_LIFETIME_SECS, default=20"`
	DBMaxIdleTime  int    `env:"DB_MAX_IDLETIME_SECS, default=10"`
}
type IpApiResponseBody struct {
	Query   string
	Status  string
	Country string
}

type IpCacheTableItem struct {
	id      int
	ip      string
	country string
}

type ApiSuccessResponseData struct {
	Country string `json:"country"`
}

type ApiErrorResponseData struct {
	Message string `json:"message"`
}

type ApiHandler struct {
	db         *sql.DB
	logger     *logrus.Logger
	httpClient *http.Client
	config     *ApiConfig
}

func writeSuccessResponse(w *http.ResponseWriter, body *ApiSuccessResponseData) {
	message, _ := json.Marshal(body)
	(*w).Write(message)
}

func writeApiError(w *http.ResponseWriter, errorMessage string) {
	message, _ := json.Marshal(ApiErrorResponseData{errorMessage})
	(*w).WriteHeader(http.StatusInternalServerError)
	(*w).Write(message)
}

func validateIp(ip *string) bool {
	return net.ParseIP(*ip) != nil
}

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

func (h *ApiHandler) writeIpCountryToCache(ip string, country string) error {
	query := fmt.Sprintf("INSERT INTO %s (ip, country) VALUES ('%s', '%s');", h.config.DBTableName, ip, country)
	h.logger.Infof("Running insert query: %s", query)
	_, err := h.db.Exec(query)
	if err != nil {
		return fmt.Errorf("writeIpCountryToCache: %s", err)
	}
	return nil
}

func (h *ApiHandler) getIPCountryFromWeb(ip string) (*string, error) {
	base_url := "http://ip-api.com"
	// sample requeset: http://ip-api.com/json/24.48.0.1
	requestUrl := fmt.Sprintf("%s/json/%s", base_url, ip)
	res, err := h.httpClient.Get(requestUrl)
	// Check if request was made ok
	if err != nil {
		log.Printf("Error when getting ip from: %s", base_url)
		return nil, err
	}
	// Check status code if not return error
	if res.StatusCode != http.StatusOK {
		log.Printf("IP service is not responding in expected way, status code recieved: %d", res.StatusCode)
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

func (h *ApiHandler) ipLocationHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("Received request: Method=%s, URL=%s, Headers=%v", r.Method, r.URL.String(), r.Header)
	// X-Real-IP is the IP of the client
	ip := r.Header.Get("X-Real-IP")
	if !validateIp(&ip) {
		h.logger.Errorf("Bad IP address given, returning error.")
		writeApiError(&w, "bad ip address")
		return
	}
	h.logger.Infof("checking if the ip is in cache for ip: %s", ip)
	country, err := h.getIPCountryFromCache(ip)
	// If it was in cache, write the response and return
	if err == nil {
		h.logger.Infof("Ip %s was found in cache, returning the result.", ip)
		writeSuccessResponse(&w, &ApiSuccessResponseData{*country})
		return
	}
	h.logger.Infof("Getting the country from web for ip: %s", ip)
	// If data was not in cache, get it from web
	country, err = h.getIPCountryFromWeb(ip)
	// if cannot get it from web, terminate the request and return error
	if err != nil {
		h.logger.Errorf("Cannot get the ip from web, got this error: %s", err)
		writeApiError(&w, "internal error")
		return
	}
	// write it to db, then return the request
	h.logger.Infof("Writing the data fetched from web to db.")
	err = h.writeIpCountryToCache(ip, *country)
	if err != nil {
		h.logger.Errorf("Cannot write the data to db, got this error: %s", err)
	}
	writeSuccessResponse(&w, &ApiSuccessResponseData{*country})
}

// CreateDBConnection establishes and returns a new database connection using the ApiConfig struct.
func (config *ApiConfig) CreateDBConnection() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(time.Second * time.Duration(config.DBMaxIdleTime))
	db.SetConnMaxLifetime(time.Second * time.Duration(config.DBMaxLifeTime))
	db.SetMaxIdleConns(config.DBMaxIdleConns)
	db.SetMaxOpenConns(config.DBMaxOpenConns)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	// initialization
	logger := logrus.New()

	ctx := context.Background()
	var config ApiConfig
	if err := envconfig.Process(ctx, &config); err != nil {
		logger.Fatalf("Cannot create the config, error: %s", err)
	}

	db, dbErr := config.CreateDBConnection()
	if dbErr != nil {
		logger.Fatalf("Cannot create the database connection, error: %s", dbErr)
	}
	defer db.Close()

	httpClient := http.Client{Timeout: 10 * time.Second}
	defer httpClient.CloseIdleConnections()

	handler := ApiHandler{db, logger, &httpClient, &config}
	http.HandleFunc("/", handler.ipLocationHandler)

	http_err := http.ListenAndServe(":3333", nil)
	if http_err != nil {
		panic(http_err)
	}
}
