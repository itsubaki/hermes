package cost

import (
	"fmt"
	"os"
	"testing"

	"github.com/itsubaki/hermes/pkg/calendar"
)

func TestFetch(t *testing.T) {
	os.Setenv("AWS_PROFILE", "example")

	m := calendar.LastNDays(1)[0]
	out, err := Fetch(m.Start, m.End, []string{})
	if err != nil {
		t.Errorf("fetch cost group by linked account: %v", err)
	}

	for _, c := range out {
		fmt.Println(c)
	}
}

func TestFetchWith(t *testing.T) {
	os.Setenv("AWS_PROFILE", "example")

	m := calendar.LastNDays(1)[0]
	out, err := FetchWith(m.Start, m.End, []string{
		"Amazon Elastic Compute Cloud - Compute",
		"Amazon ElastiCache",
		"Amazon Relational Database Service",
	}, []string{
		"UnblendedCost",
	})

	if err != nil {
		t.Errorf("fetch cost group by linked account: %v", err)
	}

	for _, c := range out {
		fmt.Println(c)
	}
}
