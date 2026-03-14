package http

import (
	"database/sql"
	"fmt"
	"net/http"

	"mymodule/internal/store"
)

func NewServer(db *sql.DB) *http.Server {
	smux := http.NewServeMux()
	smux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		states, err := store.GetAllFileStates(db)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get file states: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, "<h1>File States</h1>")
		fmt.Fprintln(w, "<table border=\"1\">")
		fmt.Fprintln(w, "<tr><th>Monitored File ID</th><th>Status</th><th>Last Checked</th><th>Error</th></tr>")
		for _, s := range states {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", s.MonitoredFileID, s.Status, s.LastCheckedAt, s.Error.String)
		}
		fmt.Fprintln(w, "</table>")
	})

	return &http.Server{
		Addr:    ":8080",
		Handler: smux,
	}
}
