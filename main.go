package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/dewkul/prom-lolminer-exporter/schema"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultEndpoint = ":8080"
const namespace = "lolminer"

var enableDebug = false
var endpoint = defaultEndpoint

// var metricsEndpoint = ""

func main() {
	fmt.Printf("%s version %s by %s.\n", appName, appVersion, appAuthor)

	parseCliArgs()
	if enableDebug {
		fmt.Printf("[DEBUG] Debug mode enabled.\n")
	}

	if err := runServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
}

func parseCliArgs() {
	flag.BoolVar(&enableDebug, "debug", false, "Show debug messages.")
	flag.StringVar(&endpoint, "endpoint", defaultEndpoint, "The address-port endpoint to bind to.")

	// Exits on error
	flag.Parse()
}

func runServer() error {
	fmt.Printf("Listening on %s.\n", endpoint)
	var mainServeMux http.ServeMux
	mainServeMux.HandleFunc("/", handleOtherRequest)
	mainServeMux.HandleFunc("/metrics", handleScrapeRequest)
	if err := http.ListenAndServe(endpoint, &mainServeMux); err != nil {
		return fmt.Errorf("error while running main http server: %s", err)
	}
	return nil
}

func handleOtherRequest(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "%s version %s by %s.\n\n", appName, appVersion, appAuthor)
	fmt.Fprintf(response, "Usage: /metrics?target=<target>.\n")
}

func handleScrapeRequest(response http.ResponseWriter, request *http.Request) {
	if enableDebug {
		fmt.Printf("[DEBUG] Request: %s\n", request.RemoteAddr)
	}

	// Get and parse target
	targetURL := parseTargetURL(response, request)
	if targetURL == nil {
		return
	}

	// Scrape target and parse data
	data := scrapeTarget(response, targetURL)
	if data == nil {
		return
	}

	// Build registry with data
	registry := buildRegistry(response, data)
	if registry == nil {
		return
	}

	// Delegare final handling to Prometheus
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(response, request)
}

// Returns the target URL if successful and nil if not.
func parseTargetURL(response http.ResponseWriter, request *http.Request) *url.URL {
	var rawTarget string
	if values, ok := request.URL.Query()["target"]; ok && len(values) > 0 && values[0] != "" {
		rawTarget = values[0]
	} else {
		http.Error(response, "400 - Missing target.\n", 400)
		return nil
	}
	if !strings.HasPrefix(rawTarget, "http://") && !strings.HasPrefix(rawTarget, "https://") {
		rawTarget = "http://" + rawTarget
	}
	targetURL, targetURLErr := url.ParseRequestURI(rawTarget)
	if targetURLErr != nil {
		message := fmt.Sprintf("400 - Invalid target: %s\n", targetURLErr)
		http.Error(response, message, 400)
		return nil
	}
	return targetURL
}

// Scrapes the target and returns the parsed data if successful or nil if not.
func scrapeTarget(response http.ResponseWriter, targetURL *url.URL) *schema.LolMinerMetric {
	// Scrape
	scrapeRequest, scrapeRequestErr := http.NewRequest("GET", targetURL.String(), nil)
	if scrapeRequestErr != nil {
		// if enableDebug {
		fmt.Printf("[ERROR] Failed to make request to scrape target:\n%v", scrapeRequestErr)
		// }
		message := fmt.Sprintf("500 - Failed to scrape target: %s\n", scrapeRequestErr)
		http.Error(response, message, 500)
		return nil
	}
	scrapeClient := http.Client{}
	scrapeResponse, scrapeResponseErr := scrapeClient.Do(scrapeRequest)
	if scrapeResponseErr != nil {
		// if enableDebug {
		fmt.Printf("[ERROR] Failed to scrape target:\n%v\n", scrapeResponseErr)
		// }
		message := fmt.Sprintf("500 - Failed to scrape target: %s\n", scrapeResponseErr)
		http.Error(response, message, 500)
		return nil
	}
	defer scrapeResponse.Body.Close()
	rawData, rawDataErr := ioutil.ReadAll(scrapeResponse.Body)
	if rawDataErr != nil {
		if enableDebug {
			fmt.Printf("[DEBUG] Failed to read data from target:\n%v", rawDataErr)
		}
		message := fmt.Sprintf("500 - Failed to scrape target: %s\n", rawDataErr)
		http.Error(response, message, 500)
		return nil
	}

	// Parse
	data := schema.LolMinerMetric{}
	if err := json.Unmarshal(rawData, &data); err != nil {
		if enableDebug {
			fmt.Printf("[DEBUG] Failed to unmarshal data from target:\n%v", err)
		}
		message := fmt.Sprintf("500 - Failed to parse scraped data: %s\n", err)
		http.Error(response, message, 500)
		return nil
	}

	// // Validate
	// if data.Session.PerformanceUnit != "mh/s" {
	// 	message := fmt.Sprintf("500 - Target returned unexpected performance unit (expected \"mh/s\"): %s\n", data.Session.PerformanceUnit)
	// 	fmt.Printf("[ERROR] Failed to validate data from target: %s", message)

	// 	http.Error(response, message, 500)
	// 	return nil
	// }

	return &data
}

