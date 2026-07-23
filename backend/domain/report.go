package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ReportInputType string

const (
	ReportInputTypeText      ReportInputType = "text"
	ReportInputTypeTextarea  ReportInputType = "textarea"
	ReportInputTypeSelect    ReportInputType = "select"
	ReportInputTypeDate      ReportInputType = "date"
	ReportInputTypeDateRange ReportInputType = "date_range"
)

type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusRunning   ReportStatus = "running"
	ReportStatusCompleted ReportStatus = "completed"
	ReportStatusFailed    ReportStatus = "failed"
)

// ReportResultCode describes the terminal outcome without overloading status.
// A report with insufficient authorized evidence is completed, not failed.
type ReportResultCode string

const (
	ReportResultSuccess              ReportResultCode = "success"
	ReportResultInsufficientEvidence ReportResultCode = "insufficient_evidence"
	ReportResultModelError           ReportResultCode = "model_error"
	ReportResultRetrievalError       ReportResultCode = "retrieval_error"
	ReportResultInvalidCitation      ReportResultCode = "invalid_citation"
	ReportResultTimeout              ReportResultCode = "timeout"
)

type ReportDateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ReportInputValue is deliberately constrained: V1 accepts a scalar text
// value or a date range, rather than arbitrary JSON supplied by clients.
type ReportInputValue struct {
	Value     string           `json:"value,omitempty"`
	DateRange *ReportDateRange `json:"date_range,omitempty"`
}

type ReportInputValues map[string]ReportInputValue

func (v *ReportInputValues) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("report input values must be an object: %w", err)
	}
	values := make(ReportInputValues, len(raw))
	for key, value := range raw {
		if len(bytes.TrimSpace(value)) == 0 {
			return fmt.Errorf("report input %q is empty", key)
		}
		var text string
		if err := json.Unmarshal(value, &text); err == nil {
			values[key] = ReportInputValue{Value: text}
			continue
		}
		var dateRange ReportDateRange
		decoder := json.NewDecoder(bytes.NewReader(value))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&dateRange); err != nil {
			return fmt.Errorf("report input %q must be string or date range: %w", key, err)
		}
		if dateRange.Start == "" && dateRange.End == "" {
			return fmt.Errorf("report input %q has empty date range", key)
		}
		values[key] = ReportInputValue{DateRange: &dateRange}
	}
	*v = values
	return nil
}

func (v ReportInputValues) MarshalJSON() ([]byte, error) {
	values := make(map[string]any, len(v))
	for key, value := range v {
		if value.DateRange != nil {
			values[key] = value.DateRange
			continue
		}
		values[key] = value.Value
	}
	return json.Marshal(values)
}

