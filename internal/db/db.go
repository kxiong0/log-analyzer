package db

import (
	"database/sql"
	"log/slog"
	"strings"

	common "log-analyzer/internal/common"

	_ "modernc.org/sqlite"
)

func NewTemplateDB(dataSourceName string) (*TemplateDB, error) {
	slog.Info("Connecting to template DB")
	sdb := TemplateDB{}

	// Open DB
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	sdb.db = db

	// Init table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS templates(
		uuid TEXT PRIMARY KEY,
		token_count INTEGER NOT NULL,
		template_text TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return nil, err
	}
	slog.Info("Connected to template DB")

	return &sdb, nil
}

type TemplateDB struct {
	db *sql.DB
}

func (tdb *TemplateDB) SaveTemplate(t common.Template) error {
	_, err := tdb.db.Exec(`
		INSERT INTO templates (uuid, token_count, template_text)
		VALUES(?, ?, ?)
	`, t.ID, len(t.Tokens), strings.Join(t.Tokens, " "))
	if err != nil {
		return err
	}

	return nil
}

func (tdb *TemplateDB) GetAllTemplates() ([]common.Template, error) {
	rows, err := tdb.db.Query("SELECT uuid, template_text FROM templates;")
	if err != nil {
		return nil, sql.ErrConnDone
	}

	templates := []common.Template{}
	for rows.Next() {
		var uuid string
		var template_text string

		err := rows.Scan(&uuid, &template_text)
		if err != nil {
			slog.Error("Failed to read template row into vars")
			continue
		}

		t := common.Template{
			ID:     uuid,
			Tokens: strings.Fields(template_text),
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (sdb *TemplateDB) IncreaseTemplateCount(uuid string) {

}
