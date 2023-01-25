package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/innermond/dots"
)

type AuthService struct {
	db *DB
}

func NewAuthService(db *DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) CreateAuth(ctx context.Context, auth *dots.Auth) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Commit()
	return nil
}

func findAuth(ctx context.Context, tx *Tx, filter dots.AuthFilter) (_ []*dots.Auth, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.UserID; v != nil {
		where, args = append(where, "user_id = ?"), append(args, *v)
	}
	if v := filter.UserID; v != nil {
		where, args = append(where, "user_id = ?"), append(args, *v)
	}
	if v := filter.Source; v != nil {
		where, args = append(where, "source = ?"), append(args, *v)
	}
	if v := filter.SourceID; v != nil {
		where, args = append(where, "source_id = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		select 
			id, user_id, 
			source, source_id, 
			access_token, refresh_token, 
			expiry, created_at, updated_at,
			count(*) over()
		from "auth"
		where `+strings.Join(where, " and ")+`
		order by id asc
		`+formatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	auths := make([]*dots.Auth, 0)
	for rows.Next() {
		var auth dots.Auth
		var expiry sql.NullTime
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		err := rows.Scan(
			&auth.ID,
			&auth.UserID,
			&auth.Source,
			&auth.SourceID,
			&auth.AccessToken,
			&auth.RefreshToken,
			&expiry,
			&createdAt,
			&updatedAt,
			&n,
		)
		if err != nil {
			return nil, 0, err
		}

		if expiry.Valid {
			v, err := time.Parse(time.RFC3339, expiry.Time.String())
			if err != nil || v.IsZero() {
				return nil, 0, err
			}
			auth.Expiry = v
		}
		auths = append(auths, &auth)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return auths, n, nil
}

func formatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("limit %d offset %d", limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf("limit %d", limit)
	} else if offset > 0 {
		return fmt.Sprintf("offset %d", offset)
	}
	return ""
}
