package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	_ "modernc.org/sqlite"
	"net"
	"net/http"
	"os"
	"strings"
)

type device struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	MAC  string `json:"mac"`
}

func main() {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, err := sql.Open("sqlite", getenv("DATABASE_PATH", "/data/wol.db"))
	if err != nil {
		l.Error("db", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS devices(id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT NOT NULL,mac TEXT NOT NULL UNIQUE)`); err != nil {
		l.Error("schema", "error", err)
		os.Exit(1)
	}
	if err = migrateDevices(db); err != nil {
		l.Error("migrate schema", "error", err)
		os.Exit(1)
	}
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /api/devices", func(w http.ResponseWriter, r *http.Request) { list(w, db) })
	m.HandleFunc("POST /api/devices", func(w http.ResponseWriter, r *http.Request) { create(w, r, db) })
	m.HandleFunc("POST /api/devices/{id}/wake", func(w http.ResponseWriter, r *http.Request) { wake(w, r, db) })
	m.HandleFunc("DELETE /api/devices/{id}", func(w http.ResponseWriter, r *http.Request) {
		db.Exec("DELETE FROM devices WHERE id=?", r.PathValue("id"))
		w.WriteHeader(204)
	})
	l.Info("starting API")
	if err := http.ListenAndServe(":"+getenv("PORT", "8080"), cors(m)); err != nil {
		l.Error("server stopped", "error", err)
	}
}
func list(w http.ResponseWriter, db *sql.DB) {
	rows, e := db.Query("SELECT id,name,mac FROM devices ORDER BY name")
	if e != nil {
		write(w, 500, map[string]string{"error": e.Error()})
		return
	}
	defer rows.Close()
	out := []device{}
	for rows.Next() {
		var d device
		if rows.Scan(&d.ID, &d.Name, &d.MAC) == nil {
			out = append(out, d)
		}
	}
	write(w, 200, out)
}

func migrateDevices(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(devices)")
	if err != nil {
		return err
	}
	legacy := false
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == "host" || name == "port" {
			legacy = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if !legacy {
		return nil
	}
	_, err = db.Exec(`
        BEGIN;
        CREATE TABLE devices_new(id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT NOT NULL,mac TEXT NOT NULL UNIQUE);
        INSERT INTO devices_new(id,name,mac) SELECT id,name,mac FROM devices;
        DROP TABLE devices;
        ALTER TABLE devices_new RENAME TO devices;
        COMMIT;
    `)
	return err
}
func create(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var d device
	if json.NewDecoder(r.Body).Decode(&d) != nil {
		write(w, 400, map[string]string{"error": "el cuerpo no contiene JSON válido"})
		return
	}
	d.Name = strings.TrimSpace(d.Name)
	d.MAC = strings.TrimSpace(d.MAC)
	parsedMAC, macErr := net.ParseMAC(d.MAC)
	if d.Name == "" {
		write(w, 400, map[string]string{"error": "el nombre es obligatorio"})
		return
	}
	if macErr != nil || len(parsedMAC) != 6 {
		write(w, 400, map[string]string{"error": "MAC inválida; usa AA:BB:CC:DD:EE:FF"})
		return
	}
	normalizedMAC := strings.ToUpper(parsedMAC.String())
	_, e := db.Exec("INSERT INTO devices(name,mac) VALUES(?,?)", strings.TrimSpace(d.Name), normalizedMAC)
	if e != nil {
		if strings.Contains(e.Error(), "UNIQUE constraint failed") {
			write(w, 409, map[string]string{"error": "dispositivo ya registrado"})
			return
		}
		write(w, 500, map[string]string{"error": e.Error()})
		return
	}
	write(w, 201, map[string]string{"status": "created"})
}
func wake(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var mac string
	if e := db.QueryRow("SELECT mac FROM devices WHERE id=?", r.PathValue("id")).Scan(&mac); e != nil {
		write(w, 404, map[string]string{"error": "no encontrado"})
		return
	}
	b, e := hex.DecodeString(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))
	if e != nil || len(b) != 6 {
		write(w, 500, map[string]string{"error": "MAC inválida"})
		return
	}
	p := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	for i := 0; i < 16; i++ {
		p = append(p, b...)
	}
	c, e := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4bcast, Port: 9})
	if e == nil {
		_, e = c.Write(p)
		c.Close()
	}
	if e != nil {
		write(w, 500, map[string]string{"error": e.Error()})
		return
	}
	write(w, 202, map[string]string{"status": "wake packet sent"})
}
func write(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	json.NewEncoder(w).Encode(v)
}
func cors(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		n.ServeHTTP(w, r)
	})
}
func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
