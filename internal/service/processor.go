package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mmmtmi/excel-template-mapper/internal/excel"
	"github.com/mmmtmi/excel-template-mapper/internal/model"
	"github.com/xuri/excelize/v2"
)

type TemplateRepo interface {
	GetTemplateByName(ctx context.Context, name string) (*model.Template, error)
	ListTemplates(ctx context.Context) ([]*model.Template, error)
}

type RuleRepo interface {
	ListRulesByTemplateID(ctx context.Context, templateID string) ([]model.Rule, error)
}

type Processor struct {
	templates TemplateRepo
	rules     RuleRepo
}

func NewProcessor(templates TemplateRepo, rules RuleRepo) *Processor {
	return &Processor{templates: templates, rules: rules}
}

var DefaultReadOptions = excel.ReadOptions{
	HeaderRow:    1,
	DataStartRow: 2,
	TrimHeader:   true,
	SkipEmptyKey: true,
}

func (p *Processor) ProcessExcelOnly(path string, opt *excel.ReadOptions) ([]map[string]any, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	if opt == nil {
		d := DefaultReadOptions
		opt = &d
	}

	table, err := excel.ReadTable(f, *opt)
	if err != nil {
		return nil, err
	}

	outRows := make([]map[string]any, 0, len(table.Rows))
	for _, row := range table.Rows {
		outRow := make(map[string]any)
		for key, value := range row.Values {
			outRow[key] = value
		}
		outRows = append(outRows, outRow)
	}

	return outRows, nil
}

func (p *Processor) ProcessWithTemplate(ctx context.Context, templateName string, path string) ([]map[string]any, error) {
	tpl, err := p.templates.GetTemplateByName(ctx, templateName)
	if err != nil {
		return nil, err
	}

	rules, err := p.rules.ListRulesByTemplateID(ctx, tpl.ID)
	if err != nil {
		return nil, err
	}

	opt := buildReadOptionsFromTemplate(tpl)
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	table, err := excel.ReadTable(f, opt)
	if err != nil {
		return nil, err
	}

	requiredRules, err := requiredHeaderRules(rules)
	if err != nil {
		return nil, err
	}
	if err := validateRequiredHeaders(table.Headers, requiredRules); err != nil {
		return nil, err
	}

	outRows := make([]map[string]any, 0, len(table.Rows))
	for _, row := range table.Rows {
		outRow := make(map[string]any)
		for _, r := range rules {
			if r.SourceType != "HEADER" {
				continue
			}

			val, err := applyTransform(row.Values[r.SourceKey], r.Transform)
			if err != nil {
				return nil, err
			}

			outRow[r.TargetLabel] = val
		}

		if err := validateRequiredValues(outRow, requiredRules, row.ExcelRow); err != nil {
			return nil, err
		}
		outRows = append(outRows, outRow)
	}

	return outRows, nil
}

func isMissing(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func buildReadOptionsFromTemplate(tpl *model.Template) excel.ReadOptions {
	opt := DefaultReadOptions
	if tpl == nil {
		return opt
	}
	if tpl.SheetName != nil {
		opt.SheetName = *tpl.SheetName
	}
	if tpl.HeaderRow != 0 {
		opt.HeaderRow = tpl.HeaderRow
	}
	if tpl.DataStartRow != 0 {
		opt.DataStartRow = tpl.DataStartRow
	}
	return opt
}

func requiredHeaderRules(rules []model.Rule) ([]model.Rule, error) {
	out := make([]model.Rule, 0, len(rules))
	for _, r := range rules {
		if !r.Required {
			continue
		}
		if r.SourceType != "HEADER" {
			// 将来CELL対応したときに、ここを拡張する。
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func validateRequiredHeaders(headers []string, requiredRules []model.Rule) error {
	set := make(map[string]bool, len(headers))
	for _, h := range headers {
		set[h] = true
	}
	for _, r := range requiredRules {
		if !set[r.SourceKey] {
			return fmt.Errorf("required header missing: %s", r.SourceKey)
		}
	}
	return nil
}

func validateRequiredValues(outRow map[string]any, requiredRules []model.Rule, excelRow int) error {
	for _, r := range requiredRules {
		val := outRow[r.TargetLabel]
		if isMissing(val) {
			return fmt.Errorf("required value missing: row=%d header=%s label=%s", excelRow, r.SourceKey, r.TargetLabel)
		}
	}
	return nil
}

func applyTransform(val any, transform *string) (any, error) {
	if transform == nil {
		return val, nil
	}
	switch *transform {
	case "trim":
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s), nil
		}
		return val, nil
	default:
		return nil, fmt.Errorf("unsupported transform: %s", *transform)
	}
}
