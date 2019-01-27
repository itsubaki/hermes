package costexp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/itsubaki/hermes/internal/costexp"
)

type Repository struct {
	Profile  string     `json:"profile"`
	Internal RecordList `json:"internal"`
}

func NewRepository(profile string, date []*costexplorer.DateInterval) (*Repository, error) {
	os.Setenv("AWS_PROFILE", profile)

	repo := &Repository{
		Profile: profile,
	}

	// start[:7] => 2018-12-01 -> 2018-12
	for i := range date {
		q, err := costexp.New().GetUsageQuantity(date[i])
		if err != nil {
			return nil, fmt.Errorf("get usage quantity: %v", err)
		}

		for _, qq := range q {
			repo.Internal = append(repo.Internal, &Record{
				AccountID:      qq.AccountID,
				Description:    qq.Description,
				Date:           qq.Date,
				UsageType:      qq.UsageType,
				Platform:       qq.Platform,
				CacheEngine:    qq.CacheEngine,
				DatabaseEngine: qq.DatabaseEngine,
				InstanceHour:   qq.InstanceHour,
				InstanceNum:    qq.InstanceNum,
			})
		}
	}

	return repo, nil
}

func Read(path string) (*Repository, error) {
	read, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}

	repo := &Repository{}
	if err := repo.Deserialize(read); err != nil {
		return nil, fmt.Errorf("new repository: %v", err)
	}

	return repo, nil
}

func (r *Repository) Write(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	bytes, err := r.Serialize()
	if err != nil {
		return fmt.Errorf("serialize: %v", err)
	}

	if err := ioutil.WriteFile(path, bytes, os.ModePerm); err != nil {
		return fmt.Errorf("write file: %v", err)
	}

	return nil
}

func (r *Repository) Serialize() ([]byte, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return []byte{}, fmt.Errorf("marshal: %v", err)
	}

	return bytes, nil
}

func (r *Repository) Deserialize(bytes []byte) error {
	if err := json.Unmarshal(bytes, r); err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}

	return nil
}

func (r *Repository) SelectAll() RecordList {
	return r.Internal
}

func (r *Repository) AccountID() []string {
	selected := make(map[string]bool)
	for i := range r.Internal {
		selected[r.Internal[i].AccountID] = true
	}

	out := []string{}
	for k := range selected {
		out = append(out, k)
	}

	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