type ReportProfileInputField struct {
	Key          string          `json:"key"`
	Label        string          `json:"label"`
	Type         ReportInputType `json:"type"`
	Required     bool            `json:"required"`
	Placeholder  string          `json:"placeholder"`
	DefaultValue string          `json:"default_value"`
	Options      []string        `json:"options"`
}
type ReportProfileSection struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Instruction string `json:"instruction"`
	Required    bool   `json:"required"`
	Order       int    `json:"order"`
}
type ReportProfile struct {
	ID               string                    `json:"id" gorm:"primaryKey"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	EditionIDs       []EditionID               `json:"edition_ids" gorm:"serializer:json;type:jsonb"`
	SystemPrompt     string                    `json:"system_prompt"`
	InputFields      []ReportProfileInputField `json:"input_fields" gorm:"serializer:json;type:jsonb"`
	Sections         []ReportProfileSection    `json:"sections" gorm:"serializer:json;type:jsonb"`
	CitationRequired bool                      `json:"citation_required"`
	Enabled          bool                      `json:"enabled"`
	Version          string                    `json:"version"`
	IsBuiltin        bool                      `json:"is_builtin"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	DeletedAt        *time.Time                `json:"-" gorm:"index"`
}
type Report struct {
	ID              string            `json:"id" gorm:"primaryKey"`
	KBID            string            `json:"kb_id"`
	ProfileID       string            `json:"profile_id"`
	ProfileVersion  string            `json:"profile_version"`
	ProfileSnapshot ReportProfile     `json:"profile_snapshot" gorm:"serializer:json;type:jsonb"`
	CreatedBy       uint              `json:"created_by"`
	Status          ReportStatus      `json:"status"`
	Title           string            `json:"title"`
	InputValues     ReportInputValues `json:"input_values" gorm:"serializer:json;type:jsonb"`
	Content         string            `json:"content"`
	ErrorMessage    string            `json:"error_message"`
	ResultCode      ReportResultCode  `json:"result_code"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	CompletedAt     *time.Time        `json:"completed_at"`
	CitationCount   int               `json:"-" gorm:"column:citation_count;->"`
}
type ReportCitation struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ReportID      string    `json:"report_id"`
	CitationIndex int       `json:"citation_index"`
	NodeID        string    `json:"node_id"`
	DocumentName  string    `json:"document_name"`
	Locator       string    `json:"locator"`
	Excerpt       string    `json:"excerpt"`
	CreatedAt     time.Time `json:"created_at"`
}

func (p ReportProfile) Validate() error {
	if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Name) == "" || len(p.Name) > 120 || len(p.SystemPrompt) > 16000 || p.Version == "" {
		return errors.New("invalid report profile identity or prompt")
	}
	validEdition := map[EditionID]bool{EditionCommon: true, EditionResearch: true, EditionLegal: true, EditionFinance: true}
	if len(p.EditionIDs) == 0 {
		return errors.New("report profile requires edition")
	}
	fields := map[string]bool{}
	validType := map[ReportInputType]bool{ReportInputTypeText: true, ReportInputTypeTextarea: true, ReportInputTypeSelect: true, ReportInputTypeDate: true, ReportInputTypeDateRange: true}
	for _, f := range p.InputFields {
		if f.Key == "" || fields[f.Key] || !validType[f.Type] || (f.Type == ReportInputTypeSelect && len(f.Options) == 0) {
			return errors.New("invalid report input field")
		}
		fields[f.Key] = true
	}
	sections := map[string]bool{}
	orders := map[int]bool{}
	if len(p.Sections) == 0 {
		return errors.New("report profile requires section")
	}
	for _, s := range p.Sections {
		if s.Key == "" || sections[s.Key] || orders[s.Order] {
			return errors.New("invalid report section")
		}
		sections[s.Key] = true
		orders[s.Order] = true
	}
	for _, id := range p.EditionIDs {
		if !validEdition[id] {
			return fmt.Errorf("invalid edition: %s", id)
		}
	}
	return nil
}

func BuiltinReportProfiles() []ReportProfile {
	specs := []struct {
		id, name string
		editions []EditionID
	}{
		{"builtin.common.knowledge-report", "知识库综合报告", []EditionID{EditionCommon}}, {"builtin.research.literature-review", "文献综述", []EditionID{EditionResearch}}, {"builtin.research.state-analysis", "研究现状分析", []EditionID{EditionResearch}}, {"builtin.research.material-review", "资料综述", []EditionID{EditionResearch}}, {"builtin.legal.opinion", "法律意见书", []EditionID{EditionLegal}}, {"builtin.legal.case-analysis", "案件材料分析", []EditionID{EditionLegal}}, {"builtin.legal.contract-risk", "合同风险审查", []EditionID{EditionLegal}}, {"builtin.finance.industry-analysis", "行业分析", []EditionID{EditionFinance}}, {"builtin.finance.company-analysis", "公司资料分析", []EditionID{EditionFinance}}, {"builtin.finance.financial-summary", "财务信息摘要", []EditionID{EditionFinance}}}
	result := make([]ReportProfile, 0, len(specs))
	for _, s := range specs {
		result = append(result, ReportProfile{ID: s.id, Name: s.name, Description: s.name, EditionIDs: s.editions, SystemPrompt: "Use only the supplied authorized materials. Cite sources when available.", InputFields: []ReportProfileInputField{{Key: "topic", Label: "Topic", Type: ReportInputTypeText, Required: true}}, Sections: []ReportProfileSection{{Key: "summary", Title: "Summary", Instruction: "Summarize the authorized materials.", Required: true, Order: 1}}, CitationRequired: true, Enabled: true, Version: "1.0.0", IsBuiltin: true})
	}
	return result
}
