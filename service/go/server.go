package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	/* Initialization */

	// Init logging
	logger := logrus.New()

	// Init config
	config, err := NewAppConfig()
	if err != nil {
		logger.Fatalf("Cannot create the config, error: %s", err)
	}

	// Init db
	db, dbErr := config.CreateDBConnection()
	if dbErr != nil {
		logger.Fatalf("Cannot create the database connection, error: %s", dbErr)
	}
	defer db.Close()

	// Init http client
	httpClient := http.Client{Timeout: 10 * time.Second}
	defer httpClient.CloseIdleConnections()

	/* Webservice */
	// Create the HTTP web server and listen on the desired port
	handler := NewApiHandler(db, logger, &httpClient, config)
	http.HandleFunc("/", handler.ipLocationHandler)
	// Expose the Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	httpErr := http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil)
	if httpErr != nil {
		logger.Fatalf("Failed to start server: %s", httpErr)
	}
}
