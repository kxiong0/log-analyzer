package db

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// Update template count stat
func (tdb *TemplateDB) CountTemplate(uuid string) error {
	// Insert new rows for new template
	_, err := tdb.db.Exec(`
		INSERT OR IGNORE INTO template_stats (template_id) VALUES (?);
	`, uuid)
	if err != nil {
		return fmt.Errorf("failed to insert new template stats row: %s", err)
	}

	// Calculate IAT stats
	count, iatMean, iatStddev, iatLastTimestamp, err := tdb.GetIATStats(uuid)
	if err != nil {
		return fmt.Errorf("failed to get IAT stats: %s", err)
	}

	newMean, newStddev, err := calculateIAT(
		iatLastTimestamp, iatMean, iatStddev, count,
	)
	if err != nil {
		return fmt.Errorf("failed to calculate IAT stats: %s", err)
	}

	// Update stats
	currTs := time.Now().UTC().Format(TimestampFormat)
	_, err = tdb.db.Exec(`
		UPDATE template_stats
		SET total_count = total_count + 1,
			last_seen = ?,
			iat_mean = ?,
			iat_stddev = ?,
			iat_last_timestamp = ?
		WHERE template_id = ?
	`, currTs, newMean, newStddev, currTs, uuid)
	if err != nil {
		return fmt.Errorf("failed to update stats: %s", err)
	}

	return nil
}

// Fetch IAT stats for uuid
func (tdb *TemplateDB) GetIATStats(uuid string) (count int, mean float64, stddev float64, lastTs string, err error) {
	row := tdb.db.QueryRow(`
		SELECT total_count, iat_mean, iat_stddev, iat_last_timestamp
		FROM template_stats
		WHERE template_id = ?;
	`, uuid)

	var iatLastTimestamp sql.NullString
	err = row.Scan(&count, &mean, &stddev, &iatLastTimestamp)
	if err != nil {
		return
	}

	lastTs = iatLastTimestamp.String
	return
}

func calculateIAT(lastTimestamp string, mean float64, stddev float64, count int) (float64, float64, error) {
	// No previous timestamp
	if len(lastTimestamp) == 0 {
		return mean, stddev, nil
	}

	// Calculate new IAT stats
	lastTime, err := time.Parse(TimestampFormat, lastTimestamp)
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
	iat := time.Now().UTC().Sub(lastTime).Seconds()

	// Calculate new mean
	newMean := 0.0
	newStddev := 0.0

	delta := iat - mean
	newMean = mean + delta/float64(newCount)
	delta2 := iat - newMean
	newM2 := oldM2 + delta*delta2

	// Calculate stddev
	if count >= 2 {
		variance := newM2 / float64(count-1)
		newStddev = math.Sqrt(variance)
	}

	return newMean, newStddev, nil
}

// Update hourly count for template
func (tdb *TemplateDB) CountTemplateHourly(uuid string) error {
	err := tdb.InsertHourlyRow(uuid)
	if err != nil {
		return err
	}

	currentHour := time.Now().UTC().Format(hourTimeFormat)
	_, err = tdb.db.Exec(`
		UPDATE template_hourly_counts
		SET count = count + 1
		WHERE template_id = ? AND hour = ?
	`, uuid, currentHour)
	if err != nil {
		return err
	}

	return nil
}

// Insert new hourly count row for new template
func (tdb *TemplateDB) InsertHourlyRow(uuid string) error {
	currentHour := time.Now().UTC().Format(hourTimeFormat)
	_, err := tdb.db.Exec(`
		INSERT OR IGNORE INTO template_hourly_counts (template_id, hour)
		VALUES (?, ?);
	`, uuid, currentHour)
	if err != nil {
		return err
	}
	return nil
}

// Increment count on template transition from template `prevId` to `uuid`
func (tdb *TemplateDB) CountTransition(uuid string, podId string) error {
	// Update prevId to next template id before returning
	defer func() {
		tdb.prevTidsMu.Lock()
		tdb.prevTids[podId] = uuid
		tdb.prevTidsMu.Unlock()
	}()

	prevTid := tdb.prevTids[podId]

	// Ignore empty IDs
	if len(prevTid) == 0 || len(uuid) == 0 {
		return nil
	}

	tdb.prevTidsMu.RLock()
	defer tdb.prevTidsMu.RUnlock()

	_, err := tdb.db.Exec(`
		INSERT OR IGNORE INTO template_transitions (src_template_id, dst_template_id, pod_id)
		VALUES (?, ?, ?);
	`, prevTid, uuid, podId)
	if err != nil {
		return err
	}

	_, err = tdb.db.Exec(`
		UPDATE template_transitions
		SET count = count + 1
		WHERE src_template_id = ? AND dst_template_id = ? AND pod_id = ?
	`, prevTid, uuid, podId)

	if err != nil {
		return err
	}

	return nil
}
