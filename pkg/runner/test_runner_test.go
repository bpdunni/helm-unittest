package runner

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/bdun1013/helm-snapshot/pkg/printer"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/assert"
)

var sectionBeginPattern = regexp.MustCompile("( PASS | FAIL |\n*###|\n*Charts:|\n*Snapshot Summary:)")
var timePattern = regexp.MustCompile("Time:\\s+([\\d\\.]+)ms")

func makeOutputSnapshotable(output string) []interface{} {
	timeLoc := timePattern.FindStringSubmatchIndex(output)[2:4]
	timeAgnosticOutput := output[:timeLoc[0]] + "XX.XXX" + output[timeLoc[1]:]

	sectionBeggingLocs := sectionBeginPattern.FindAllStringIndex(timeAgnosticOutput, -1)
	sections := make([]string, len(sectionBeggingLocs))

	suiteBeginIdx := -1
	for sectionIdx := 0; sectionIdx < len(sections); sectionIdx++ {
		start := sectionBeggingLocs[sectionIdx][0]
		var end int
		if sectionIdx >= len(sections)-1 {
			end = len(timeAgnosticOutput)
		} else {
			end = sectionBeggingLocs[sectionIdx+1][0]
		}

		sectionContent := timeAgnosticOutput[start:end]
		sectionBegin := sectionContent[:6]
		if sectionBegin == " PASS " || sectionBegin == " FAIL " {
			sections[sectionIdx] = strings.TrimRight(sectionContent, "\n")
			if suiteBeginIdx == -1 {
				suiteBeginIdx = sectionIdx
			}
		} else {
			sections[sectionIdx] = sectionContent
			if suiteBeginIdx != -1 {
				sort.Strings(sections[suiteBeginIdx:sectionIdx])
				suiteBeginIdx = -1
			}
		}
	}

	sectionsToRetrun := make([]interface{}, len(sections))
	for idx, section := range sections {
		sectionsToRetrun[idx] = section
	}
	return sectionsToRetrun
}

func TestRunnerOkWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: printer.NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/basic"})
	fmt.Println(buffer.String())
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestRunnerOkWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: printer.NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests_failed/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/basic"})
	assert.False(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestRunnerWithTestsInSubchart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: printer.NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: true,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/with-subchart"})
	fmt.Println(buffer.String())
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestRunnerWithTestsInSubchartButFlagFalse(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: printer.NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: false,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/with-subchart"})
	fmt.Println(buffer.String())
	assert.True(t, passed)
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}
