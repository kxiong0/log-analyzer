package db

import (
	"database/sql"
	"log/slog"
	"math"
	"strings"
	"time"

	common "log-analyzer/internal/common"

	_ "modernc.org/sqlite"
)

const (
	timestampFormat = "2006-01-02 15:04:05"
	hourTimeFormat  = "2006-01-02 15"
)

func NewTemplateDB(dataSourceName string) (*TemplateDB, error) {
	slog.Info("Connecting to template DB")
	tdb := TemplateDB{}

	// Open DB
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	tdb.db = db
	slog.Info("Connected to template DB")

	// Init tables
	tdb.InitTables()

	return &tdb, nil
}

type TemplateDB struct {
	db *sql.DB
}

// Create DB tables if they don't exist
func (tdb *TemplateDB) InitTables() error {
	slog.Debug("Creating template tables...")

	_, err := tdb.db.Exec(`
	CREATE TABLE IF NOT EXISTS templates(
		uuid TEXT PRIMARY KEY,
		token_count INTEGER NOT NULL,
		template_text TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return err
	}

	_, err = tdb.db.Exec(`
	CREATE TABLE IF NOT EXISTS template_transitions (
		src_template_id TEXT NOT NULL,
		dst_template_id TEXT NOT NULL,
		count INTEGER NOT NULL DEFAULT 1,
		last_seen TEXT DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY (src_template_id, dst_template_id),

		FOREIGN KEY (src_template_id) REFERENCES templates(uuid),
		FOREIGN KEY (dst_template_id) REFERENCES templates(uuid)
	);`)
	if err != nil {
		return err
	}

	_, err = tdb.db.Exec(`
	CREATE TABLE IF NOT EXISTS template_stats (
		template_id TEXT PRIMARY KEY,
		
		-- Aggregated frequency over time
		total_count INTEGER NOT NULL DEFAULT 0,
		last_seen TEXT,
		
		-- Interarrival time statistics (in seconds)
		iat_mean REAL NOT NULL DEFAULT 0.0,            -- mean interarrival time
		iat_stddev REAL NOT NULL DEFAULT 0.0,          -- std dev of interarrival time
		iat_last_timestamp TEXT,

		FOREIGN KEY (template_id) REFERENCES templates(uuid)
	);`)
	if err != nil {
		return err
	}

	_, err = tdb.db.Exec(`
	CREATE TABLE IF NOT EXISTS template_hourly_counts (
		template_id TEXT NOT NULL,
		hour TEXT NOT NULL,         -- "2025-11-24 13"
		count INTEGER NOT NULL DEFAULT 0,

		PRIMARY KEY (template_id, hour)
	);`)
	if err != nil {
		return err
	}

	slog.Debug("All tables created successfully")
	return nil
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

// Get all templates from DB and return a map of token count -> Templates
func (tdb *TemplateDB) GetAllTemplates() (map[int][]common.Template, error) {
	rows, err := tdb.db.Query("SELECT uuid, token_count, template_text FROM templates;")
	if err != nil {
		return nil, sql.ErrConnDone
	}

	templates := make(map[int][]common.Template)
	for rows.Next() {
		var uuid string
		var token_count int
		var template_text string

		err := rows.Scan(&uuid, &token_count, &template_text)
		if err != nil {
			slog.Error("Failed to read template row into vars")
			continue
		}

		t := common.Template{
			ID:     uuid,
			Tokens: strings.Fields(template_text),
		}
		templates[token_count] = append(templates[token_count], t)
	}
	return templates, nil
}

func (tdb *TemplateDB) CountTemplate(uuid string) error {
	// Insert new rows for new template
	currentHour := time.Now().UTC().Format(hourTimeFormat)
	_, err := tdb.db.Exec(`
		INSERT OR IGNORE INTO template_stats (template_id) VALUES (?);
		INSERT OR IGNORE INTO template_hourly_counts (template_id, hour)
		VALUES (?, ?);
	`, uuid, uuid, currentHour)
	if err != nil {
		return err
	}

	// Calculate IAT stats
	row := tdb.db.QueryRow(`
		SELECT total_count, iat_mean, iat_stddev, iat_last_timestamp
		FROM template_stats
		WHERE template_id = ?;
	`, uuid)

	var count int
	var iatMean float64
	var iatStddev float64
	var iatLastTimestamp interface{}
	err = row.Scan(&count, &iatMean, &iatStddev, &iatLastTimestamp)
	if err != nil {
		return err
	}

	newMean, newStddev, err := calculateIAT(
		iatLastTimestamp, iatMean, iatMean, count,
	)
	if err != nil {
		return err
	}

	// Update stats
	_, err = tdb.db.Exec(`
		UPDATE template_stats
		SET total_count = total_count + 1,
			iat_mean = ?,
			iat_stddev = ?,
			iat_last_timestamp = ?
		WHERE template_id = ?
	`, newMean, newStddev, time.Now().UTC().Format(timestampFormat), uuid)
	if err != nil {
		return err
	}
	return nil
}

func calculateIAT(lastTimestamp interface{}, mean float64, stddev float64, count int) (float64, float64, error) {
	newMean := 0.0
	newStddev := 0.0
	currTime := time.Now().UTC()
	if lastTimestamp == nil {
		// No previous data
		lastTimestamp = currTime.Format(timestampFormat)
	} else {
		// Calculate new IAT stats
		lastTime, err := time.Parse(timestampFormat, lastTimestamp.(string))
		if err != nil {
			return 0.0, 0.0, err
		}

		// Recover old M2
		var oldM2 float64
		if count >= 2 {
			oldM2 = stddev * stddev * float64(count-1)
		}

		newCount := count + 1

		// Welford update
		iat := currTime.Sub(lastTime).Seconds()

		// Calculate new mean
		delta := iat - mean
		newMean = mean + delta/float64(newCount)
		delta2 := iat - newMean
		newM2 := oldM2 + delta*delta2

		// Calculate stddev
		if count >= 2 {
			variance := newM2 / float64(count-1)
			newStddev = math.Sqrt(variance)
		}
	}
	return newMean, newStddev, nil
}
