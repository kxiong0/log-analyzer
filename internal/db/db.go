package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	common "log-analyzer/internal/common"

	_ "modernc.org/sqlite"
)

const (
	timestampFormat      = "2006-01-02 15:04:05"
	hourTimeFormat       = "2006-01-02 15"
	metricsLookbackHours = 24 * 7
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

	prevTid   string
	prevTidMu sync.Mutex
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
		count INTEGER NOT NULL DEFAULT 0,
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

// Get mean and stddev of hourly counts from the past metricsLookbackHours hours
func (tdb *TemplateDB) GetHourlyStats(tid string) (mean float64, stddev float64, err error) {
	cutoff := time.Now().UTC().Add(-time.Hour * metricsLookbackHours)
	row := tdb.db.QueryRow(`
		SELECT AVG(count), SQRT(
            (COUNT(count) * SUM(count * count) - (SUM(count) * SUM(count))) / ((COUNT(count) - 1) * COUNT(count))
        ) AS stddev
		FROM template_hourly_counts
		WHERE template_id = ? AND hour > ?
	`, tid, cutoff)

	var meanNull sql.NullFloat64
	var stddevNull sql.NullFloat64
	if err = row.Scan(&meanNull, &stddevNull); err != nil {
		slog.Error("Failed to calculate mean and/or stddev of hourly counts")
		return
	}

	mean = meanNull.Float64
	stddev = stddevNull.Float64
	return
}

func (tdb *TemplateDB) GetCurrHourlyCount(tid string) (int, error) {
	currentHour := time.Now().UTC().Format(hourTimeFormat)
	row := tdb.db.QueryRow(`
		SELECT count
		FROM template_hourly_counts
		WHERE template_id = ? AND hour = ?;
	`, tid, currentHour)

	var count int
	if err := row.Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		slog.Error("Failed to get current hourly count")
		return 0, err
	}

	return count, nil
}

// Get probability of transitioning from prevTid to the give tid
func (tdb *TemplateDB) GetTransitionProbability(tid string) (float64, error) {
	// Return 100% expected if there is no prev tid yet
	if len(tdb.prevTid) == 0 {
		return 1.0, nil
	}

	tdb.prevTidMu.Lock()
	defer tdb.prevTidMu.Unlock()

	slog.Debug(fmt.Sprintf("tid %s prev tid %s", tid, tdb.prevTid))
	row := tdb.db.QueryRow(`
		SELECT count * 1.0 /
				SUM(count) OVER (PARTITION BY src_template_id) AS probability
		FROM template_transitions
		WHERE src_template_id = ? AND dst_template_id = ?;
	`, tdb.prevTid, tid)

	var probability float64

	if err := row.Scan(&probability); err != nil {
		return 0.0, err
	}
	return probability, nil
}
