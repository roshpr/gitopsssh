package store

import (
	"database/sql"
	"fmt"
)

func GetAllServers(db *sql.DB) ([]Server, error) {
	rows, err := db.Query("SELECT product_id, id, name, host, port, user, sudo FROM servers")
	if err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var s Server
		if err := rows.Scan(&s.ProductID, &s.ID, &s.Name, &s.Host, &s.Port, &s.User, &s.Sudo); err != nil {
			return nil, fmt.Errorf("failed to scan server row: %w", err)
		}
		servers = append(servers, s)
	}

	return servers, nil
}
