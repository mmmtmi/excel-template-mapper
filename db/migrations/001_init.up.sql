CREATE TABLE IF NOT EXISTS mapping_templates (
  id            VARCHAR(36)  NOT NULL,
  name          VARCHAR(255) NOT NULL,
  target        VARCHAR(255) NOT NULL,
  sheet_name    VARCHAR(255) NULL,
  header_row    INT          NOT NULL,
  data_start_row INT         NOT NULL,
  notes         VARCHAR(1024) NULL,
  created_at    DATETIME(6)  NOT NULL,
  updated_at    DATETIME(6)  NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_mapping_templates_name (name)
);

CREATE TABLE IF NOT EXISTS mapping_rules (
  id            VARCHAR(36)  NOT NULL,
  template_id   VARCHAR(36)  NOT NULL,
  source_type   VARCHAR(16)  NOT NULL,
  source_key    VARCHAR(255) NOT NULL,
  target_id     VARCHAR(36)  NOT NULL,
  target_label  VARCHAR(255) NOT NULL,
  canonical_key VARCHAR(255) NULL,
  transform     VARCHAR(255) NULL,
  required      TINYINT(1)   NOT NULL,
  priority      INT          NOT NULL,
  evidence      TEXT         NULL,
  created_at    DATETIME(6)  NOT NULL,
  updated_at    DATETIME(6)  NOT NULL,
  PRIMARY KEY (id),
  KEY idx_mapping_rules_template_id (template_id),
  CONSTRAINT fk_mapping_rules_template
    FOREIGN KEY (template_id) REFERENCES mapping_templates(id)
    ON DELETE CASCADE
);