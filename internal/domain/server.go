package domain

import "time"

type Server struct {
	ID             string
	ProductID      string
	Name           string
	Host           string
	Port           int
	User           string
	Sudo           bool
	SSHKeyPath     string
	LastPollAt     *time.Time
	LastError      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	MonitoredFiles []MonitoredFile
}

type MonitoredFile struct {
	ID          string
	ProductID   string
	ServerID    string
	DestPath    string
	RepoRelPath string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
