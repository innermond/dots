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

	others, _, err := findAuth(ctx, tx, dots.AuthFilter{Source: &auth.Source, SourceID: &auth.SourceID})

	if err != nil {
		return err
	}
	if len(others) == 0 {
		if auth.UserID == 0 && auth.User != nil {
			uu, _, err := findUser(ctx, tx, dots.UserFilter{Email: &auth.User.Email, Limit: 1})
			if err != nil {
				return fmt.Errorf("postgres.auth: cannot find user by email %w", err)
			}
			if len(uu) == 0 {
				err = createUser(ctx, tx, auth.User)
				if err != nil {
					return fmt.Errorf("postgres.auth: cannot create new user %w", err)
				}
				auth.UserID = auth.User.ID
			} else {
				auth.User = uu[0]
			}
		}
	} else {
		other := others[0]
		other, err = updateAuth(ctx, tx, other.ID, auth.AccessToken, auth.RefreshToken, auth.Expiry)
		if err != nil {
			return err
		}
		// TODO attach user
		*auth = *other
		return tx.Commit()
	}

	err = createAuth(ctx, tx, auth)
	if err != nil {
		return fmt.Errorf("postgres.auth: create auth: %s", err)
	}
	//u, _, err := findUser(ctx, tx, dots.UserFilter{ID: &auth

	auth.UserID = auth.User.ID

	tx.Commit()
	return nil
}

func createAuth(ctx context.Context, tx *Tx, auth *dots.Auth) (err error) {
	if err = auth.Validate(); err != nil {
		return err
	}

	auth.CreatedAt = tx.now
	auth.UpdatedAt = auth.CreatedAt
	var expiry *time.Time
	if auth.Expiry != nil {
		*expiry, err = time.Parse(time.RFC3339, auth.Expiry.String())
		if err != nil {
			return err
		}
	}

	sqlstr := `INSERT INTO "auth" (
		user_id, 
		source, source_id, 
		access_token, refresh_token, 
		expiry,
		created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id`

	var id int = 0
	err = tx.QueryRowContext(ctx, sqlstr,
		auth.UserID,
		auth.Source, auth.SourceID,
		auth.AccessToken, auth.RefreshToken,
		expiry,
		auth.CreatedAt, auth.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return err
	}
	auth.ID = id
	if err := attachAuthUser(ctx, tx, auth); err != nil {
		return err
	}
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
	for inx, v := range where {
		if !strings.Contains(v, "?") {
			continue
		}
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx), 1)
		where[inx] = v
	}

	rows, err := tx.QueryContext(ctx, `
		select 
			id, user_id, 
			source, source_id, 
			access_token, refresh_token, 
			expiry, created_at, updated_at,
			count(*) over(),
			(select coalesce(jsonb_agg(u.*), '[{}]'::jsonb) from "user" u where u.id = user_id limit 1)::json->0 "user"
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
		var expiry *time.Time
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var u dots.User
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
			&u,
		)
		if err == sql.ErrNoRows {
			return nil, 0, dots.Errorf(dots.ENOTFOUND, "auth not found")
		}
		if err != nil {
			return nil, 0, err
		}
		if expiry != nil {
			auth.Expiry = expiry
		}
		auth.CreatedAt = timeRFC3339(createdAt)
		auth.UpdatedAt = timeRFC3339(updatedAt)
		auths = append(auths, &auth)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return auths, n, nil
}

func updateAuth(
	ctx context.Context,
	tx *Tx,
	id int,
	accessToken string,
	refreshToken string,
	expiry *time.Time) (*dots.Auth, error) {
	aa, _, err := findAuth(ctx, tx, dots.AuthFilter{ID: &id, Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(aa) == 0 {
		return nil, dots.Errorf(dots.ENOTFOUND, "auth not found")
	}
	auth := aa[0]
	auth.AccessToken = accessToken
	auth.RefreshToken = refreshToken
	if expiry != nil {
		rfctime, err := time.Parse(time.RFC3339, expiry.Format(time.RFC3339))
		if err != nil {
			return nil, err
		}
		auth.Expiry = &rfctime
	}
	auth.UpdatedAt = tx.now

	err = auth.Validate()
	if err != nil {
		return auth, err
	}

	_, err = tx.ExecContext(ctx, `
		update "auth"
		set access_token = $1, refresh_token = $2, expiry = $3, updated_at = $4
	`, auth.AccessToken, auth.RefreshToken, auth.Expiry, auth.UpdatedAt)
	if err != nil {
		return auth, err
	}
	return auth, nil
}

func deleteAuth(ctx context.Context, tx *Tx, id int) error {
	aa, _, err := findAuth(ctx, tx, dots.AuthFilter{ID: &id})
	if err != nil {
		return err
	}
	if len(aa) == 0 {
		return dots.Errorf(dots.ENOTFOUND, "auth not found")
	}

	_, err = tx.ExecContext(ctx, `delete from "auth" where id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres.auth: cannot delete auth: %w", err)
	}

	return nil
}

func attachAuthUser(ctx context.Context, tx *Tx, a *dots.Auth) (err error) {
	a.User, err = findUserByID(ctx, tx, a.UserID)
	if err != nil {
		return err
	}
	return nil
}
