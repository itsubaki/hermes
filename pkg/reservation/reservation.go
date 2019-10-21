package reservation

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/itsubaki/hermes/pkg/usage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
)

type Utilization struct {
	AccountID        string  `json:"account_id"`
	Description      string  `json:"description"`
	Region           string  `json:"region"`
	InstanceType     string  `json:"instance_type"`
	Platform         string  `json:"platform,omitempty"`
	CacheEngine      string  `json:"cache_engine,omitempty"`
	DatabaseEngine   string  `json:"database_engine,omitempty"`
	DeploymentOption string  `json:"deployment_option,omitempty"`
	Date             string  `json:"date"`
	Hours            float64 `json:"hours"`
}

func (u Utilization) UsageType() string {
	return fmt.Sprintf("%s-%s:%s", region[u.Region], u.Usage(), u.InstanceType)
}

func (u Utilization) PFEngine() string {
	return fmt.Sprintf("%s%s%s", u.Platform, u.CacheEngine, u.DatabaseEngine)
}

func (u Utilization) OSEngine() string {
	return fmt.Sprintf("%s%s%s", usage.OperatingSystem[u.Platform], u.CacheEngine, u.DatabaseEngine)
}

func (u Utilization) Usage() string {
	if len(u.Platform) > 0 {
		return "BoxUsage"
	}

	if len(u.CacheEngine) > 0 {
		return "NodeUsage"
	}

	if len(u.DatabaseEngine) > 0 && u.DeploymentOption == "Single-AZ" {
		return "InstanceUsage"
	}

	if len(u.DatabaseEngine) > 0 && u.DeploymentOption == "Multi-AZ" {
		return "Multi-AZUsage"
	}

	panic("invalid usage")
}

func (u Utilization) String() string {
	return u.JSON()
}

