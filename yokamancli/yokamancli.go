// Package yokamancli is a local mock replacement for gitlab.qa.dobest.com/qa/yokamancli
// It provides no-op implementations of the metrics reporting client,
// allowing the stress test tool to compile and run without the private repository.
package yokamancli

// EnvType represents the environment type for the metrics server.
type EnvType int

const (
	// Test environment
	Test EnvType = iota
	// Prod environment
	Prod
	// Default region
	Default = "default"
)

// SvrAddrOptions configures the metrics server address and environment.
type SvrAddrOptions struct {
	Env    EnvType
	Region string
}

// Request represents a metrics request record.
type Request struct {
	Name    string
	ReqTime int64
}

// Response represents a metrics response record.
type Response struct {
	Name     string
	ReqTime  int64
	RespTime int64
	Code     uint16
	RobotId  uint
	ErrCode  interface{}
}

// YoKaMan is the metrics reporting client (mock implementation).
type YoKaMan struct {
	projectID   uint
	userName    string
	caseName    string
	podRobotNum uint
	podID       int8
	reportID    int32
	podID2      uint
}

var sharedClient *YoKaMan

// YoKaManCli returns a singleton instance of the metrics client.
func YoKaManCli() *YoKaMan {
	if sharedClient == nil {
		sharedClient = &YoKaMan{}
	}
	return sharedClient
}

// SetProjectInfo sets the project ID and username.
func (y *YoKaMan) SetProjectInfo(projectID uint, userName string) {
	y.projectID = projectID
	y.userName = userName
}

// SetCaseName sets the test case name.
func (y *YoKaMan) SetCaseName(caseName string) {
	y.caseName = caseName
}

// SetPodInfo sets the pod robot count and pod ID.
func (y *YoKaMan) SetPodInfo(podRobotNum uint, podID int8) {
	y.podRobotNum = podRobotNum
	y.podID = podID
}

// SetTestInfo sets the report ID and pod ID.
func (y *YoKaMan) SetTestInfo(reportID int32, podID uint) {
	y.reportID = reportID
	y.podID2 = podID
}

// SetEnvOptions sets the server environment options.
func (y *YoKaMan) SetEnvOptions(opts SvrAddrOptions) {
	// no-op
}

// Start starts the metrics reporting client (mock: does nothing).
func (y *YoKaMan) Start() error {
	// no-op - mock implementation does not report metrics
	return nil
}

// Stop stops the metrics reporting client.
func (y *YoKaMan) Stop() {
	// no-op
}

// StatRequest records a request metric (mock: does nothing).
func (y *YoKaMan) StatRequest(req Request) error {
	// no-op
	return nil
}

// StatResponse records a response metric (mock: does nothing).
func (y *YoKaMan) StatResponse(resp Response) error {
	// no-op
	return nil
}
