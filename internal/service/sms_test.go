package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ssddi456/sms_server/internal/service"
)

type WebhookRequest struct {
	From        string    `json:"from"`
	ContactName string    `json:"contact_name"`
	PhoneArea   string    `json:"phone_area"`
	SMS         string    `json:"sms"`
	SimSlot     string    `json:"sim_slot"`
	SimSubID    string    `json:"sim_sub_id"`
	DeviceName  string    `json:"device_name"`
	ReceiveTime time.Time `json:"receive_time"`
}

type SMSRecord struct {
	ID            int       `json:"id"`
	From          string    `json:"from"`
	ContactName   string    `json:"contact_name,omitempty"`
	PhoneArea     string    `json:"phone_area,omitempty"`
	SMS           string    `json:"sms"`
	SimSlot       string    `json:"sim_slot,omitempty"`
	SimSubID      string    `json:"sim_sub_id,omitempty"`
	DeviceName    string    `json:"device_name,omitempty"`
	ReceiveTime   time.Time `json:"receive_time"`
	ExtractedCode string    `json:"extracted_code,omitempty"`
	PhoneNumber   string    `json:"phone_number,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dataDir, "sms.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sms_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_phone TEXT NOT NULL,
			contact_name TEXT,
			phone_area TEXT,
			sms TEXT NOT NULL,
			sim_slot TEXT,
			sim_sub_id TEXT,
			device_name TEXT,
			receive_time TEXT NOT NULL,
			extracted_code TEXT,
			phone_number TEXT,
			created_at TEXT NOT NULL
		);`); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) SaveWebhook(req WebhookRequest) (SMSRecord, error) {
	if strings.TrimSpace(req.SMS) == "" {
		return SMSRecord{}, fmt.Errorf("sms is required")
	}
	if strings.TrimSpace(req.From) == "" {
		return SMSRecord{}, fmt.Errorf("from is required")
	}

	receiveTime := req.ReceiveTime
	if receiveTime.IsZero() {
		receiveTime = time.Now().UTC()
	}
	code := service.ExtractCode(req.SMS)
	phone := service.ExtractPhoneNumber(req.SimSlot)
	createdAt := time.Now().UTC()

	result, err := s.db.Exec(
		`INSERT INTO sms_records (from_phone, contact_name, phone_area, sms, sim_slot, sim_sub_id, device_name, receive_time, extracted_code, phone_number, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.From,
		req.ContactName,
		req.PhoneArea,
		req.SMS,
		req.SimSlot,
		req.SimSubID,
		req.DeviceName,
		receiveTime.UTC().Format(time.RFC3339),
		code,
		phone,
		createdAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return SMSRecord{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return SMSRecord{}, err
	}
	return SMSRecord{
		ID:            int(id),
		From:          req.From,
		ContactName:   req.ContactName,
		PhoneArea:     req.PhoneArea,
		SMS:           req.SMS,
		SimSlot:       req.SimSlot,
		SimSubID:      req.SimSubID,
		DeviceName:    req.DeviceName,
		ReceiveTime:   receiveTime,
		ExtractedCode: code,
		PhoneNumber:   phone,
		CreatedAt:     createdAt,
	}, nil
}

func (s *Store) History(limit, offset int) ([]SMSRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(`
		SELECT id, from_phone, contact_name, phone_area, sms, sim_slot, sim_sub_id, device_name, receive_time, extracted_code, phone_number, created_at
		FROM sms_records
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]SMSRecord, 0)
	for rows.Next() {
		var r SMSRecord
		var receiveTime, createdAt string
		if err := rows.Scan(&r.ID, &r.From, &r.ContactName, &r.PhoneArea, &r.SMS, &r.SimSlot, &r.SimSubID, &r.DeviceName, &receiveTime, &r.ExtractedCode, &r.PhoneNumber, &createdAt); err != nil {
			return nil, err
		}
		r.ReceiveTime, _ = time.Parse(time.RFC3339, receiveTime)
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		records = append(records, r)
	}
	return records, nil
}