func (u Utilization) JSON() string {
	bytes, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

// Service Filter are
// Amazon Elastic Compute Cloud - Compute
// Amazon Relational Database Service
// Amazon ElastiCache, Amazon Redshift
// Amazon Elasticsearch Service
type getInputFunc func() (*costexplorer.Expression, []*costexplorer.GroupDefinition)

func getComputeInput() (*costexplorer.Expression, []*costexplorer.GroupDefinition) {
	return &costexplorer.Expression{
			Dimensions: &costexplorer.DimensionValues{
				Key:    aws.String("SERVICE"),
				Values: []*string{aws.String("Amazon Elastic Compute Cloud - Compute")},
			},
		}, []*costexplorer.GroupDefinition{
			{
				Key:  aws.String("INSTANCE_TYPE"),
				Type: aws.String("DIMENSION"),
			},
			{
				Key:  aws.String("REGION"),
				Type: aws.String("DIMENSION"),
			},
			{
				Key:  aws.String("PLATFORM"),
				Type: aws.String("DIMENSION"),
			},
		}
}

func getCacheInput() (*costexplorer.Expression, []*costexplorer.GroupDefinition) {
	return &costexplorer.Expression{
			Dimensions: &costexplorer.DimensionValues{
				Key:    aws.String("SERVICE"),
				Values: []*string{aws.String("Amazon ElastiCache")},
			},
		}, []*costexplorer.GroupDefinition{
			{
				Key:  aws.String("INSTANCE_TYPE"),
				Type: aws.String("DIMENSION"),
			},
			{
				Key:  aws.String("REGION"),
				Type: aws.String("DIMENSION"),
			},
			{
				Key:  aws.String("CACHE_ENGINE"),
				Type: aws.String("DIMENSION"),
			},
		}
}

func getDatabaseInput() (*costexplorer.Expression, []*costexplorer.GroupDefinition) {
	return &costexplorer.Expression{
			Dimensions: &costexplorer.DimensionValues{
				Key:    aws.String("SERVICE"),
				Values: []*string{aws.String("Amazon Relational Database Service")},
			},
		}, []*costexplorer.GroupDefinition{
			{
				Key:  aws.String("INSTANCE_TYPE"),
				Type: aws.String("DIMENSION"),
			},
			{
				Key:  aws.String("REGION"),
				Type: aws.String("DIMENSION"),
			},
			{
				Key:  aws.String("DATABASE_ENGINE"),
				Type: aws.String("DIMENSION"),
			},
			{
				Key:  aws.String("DEPLOYMENT_OPTION"),
				Type: aws.String("DIMENSION"),
			},
		}
}

var getInputFuncList = []getInputFunc{
	getComputeInput,
	getCacheInput,
	getDatabaseInput,
}

func fetch(input costexplorer.GetReservationCoverageInput) ([]Utilization, error) {
	out := make([]Utilization, 0)

	c := costexplorer.New(session.Must(session.NewSession()))
	var token *string
	for {
		input.NextPageToken = token
		rc, err := c.GetReservationCoverage(&input)
		if err != nil {
			return out, fmt.Errorf("get reservation coverage: %v", err)
		}

		for _, t := range rc.CoveragesByTime {
			for _, g := range t.Groups {
				if *g.Coverage.CoverageHours.ReservedHours == "0" {
					continue
				}

				index := strings.LastIndex(*input.TimePeriod.Start, "-")
				date := (*input.TimePeriod.Start)[:index]

				hours, err := strconv.ParseFloat(*g.Coverage.CoverageHours.ReservedHours, 64)
				if err != nil {
					return out, fmt.Errorf("parse float reserved hours: %v", err)
				}

				u := Utilization{
					Region:       *g.Attributes["region"],
					InstanceType: *g.Attributes["instanceType"],
					Date:         date,
					Hours:        hours,
				}

				if g.Attributes["platform"] != nil {
					u.Platform = *g.Attributes["platform"]
				}

				if g.Attributes["cacheEngine"] != nil {
					u.CacheEngine = *g.Attributes["cacheEngine"]
				}

				if g.Attributes["databaseEngine"] != nil {
					u.DatabaseEngine = *g.Attributes["databaseEngine"]
					u.DeploymentOption = *g.Attributes["deploymentOption"]
				}

				out = append(out, u)
			}
		}

		if rc.NextPageToken == nil {
			break
		}
		token = rc.NextPageToken
	}

	return out, nil
}

func Fetch(start, end string) ([]Utilization, error) {
	linked, err := usage.FetchLinkedAccount(start, end)
	if err != nil {
		return nil, fmt.Errorf("get linked account: %v", err)
	}

	out := make([]Utilization, 0)
	for _, f := range getInputFuncList {
		for _, a := range linked {
			exp, groupby := f()

			and := make([]*costexplorer.Expression, 0)
			and = append(and, &costexplorer.Expression{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("LINKED_ACCOUNT"),
					Values: []*string{aws.String(a.ID)},
				},
			})
			and = append(and, exp)

			input := costexplorer.GetReservationCoverageInput{
				Metrics: []*string{aws.String("Hour")},
				Filter: &costexplorer.Expression{
					And: and,
				},
				GroupBy: groupby,
				TimePeriod: &costexplorer.DateInterval{
					Start: &start,
					End:   &end,
				},
			}

			u, err := fetch(input)
			if err != nil {
				return out, fmt.Errorf("fetch: %v", err)
			}

			for i := range u {
				u[i].AccountID = a.ID
				u[i].Description = a.Description
			}

			out = append(out, u...)
		}
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].DeploymentOption < out[j].DeploymentOption })
	sort.SliceStable(out, func(i, j int) bool { return out[i].DatabaseEngine < out[j].DatabaseEngine })
	sort.SliceStable(out, func(i, j int) bool { return out[i].CacheEngine < out[j].CacheEngine })
	sort.SliceStable(out, func(i, j int) bool { return out[i].Platform < out[j].Platform })
	sort.SliceStable(out, func(i, j int) bool { return out[i].Region < out[j].Region })
	sort.SliceStable(out, func(i, j int) bool { return out[i].AccountID < out[j].AccountID })

	return out, nil
}