// Builds a new registry, adds scraped data to it and returns it if successful or nil if not.
func buildRegistry(response http.ResponseWriter, data *schema.LolMinerMetric) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())

	addExporterMetrics(registry)
	addSoftwareMetrics(registry, data)
	for i, algo := range data.Algorithms {
		addAlgoMetrics(int64(i), registry, &algo)
	}
	// // addStratumMetrics(registry, &data.Stratum)
	addSessionMetrics(registry, &data.Session)
	for i, gpuData := range data.Workers {
		addGPUMetrics(int64(i), registry, &gpuData) //&data.Workers[0]
	}

	return registry
}

func addExporterMetrics(registry *prometheus.Registry) {
	// Info
	infoLabels := make(prometheus.Labels)
	infoLabels["version"] = appVersion
	var infoMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "exporter_info",
		Help:      "Metadata about the exporter.",
	}, labelsKeys(infoLabels))
	infoMetric.With(infoLabels).Set(1)
	registry.MustRegister(infoMetric)
}

func addSoftwareMetrics(registry *prometheus.Registry, data *schema.LolMinerMetric) {
	// Info
	infoLabels := make(prometheus.Labels)
	infoLabels["software"] = data.Software
	var infoMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "software_info",
		Help:      "Metadata about the software.",
	}, labelsKeys(infoLabels))
	infoMetric.With(infoLabels).Set(1)
	registry.MustRegister(infoMetric)

	// // Active GPUs - Num_Workers
	var activeGPUsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "session_active_gpus",
		Help:      "Number of active GPUs.",
	}, labelsKeys(infoLabels))
	activeGPUsMetric.With(infoLabels).Set(float64(data.NumWorkers))
	registry.MustRegister(activeGPUsMetric)
}

func addAlgoMetrics(i int64, registry *prometheus.Registry, data *schema.LolMinerAlgoMetric) {
	// Info
	infoLabels := make(prometheus.Labels)
	infoLabels["algorithm"] = data.Algorithm
	var infoMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "mining_info_" + strconv.FormatInt(i, 10),
		Help:      "Metadata about mining.",
	}, labelsKeys(infoLabels))
	infoMetric.With(infoLabels).Set(1)
	registry.MustRegister(infoMetric)

	// Accepted shares
	var acceptedSharesMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "session_accepted_shares_total",
		Help:      "Number of accepted shares for this session.",
	}, labelsKeys(infoLabels))
	acceptedSharesMetric.With(infoLabels).Add(float64(data.TotalAccepted))
	registry.MustRegister(acceptedSharesMetric)

}

// func addStratumMetrics(registry *prometheus.Registry, data *schema.Lol) {
// 	// Common labels for subsystem
// 	commonLabels := make(prometheus.Labels)
// 	commonLabels["stratum_pool"] = data.CurrentPool
// 	commonLabels["stratum_user"] = data.CurrentUser

// 	// Info
// 	infoLabels := make(prometheus.Labels, len(commonLabels))
// 	for k, v := range commonLabels {
// 		infoLabels[k] = v
// 	}
// 	var infoMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 		Namespace: namespace,
// 		Name:      "stratum_info",
// 		Help:      "Metadata about the stratum.",
// 	}, labelsKeys(infoLabels))
// 	infoMetric.With(infoLabels).Set(1)
// 	registry.MustRegister(infoMetric)

// 	// Avg. latency
// 	var avgLatencyMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 		Namespace: namespace,
// 		Name:      "stratum_average_latency_seconds",
// 		Help:      "Average latency for the stratum (s).",
// 	}, labelsKeys(commonLabels))
// 	avgLatencyMetric.With(commonLabels).Set(data.AverageLatencyMs / 1000)
// 	registry.MustRegister(avgLatencyMetric)
// }

