package period

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Selection struct {
	Label         string
	SemesterID    string
	TahunAjaranID int64
}

func Resolve(ctx context.Context, pool *pgxpool.Pool, label string) (Selection, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return Selection{}, errors.New("SEMESTER is required (example: 2026/2027 Gasal)")
	}
	rows, err := pool.Query(ctx, `SELECT btrim(semester_id), tahun_ajaran_id::bigint FROM ref.semester WHERE expired_date IS NULL AND lower(btrim(nama)) = lower($1)`, label)
	if err != nil {
		return Selection{}, fmt.Errorf("resolve SEMESTER %q: %w", label, err)
	}
	defer rows.Close()
	var selected Selection
	selected.Label = label
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return Selection{}, fmt.Errorf("read SEMESTER %q: %w", label, err)
		}
		return Selection{}, fmt.Errorf("SEMESTER %q tidak ditemukan di ref.semester", label)
	}
	if err := rows.Scan(&selected.SemesterID, &selected.TahunAjaranID); err != nil {
		return Selection{}, fmt.Errorf("scan SEMESTER %q: %w", label, err)
	}
	if rows.Next() {
		return Selection{}, fmt.Errorf("SEMESTER %q cocok dengan lebih dari satu baris ref.semester", label)
	}
	if err := rows.Err(); err != nil {
		return Selection{}, fmt.Errorf("read SEMESTER %q: %w", label, err)
	}
	return selected, nil
}
