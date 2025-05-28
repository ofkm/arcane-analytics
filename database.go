package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type InstancesStats struct {
	Total   int                `json:"total"`
	History []InstancesHistory `json:"history"`
}

type InstancesHistory struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func UpsertInstance(db *sql.DB, instanceID, version string) error {
	now := time.Now()

	// Check if instance exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM instances WHERE id = ?)", instanceID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update existing instance
		_, err = db.Exec(
			"UPDATE instances SET last_seen = ?, latest_version = ? WHERE id = ?",
			now, version, instanceID,
		)
	} else {
		// Insert new instance
		_, err = db.Exec(
			"INSERT INTO instances (id, first_seen, last_seen, latest_version) VALUES (?, ?, ?, ?)",
			instanceID, now, now, version,
		)
	}

	return err
}

func GetTotalInstances(db *sql.DB) (int, error) {
	var count int
	// Only count instances that:
	// 1. Are older than 1 day
	// 2. Have been active in the last 2 days
	query := `
		SELECT COUNT(*) 
		FROM instances 
		WHERE first_seen < datetime('now', '-1 day') 
		AND last_seen >= datetime('now', '-2 days')
	`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

func GetInstancesOverTime(db *sql.DB, timeframe string) ([]InstancesHistory, error) {
	var query string

	switch timeframe {
	case "daily":
		// Get daily instance counts for the last 30 days
		// Only include instances that are older than 1 day and were active in the last 2 days
		query = `
		SELECT 
			DATE(first_seen) as date,
			COUNT(*) as daily_new,
			(SELECT COUNT(*) 
			 FROM instances i2 
			 WHERE DATE(i2.first_seen) <= DATE(i1.first_seen)
			 AND i2.first_seen < datetime('now', '-1 day')
			 AND i2.last_seen >= datetime('now', '-2 days')) as cumulative_count
		FROM instances i1
		WHERE first_seen >= datetime('now', '-30 days')
		AND first_seen < datetime('now', '-1 day')
		AND last_seen >= datetime('now', '-2 days')
		GROUP BY DATE(first_seen)
		ORDER BY date
		`
	case "monthly":
		// Get monthly instance counts for all time
		// Only include instances that are older than 1 day and were active in the last 2 days
		query = `
		SELECT 
			strftime('%Y-%m', first_seen) as date,
			COUNT(*) as monthly_new,
			(SELECT COUNT(*) 
			 FROM instances i2 
			 WHERE strftime('%Y-%m', i2.first_seen) <= strftime('%Y-%m', i1.first_seen)
			 AND i2.first_seen < datetime('now', '-1 day')
			 AND i2.last_seen >= datetime('now', '-2 days')) as cumulative_count
		FROM instances i1
		WHERE first_seen < datetime('now', '-1 day')
		AND last_seen >= datetime('now', '-2 days')
		GROUP BY strftime('%Y-%m', first_seen)
		ORDER BY date
		`
	default:
		return nil, fmt.Errorf("invalid timeframe: %s. Use 'daily' or 'monthly'", timeframe)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chartData []InstancesHistory
	for rows.Next() {
		var date string
		var newCount, cumulativeCount int

		err := rows.Scan(&date, &newCount, &cumulativeCount)
		if err != nil {
			return nil, err
		}

		chartData = append(chartData, InstancesHistory{
			Date:  date,
			Count: cumulativeCount,
		})
	}

	return chartData, nil
}

func initDB() (*sql.DB, error) {
	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", "./data/pocket-id-analytics.db")
	if err != nil {
		return nil, err
	}

	// Create instances table
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS instances (
        id TEXT PRIMARY KEY,
        first_seen DATETIME NOT NULL,
        last_seen DATETIME NOT NULL,
        latest_version TEXT NOT NULL
    );
    
    CREATE INDEX IF NOT EXISTS idx_first_seen ON instances(first_seen);
    CREATE INDEX IF NOT EXISTS idx_last_seen ON instances(last_seen);
    `

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
