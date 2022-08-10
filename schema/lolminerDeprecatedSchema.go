package schema

// LolMinerDeprecatedResult
// Deprecated: The rule Title uses for word boundaries does not handle Unicode
// punctuation properly. Use golang.org/x/text/cases instead.
type LolMinerDeprecatedResult struct {
	Software string                         `json:"Software"`
	Mining   LolMinerDeprecateMiningResult  `json:"Mining"`
	Stratum  LolMinerDeprecateStratumResult `json:"Stratum"`
	Session  LolMinerDeprecateSessionResult `json:"Session"`
	GPUs     []LolMinerGPUDeprecateResult   `json:"GPUs"`
}

// LolMinerDeprecateMiningResult
// Deprecated: The rule Title uses for word boundaries does not handle Unicode
// punctuation properly. Use golang.org/x/text/cases instead.
type LolMinerDeprecateMiningResult struct {
	Algorithm string `json:"Algorithm"`
}

// LolMinerDeprecateStratumResult
// Deprecated: The rule Title uses for word boundaries does not handle Unicode
// punctuation properly. Use golang.org/x/text/cases instead.
type LolMinerDeprecateStratumResult struct {
	CurrentPool      string  `json:"Current_Pool"`
	CurrentUser      string  `json:"Current_User"`
	AverageLatencyMs float64 `json:"Average_Latency"`
}

// LolMinerDeprecateSessionResult
// Deprecated: The rule Title uses for word boundaries does not handle Unicode
// punctuation properly. Use golang.org/x/text/cases instead.
type LolMinerDeprecateSessionResult struct {
	Startup          int64   `json:"Startup"`
	StartupString    string  `json:"Startup_String"`
	Uptime           int64   `json:"Uptime"`
	LastUpdate       int64   `json:"Last_Update"`
	ActiveGPUs       int64   `json:"Active_GPUs"`
	TotalPerformance float64 `json:"Performance_Summary"`
	PerformanceUnit  string  `json:"Performance_Unit"`
	AcceptedShares   int64   `json:"Accepted"`
	SubmittedShares  int64   `json:"Submitted"`
	TotalPower       float64 `json:"TotalPower"`
}

// LolMinerGPUDeprecateResult
// Deprecated: The rule Title uses for word boundaries does not handle Unicode
// punctuation properly. Use golang.org/x/text/cases instead.
type LolMinerGPUDeprecateResult struct {
	Index                  int64   `json:"Index"`
	Name                   string  `json:"Name"`
	Performance            float64 `json:"Performance"`
	Power                  float64 `json:"Consumption (W)"`
	FanSpeedPercent        float64 `json:"Fan Speed (%)"`
	Temperature            float64 `json:"Temp (deg C)"`
	MinTemperature         float64 `json:"Mem Temp (deg C)"`
	SessionAcceptedShares  int64   `json:"Session_Accepted"`
	SessionSubmittedShares int64   `json:"Session_Submitted"`
	SessionHWErrors        int64   `json:"Session_HWErr"`
	PCIEAddress            string  `json:"PCIE_Address"`
}