func addSessionMetrics(registry *prometheus.Registry, data *schema.LolMinerSessionMetric) {
	// Common labels for subsystem
	commonLabels := make(prometheus.Labels)
	commonLabels["session_startup_time"] = data.StartupString

	// Info
	infoLabels := make(prometheus.Labels, len(commonLabels))
	for k, v := range commonLabels {
		infoLabels[k] = v
	}
	var infoMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "session_info",
		Help:      "Metadata about the session.",
	}, labelsKeys(infoLabels))
	infoMetric.With(infoLabels).Set(1)
	registry.MustRegister(infoMetric)

	// Startup
	var startupMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "session_startup_seconds_timestamp",
		Help:      "Timestamp for the start of the session.",
	}, labelsKeys(commonLabels))
	startupMetric.With(commonLabels).Set(float64(data.Startup))
	registry.MustRegister(startupMetric)

	// Uptime
	var uptimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "session_uptime_seconds",
		Help:      "Uptime for the session (s).",
	}, labelsKeys(commonLabels))
	uptimeMetric.With(commonLabels).Set(float64(data.Uptime))
	registry.MustRegister(uptimeMetric)

	// Last update
	var lastUpdateMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "session_last_update_seconds_timestamp",
		Help:      "Timestamp for last update.",
	}, labelsKeys(commonLabels))
	lastUpdateMetric.With(commonLabels).Set(float64(data.LastUpdate))
	registry.MustRegister(lastUpdateMetric)

	// // Performance
	// var totalPerformanceMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace: namespace,
	// 	Name:      "session_performance_total_mhps",
	// 	Help:      "Total current performance for the session (Mh/s).",
	// }, labelsKeys(commonLabels))
	// totalPerformanceMetric.With(commonLabels).Set(float64(data.))
	// registry.MustRegister(totalPerformanceMetric)

	// // Submitted shares - Replaced with Algorithms.TotalRejected
	// var submittedSharesMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Name:      "session_submitted_shares_total",
	// 	Help:      "Number of submitted shares for this session.",
	// }, labelsKeys(commonLabels))
	// submittedSharesMetric.With(commonLabels).Add(float64(data.SubmittedShares))
	// registry.MustRegister(submittedSharesMetric)

	// // Total power
	// var totalPowerMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace: namespace,
	// 	Name:      "session_power_total_watts",
	// 	Help:      "Total current power usage for the session (Watt).",
	// }, labelsKeys(commonLabels))
	// totalPowerMetric.With(commonLabels).Set(float64(data.TotalPower))
	// registry.MustRegister(totalPowerMetric)
}

func addGPUMetrics(i int64, registry *prometheus.Registry, data *schema.LolMinerWorkerMetric) {
	// Common labels for subsystem
	commonLabels := make(prometheus.Labels)
	commonLabels["gpu_index"] = fmt.Sprintf("%d", data.Index)

	// Info
	infoLabels := make(prometheus.Labels, len(commonLabels))
	for k, v := range commonLabels {
		infoLabels[k] = v
	}
	infoLabels["name"] = data.Name
	infoLabels["pcie_address"] = data.PcieAddress
	var infoMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "gpu_info_" + strconv.FormatInt(i, 10),
		Help:      "Metadata about a GPU.",
	}, labelsKeys(infoLabels))
	infoMetric.With(infoLabels).Set(1)
	registry.MustRegister(infoMetric)

	// // Performance
	// var performanceMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace: namespace,
	// 	Name:      "gpu_performance_mhps",
	// 	Help:      "GPU performance (Mh/s).",
	// }, labelsKeys(commonLabels))
	// performanceMetric.With(commonLabels).Set(float64(data.Performance))
	// registry.MustRegister(performanceMetric)

	// Power
	var powerMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "gpu_power_watts_" + strconv.FormatInt(i, 10),
		Help:      "GPU power usage (Watt).",
	}, labelsKeys(commonLabels))
	powerMetric.With(commonLabels).Set(float64(data.Power))
	registry.MustRegister(powerMetric)

	// Fan speed
	var fanSpeedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "gpu_fan_speed_" + strconv.FormatInt(i, 10),
		Help:      "GPU fan speed (%).",
	}, labelsKeys(commonLabels))
	fanSpeedMetric.With(commonLabels).Set(float64(data.FanSpeed))
	registry.MustRegister(fanSpeedMetric)

	// Temperature
	var temperatureMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "gpu_temperature_celsius_" + strconv.FormatInt(i, 10),
		Help:      "GPU temperature (deg. C).",
	}, labelsKeys(commonLabels))
	temperatureMetric.With(commonLabels).Set(float64(data.CoreTemp))
	registry.MustRegister(temperatureMetric)

	// Locate in algorithm
	// // Session accepted shares
	// var sessionAcceptedSharesMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Name:      "gpu_session_accepted_shares_total",
	// 	Help:      "Number of accepted shared for the GPU during the current session.",
	// }, labelsKeys(commonLabels))
	// sessionAcceptedSharesMetric.With(commonLabels).Add(float64(data.SessionAcceptedShares))
	// registry.MustRegister(sessionAcceptedSharesMetric)

	// // Session submitted shares
	// var sessionSubmittedSharesMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Name:      "gpu_session_submitted_shares_total",
	// 	Help:      "Number of submitted shared for the GPU during the current session.",
	// }, labelsKeys(commonLabels))
	// sessionSubmittedSharesMetric.With(commonLabels).Add(float64(data.SessionSubmittedShares))
	// registry.MustRegister(sessionSubmittedSharesMetric)

	// // Session HW errors
	// var sessionHwErrorsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Name:      "gpu_session_hardware_errors_total",
	// 	Help:      "Number of hardware errors for the GPU during the current session.",
	// }, labelsKeys(commonLabels))
	// sessionHwErrorsMetric.With(commonLabels).Add(float64(data.SessionHWErrors))
	// registry.MustRegister(sessionHwErrorsMetric)
}

func labelsKeys(fullMap prometheus.Labels) []string {
	keys := make([]string, len(fullMap))
	i := 0
	for key := range fullMap {
		keys[i] = key
		i++
	}
	return keys
}
