package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mmmtmi/excel-template-mapper/internal/model"
)

type fakeTemplateRepo struct {
	tpl *model.Template
	err error
}

func (r *fakeTemplateRepo) GetTemplateByName(ctx context.Context, name string) (*model.Template, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.tpl, nil
}

func (r *fakeTemplateRepo) ListTemplates(ctx context.Context) ([]*model.Template, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.tpl == nil {
		return []*model.Template{}, nil
	}
	return []*model.Template{r.tpl}, nil
}

type fakeRuleRepo struct {
	rules []model.Rule
	err   error
}

func (r *fakeRuleRepo) ListRulesByTemplateID(ctx context.Context, templateID string) ([]model.Rule, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.rules, nil
}

func TestProcessWithTemplate_Book1_TrimAndMapping(t *testing.T) {
	tplID := "tpl1"
	tpl := &model.Template{
		ID:           tplID,
		Name:         "demo_v1",
		Target:       "Demo",
		HeaderRow:    1,
		DataStartRow: 2,
	}

	trim := "trim"
	rules := []model.Rule{
		{
			SourceType:  "HEADER",
			SourceKey:   "カスタマID",
			TargetLabel: "顧客ID",
			Transform:   &trim,
			Required:    true,
			Priority:    1,
		},
	}

	proc := NewProcessor(&fakeTemplateRepo{tpl: tpl}, &fakeRuleRepo{rules: rules})

	book1 := filepath.Join("..", "excel", "Book1.xlsx")
	out, err := proc.ProcessWithTemplate(context.Background(), "demo_v1", book1)
	if err != nil {
		t.Fatalf("ProcessWithTemplate error: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected rows, got 0")
	}

	v, ok := out[0]["顧客ID"]
	if !ok {
		t.Fatalf("expected key %q in first output row", "顧客ID")
	}
	if v != "あああ" {
		t.Fatalf("expected trimmed value %q, got %#v", "あああ", v)
	}
}

func TestProcessWithTemplate_RequiredHeaderMissing(t *testing.T) {
	tpl := &model.Template{
		ID:           "tpl1",
		Name:         "demo_v1",
		Target:       "Demo",
		HeaderRow:    1,
		DataStartRow: 2,
	}

	rules := []model.Rule{
		{
			SourceType:  "HEADER",
			SourceKey:   "存在しないヘッダ",
			TargetLabel: "X",
			Required:    true,
			Priority:    1,
		},
	}

	proc := NewProcessor(&fakeTemplateRepo{tpl: tpl}, &fakeRuleRepo{rules: rules})
	book1 := filepath.Join("..", "excel", "Book1.xlsx")

	_, err := proc.ProcessWithTemplate(context.Background(), "demo_v1", book1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "required header missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyTransform(t *testing.T) {
	trim := "trim"
	got, err := applyTransform(" a ", &trim)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "a" {
		t.Fatalf("expected %q, got %#v", "a", got)
	}

	unknown := "date:2006-01-02"
	_, err = applyTransform("2026-02-01", &unknown)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
