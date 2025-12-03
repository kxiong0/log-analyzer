package db

import (
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
	currTs := time.Now().UTC().Format(timestampFormat)
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

// Update hourly count for template
func (tdb *TemplateDB) CountTemplateHourly(uuid string) error {
	// Insert new rows for new template
	currentHour := time.Now().UTC().Format(hourTimeFormat)
	_, err := tdb.db.Exec(`
		INSERT OR IGNORE INTO template_hourly_counts (template_id, hour)
		VALUES (?, ?);
	`, uuid, currentHour)
	if err != nil {
		return err
	}

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

// Increment count on template transition from template `prevId` to `uuid`
func (tdb *TemplateDB) CountTransition(uuid string) error {
	// Update prevId to next template id before returning
	defer func() { tdb.prevTid = uuid }()

	// Ignore empty IDs
	if len(tdb.prevTid) == 0 || len(uuid) == 0 {
		return nil
	}

	tdb.prevTidMu.Lock()
	defer tdb.prevTidMu.Unlock()

	_, err := tdb.db.Exec(`
		INSERT OR IGNORE INTO template_transitions (src_template_id, dst_template_id)
		VALUES (?, ?);
	`, tdb.prevTid, uuid)
	if err != nil {
		return err
	}

	_, err = tdb.db.Exec(`
		UPDATE template_transitions
		SET count = count + 1
		WHERE src_template_id = ? AND dst_template_id = ?
	`, tdb.prevTid, uuid)

	if err != nil {
		return err
	}

	return nil
}
