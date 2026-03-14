package store

import "database/sql"

type Server struct {
	ProductID  string
	ID         string
	Name       string
	Host       string
	Port       int
	User       string
	Sudo       bool
	SSHKeyPath string
}

type MonitoredFile struct {
	ID          string
	ProductID   string
	ServerID    string
	DestPath    string
	RepoRelPath string
}

type FileState struct {
	ID              int64
	MonitoredFileID string
	Status          string
	LastCheckedAt   string
	Diff            sql.NullString
	Error           sql.NullString
}
