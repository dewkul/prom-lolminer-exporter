package schema

type LolMinerMetric struct {
	Software   string
	Session    LolMinerSessionMetric
	NumWorkers uint8 `json:"Num_Workers"`
	Workers    []LolMinerWorkerMetric
	NumAlgo    uint8 `json:"Num_Algorithms"`
	Algorithms []LolMinerAlgoMetric
}

type LolMinerSessionMetric struct {
	Startup       uint
	StartupString string `json:"StartupString"`
	Uptime        uint
	LastUpdate    uint64 `json:"Last_Update"`
}

type LolMinerWorkerMetric struct {
	Index       uint8
	Name        string
	Power       float64
	CCLK        uint32
	MCLK        uint32
	CoreTemp    int8   `json:"Core_Temp"`
	JuncTemp    int8   `json:"Juc_Temp"`
	MemTemp     int8   `json:"Mem_Temp"`
	FanSpeed    uint8  `json:"Fan_Speed"`
	PcieAddress string `json:"PCIE_Address"`
}

type LolMinerAlgoMetric struct {
	Algorithm          string
	AlgorithmAppendix  string `json:"Algorithm_Appendix"`
	Pool               string
	User               string
	Worker             string
	PerformanceUnit    string    `json:"Performance_Unit"`
	PerformanceFactor  int64     `json:"Performance_Factor"`
	TotalAccepted      uint64    `json:"Total_Accepted"`
	TotalRejected      uint64    `json:"Total_Rejected"`
	TotalStale         uint64    `json:"Total_Stales"`
	TotalError         uint64    `json:"Total_Errors"`
	WorkerPerformances []float32 `json:"Worker_Performances"`
	WorkerAccepted     []uint64  `json:"Worker_Accepted"`
	WorkerRejected     []uint64  `json:"Worker_Rejected"`
	WorkerStales       []uint64  `json:"Worker_Stales"`
	WorkerErrors       []uint64  `json:"Worker_Errors"`
}
