package mysql

import (
	"context"
	"database/sql"

	"github.com/mmmtmi/excel-template-mapper/internal/model"
)

// Edit by CODEX: small repos to satisfy internal/service interfaces.

type TemplateRepository struct {
	DB *sql.DB
}

func (r *TemplateRepository) GetTemplateByName(ctx context.Context, name string) (*model.Template, error) {
	return GetTemplateByName(ctx, r.DB, name)
}

func (r *TemplateRepository) ListTemplates(ctx context.Context) ([]*model.Template, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, target, sheet_name, header_row, data_start_row
		FROM mapping_templates
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.Template
	for rows.Next() {
		var sheet sql.NullString
		tpl := model.Template{}
		if err := rows.Scan(&tpl.ID, &tpl.Name, &tpl.Target, &sheet, &tpl.HeaderRow, &tpl.DataStartRow); err != nil {
			return nil, err
		}
		if sheet.Valid {
			tpl.SheetName = &sheet.String
		}
		out = append(out, &tpl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type RuleRepository struct {
	DB *sql.DB
}

func (r *RuleRepository) ListRulesByTemplateID(ctx context.Context, templateID string) ([]model.Rule, error) {
	return ListRulesByTemplateID(ctx, r.DB, templateID)
}

