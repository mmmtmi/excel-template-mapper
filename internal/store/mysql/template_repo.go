package mysql

import (
	"context"
	"database/sql"

	"github.com/mmmtmi/excel-template-mapper/internal/model"
)

func GetTemplateByName(ctx context.Context, db *sql.DB, name string) (*model.Template, error) {
	var sheet sql.NullString

	tpl := model.Template{}
	err := db.QueryRowContext(ctx,
		`SELECT id, name, target, sheet_name, header_row, data_start_row
	FROM mapping_templates
	WHERE name = ?`, name).Scan(
		&tpl.ID, &tpl.Name, &tpl.Target, &sheet, &tpl.HeaderRow, &tpl.DataStartRow,
	)
	if err != nil {
		return nil, err
	}

	if sheet.Valid {
		tpl.SheetName = &sheet.String
	}

	return &tpl, nil
}

func ListRulesByTemplateID(ctx context.Context, db *sql.DB, templateID string) ([]model.Rule, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT source_type, source_key, target_label, transform, required, priority
		FROM mapping_rules
		WHERE template_id = ?
		ORDER BY priority DESC, created_at ASC
	`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Rule
	for rows.Next() {
		var transform sql.NullString
		r := model.Rule{}
		if err := rows.Scan(&r.SourceType, &r.SourceKey, &r.TargetLabel, &transform, &r.Required, &r.Priority); err != nil {
			return nil, err
		}
		if transform.Valid {
			r.Transform = &transform.String
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
