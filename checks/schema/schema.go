package schema

import (
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

type Check interface {
	Run(wg *sync.WaitGroup, out chan<- CheckResult)
	Init(id string, name string, def string) error
}

// CheckResult : Information about the results of executing a check.
type CheckResult struct {
	Timestamp time.Time
	ID        string
	Name      string
	CheckType string
	Passed    bool
	Message   string
	Details   map[string]string
}

// Check : Information about a check to be run.
type Check struct {
	ID             string
	Name           string
	Definition     map[string]string
	DefinitionList []map[string]string
	WaitGroup      *sync.WaitGroup
	Output         chan<- CheckResult
}

// CheckDefinitions : Intermediate storage of check definitions and attributes
type CheckDefinitions struct {
	Checks     []map[string]gjson.Result
	Attributes map[string]map[string]string
}
