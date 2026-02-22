package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/mmmtmi/excel-template-mapper/internal/model"
)

// Edit by CODEX: minimal INSERT/UPDATE helpers for PoC.

func InsertTemplate(ctx context.Context, db *sql.DB, tpl *model.Template, now time.Time) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO mapping_templates (id, name, target, sheet_name, header_row, data_start_row, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tpl.ID, tpl.Name, tpl.Target, tpl.SheetName, tpl.HeaderRow, tpl.DataStartRow, nil, now, now)
	return err
}

func GetTemplateByID(ctx context.Context, db *sql.DB, id string) (*model.Template, *string, time.Time, time.Time, error) {
	var sheet sql.NullString
	var notes sql.NullString
	var createdAt time.Time
	var updatedAt time.Time

	tpl := model.Template{}
	err := db.QueryRowContext(ctx, `
		SELECT id, name, target, sheet_name, header_row, data_start_row, notes, created_at, updated_at
		FROM mapping_templates
		WHERE id = ?
	`, id).Scan(&tpl.ID, &tpl.Name, &tpl.Target, &sheet, &tpl.HeaderRow, &tpl.DataStartRow, &notes, &createdAt, &updatedAt)
	if err != nil {
		return nil, nil, time.Time{}, time.Time{}, err
	}
	if sheet.Valid {
		tpl.SheetName = &sheet.String
	}
	var notesPtr *string
	if notes.Valid {
		notesPtr = &notes.String
	}
	return &tpl, notesPtr, createdAt, updatedAt, nil
}

func UpdateTemplate(ctx context.Context, db *sql.DB, tpl *model.Template, now time.Time) error {
	_, err := db.ExecContext(ctx, `
		UPDATE mapping_templates
		SET name = ?, sheet_name = ?, header_row = ?, data_start_row = ?, updated_at = ?
		WHERE id = ?
	`, tpl.Name, tpl.SheetName, tpl.HeaderRow, tpl.DataStartRow, now, tpl.ID)
	return err
}

type RuleRow struct {
	ID           string
	TemplateID   string
	SourceType   string
	SourceKey    string
	TargetID     string
	TargetLabel  string
	CanonicalKey *string
	Transform    *string
	Required     bool
	Priority     int
	Evidence     *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ListRuleRowsByTemplateID(ctx context.Context, db *sql.DB, templateID string) ([]*RuleRow, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, template_id, source_type, source_key, target_id, target_label, canonical_key, transform,
		       required, priority, evidence, created_at, updated_at
		FROM mapping_rules
		WHERE template_id = ?
		ORDER BY priority DESC, created_at ASC
	`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*RuleRow
	for rows.Next() {
		var canonical sql.NullString
		var transform sql.NullString
		var evidence sql.NullString
		var createdAt time.Time
		var updatedAt time.Time

		rr := RuleRow{}
		if err := rows.Scan(
			&rr.ID,
			&rr.TemplateID,
			&rr.SourceType,
			&rr.SourceKey,
			&rr.TargetID,
			&rr.TargetLabel,
			&canonical,
			&transform,
			&rr.Required,
			&rr.Priority,
			&evidence,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if canonical.Valid {
			rr.CanonicalKey = &canonical.String
		}
		if transform.Valid {
			rr.Transform = &transform.String
		}
		if evidence.Valid {
			rr.Evidence = &evidence.String
		}
		rr.CreatedAt = createdAt
		rr.UpdatedAt = updatedAt
		out = append(out, &rr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func InsertRule(ctx context.Context, db *sql.DB, rr *RuleRow, now time.Time) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO mapping_rules (
			id, template_id, source_type, source_key, target_id, target_label, canonical_key, transform,
			required, priority, evidence, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rr.ID, rr.TemplateID, rr.SourceType, rr.SourceKey, rr.TargetID, rr.TargetLabel, rr.CanonicalKey, rr.Transform,
		rr.Required, rr.Priority, rr.Evidence, now, now,
	)
	return err
}

func GetRuleByID(ctx context.Context, db *sql.DB, id string) (*RuleRow, error) {
	var canonical sql.NullString
	var transform sql.NullString
	var evidence sql.NullString
	var createdAt time.Time
	var updatedAt time.Time

	rr := RuleRow{}
	err := db.QueryRowContext(ctx, `
		SELECT id, template_id, source_type, source_key, target_id, target_label, canonical_key, transform,
		       required, priority, evidence, created_at, updated_at
		FROM mapping_rules
		WHERE id = ?
	`, id).Scan(
		&rr.ID,
		&rr.TemplateID,
		&rr.SourceType,
		&rr.SourceKey,
		&rr.TargetID,
		&rr.TargetLabel,
		&canonical,
		&transform,
		&rr.Required,
		&rr.Priority,
		&evidence,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if canonical.Valid {
		rr.CanonicalKey = &canonical.String
	}
	if transform.Valid {
		rr.Transform = &transform.String
	}
	if evidence.Valid {
		rr.Evidence = &evidence.String
	}
	rr.CreatedAt = createdAt
	rr.UpdatedAt = updatedAt
	return &rr, nil
}

func UpdateRule(ctx context.Context, db *sql.DB, rr *RuleRow, now time.Time) error {
	_, err := db.ExecContext(ctx, `
		UPDATE mapping_rules
		SET source_type = ?, source_key = ?, target_label = ?, transform = ?, required = ?, updated_at = ?
		WHERE id = ?
	`, rr.SourceType, rr.SourceKey, rr.TargetLabel, rr.Transform, rr.Required, now, rr.ID)
	return err
}
