// Package userpg
package userpg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/port"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *userDomain.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// =====================
	// USERS
	// =====================
	_, err = tx.Exec(ctx, `
		INSERT INTO users (
			id, username, email,
			first_name, last_name,
			avatar_url,
			status, 
			last_seen_at,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		user.ID(),
		user.Username().String(),
		user.Email().String(),
		user.FirstName(),
		user.LastName(),
		user.AvatarURL(),
		user.Status(),
		user.LastSeenAt(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)
	if err != nil {
		return mapPGError(err)
	}

	// =====================
	// PROFILE
	// =====================
	profile := user.Profile()

	_, err = tx.Exec(ctx, `
		INSERT INTO user_profiles (
			user_id, bio, location, website, birthday,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
	`,
		user.ID(),
		profile.Bio(),
		profile.Location(),
		profile.Website(),
		profile.Birthday(),
		profile.CreatedAt(),
		profile.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert profile: %w", err)
	}

	// =====================
	// SETTINGS
	// =====================
	settings := user.Settings()

	_, err = tx.Exec(ctx, `
		INSERT INTO user_settings (
			user_id, is_public, allow_notifications, theme,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6)
	`,
		user.ID(),
		settings.IsPublic(),
		settings.AllowNotifications(),
		settings.Theme(),
		settings.CreatedAt(),
		settings.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert settings: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *userRepository) FindByID(ctx context.Context, userID uuid.UUID) (*userDomain.User, error) {
	query := `
	SELECT 
		u.id,
		u.username,
		u.email,
		u.first_name,
		u.last_name,
		u.avatar_url,
		u.status,
		u.last_seen_at,
		u.created_at,
		u.updated_at,
		u.deleted_at,

		p.bio,
		p.location,
		p.website,
		p.birthday,
		p.created_at,
		p.updated_at,

		s.is_public,
		s.allow_notifications,
		s.theme,
		s.created_at,
		s.updated_at 
		FROM users u
	LEFT JOIN user_profiles p ON u.id = p.user_id
	LEFT JOIN user_settings s ON u.id = s.user_id
	WHERE u.id = $1
	`

	row := r.db.QueryRow(ctx, query, userID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userDomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email userDomain.Email) (*userDomain.User, error) {
	query := `
	SELECT 
		u.id,
		u.username,
		u.email,
		u.first_name,
		u.last_name,
		u.avatar_url,
		u.status,
		u.last_seen_at,
		u.created_at,
		u.updated_at,
		u.deleted_at,

		p.bio,
		p.location,
		p.website,
		p.birthday,
		p.created_at,
		p.updated_at,

		s.is_public,
		s.allow_notifications,
		s.theme,
		s.created_at,
		s.updated_at 
		FROM users u
	LEFT JOIN user_profiles p ON u.id = p.user_id
	LEFT JOIN user_settings s ON u.id = s.user_id
	WHERE u.email = $1
	`
	row := r.db.QueryRow(ctx, query, email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userDomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username userDomain.Username) (*userDomain.User, error) {
	query := `
	SELECT 
		u.id,
		u.username,
		u.email,
		u.first_name,
		u.last_name,
		u.avatar_url,
		u.status,
		u.last_seen_at,
		u.created_at,
		u.updated_at,
		u.deleted_at,

		p.bio,
		p.location,
		p.website,
		p.birthday,
		p.created_at,
		p.updated_at,

		s.is_public,
		s.allow_notifications,
		s.theme,
		s.created_at,
		s.updated_at 
		FROM users u
	LEFT JOIN user_profiles p ON u.id = p.user_id
	LEFT JOIN user_settings s ON u.id = s.user_id
	WHERE u.username = $1
	`
	row := r.db.QueryRow(ctx, query, username.String())
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userDomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *userDomain.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// =====================
	// USERS
	// =====================
	cmd, err := tx.Exec(ctx, `
		UPDATE users SET
			username = $1,
			email = $2,
			first_name = $3,
			last_name = $4,
			avatar_url = $5,
			status = $6,
			last_seen_at = $7,
			updated_at = $8,
			deleted_at = $9
		WHERE id = $10
	`,
		user.Username().String(),
		user.Email().String(),
		user.FirstName(),
		user.LastName(),
		user.AvatarURL(),
		user.Status(),
		user.LastSeenAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.ID(),
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return userDomain.ErrUserNotFound
	}

	// =====================
	// PROFILE
	// =====================
	p := user.Profile()

	_, err = tx.Exec(ctx, `
		UPDATE user_profiles SET
			bio = $1,
			location = $2,
			website = $3,
			birthday = $4,
			updated_at = $5
		WHERE user_id = $6
	`,
		p.Bio(),
		p.Location(),
		p.Website(),
		p.Birthday(),
		p.UpdatedAt(),
		user.ID(),
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	// =====================
	// SETTINGS
	// =====================
	s := user.Settings()

	_, err = tx.Exec(ctx, `
		UPDATE user_settings SET
			is_public = $1,
			allow_notifications = $2,
			theme = $3,
			updated_at = $4
		WHERE user_id = $5
	`,
		s.IsPublic(),
		s.AllowNotifications(),
		s.Theme(),
		s.UpdatedAt(),
		user.ID(),
	)
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, tx port.Tx, userID uuid.UUID, status userDomain.UserStatus) error {
	query := `
		UPDATE users SET
			status = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	err := tx.Exec(ctx, query, status, userID)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	cmd, err := r.db.Exec(ctx, `
		UPDATE users
		SET 
			status = 'inactive',
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return userDomain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username userDomain.Username) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE username = $1
		)
	`

	err := r.db.QueryRow(ctx, query, username.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check user exists by username: %w", err)
	}

	return exists, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email userDomain.Email) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE email = $1
		)
	`, email.String()).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("exists by email: %w", err)
	}

	return exists, nil
}

func (r *userRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE id = $1
		)
	`, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists by id: %w", err)
	}

	return exists, nil
}

func mapPGError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return fmt.Errorf("insert user: %w", err)
	}

	switch pgErr.Code {
	case "23505": // unique_violation
		switch pgErr.ConstraintName {
		case "users_username_key":
			return userDomain.ErrUsernameAlreadyExists
		case "users_email_key":
			return userDomain.ErrEmailAlreadyExists
		}
	}

	return fmt.Errorf("insert user: %w", err)
}

func (r *userRepository) List(ctx context.Context, f userDomain.UserFilter) ([]*userDomain.User, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			u.email,
			u.first_name,
			u.last_name,
			u.avatar_url,
			u.status,
			u.last_seen_at,
			u.created_at,
			u.updated_at,
			u.deleted_at,

			p.bio,
			p.location,
			p.website,
			p.birthday,
			p.created_at,
			p.updated_at,

			s.is_public,
			s.allow_notifications,
			s.theme,
			s.created_at,
			s.updated_at

		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN user_settings s ON s.user_id = u.id
	`

	var (
		args   []interface{}
		where  []string
		argPos = 1
	)

	// =====================
	// FILTERS
	// =====================

	if f.Status() != nil {
		where = append(where, fmt.Sprintf("u.status = $%d", argPos))
		args = append(args, *f.Status())
		argPos++
	}

	if f.Search() != nil && *f.Search() != "" {
		where = append(where, fmt.Sprintf("(u.username ILIKE $%d OR u.email ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+*f.Search()+"%")
		argPos++
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	// =====================
	// SORT
	// =====================

	sortBy := "u.created_at"
	switch f.SortBy() {
	case "username":
		sortBy = "u.username"
	case "created_at":
		sortBy = "u.created_at"
	}

	order := "DESC"
	if strings.ToLower(f.SortOrder()) == "asc" {
		order = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, order)

	// =====================
	// PAGINATION
	// =====================

	limit := 20
	if f.Limit() > 0 && f.Limit() <= 100 {
		limit = f.Limit()
	}

	offset := 0
	if f.Offset() > 0 {
		offset = f.Offset()
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	// =====================
	// QUERY
	// =====================

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*userDomain.User

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context, f userDomain.UserFilter) (int, error) {
	query := `SELECT COUNT(*) FROM users u`

	var (
		args   []interface{}
		where  []string
		argPos = 1
	)

	if f.Status() != nil {
		where = append(where, fmt.Sprintf("u.status = $%d", argPos))
		args = append(args, *f.Status())
		argPos++
	} else {
		where = append(where, "u.status = 'active'")
	}

	if f.Search() != nil && *f.Search() != "" {
		where = append(where, fmt.Sprintf("(u.username ILIKE $%d OR u.email ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+*f.Search()+"%")
		argPos++
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return count, nil
}

func scanUser(row pgx.Row) (*userDomain.User, error) {
	var (
		// user
		id         uuid.UUID
		username   string
		email      string
		firstName  string
		lastName   string
		avatarURL  *string
		status     string
		lastSeenAt *time.Time
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  *time.Time

		// profile
		bio      *string
		location *string
		website  *string
		birthday *time.Time
		pCreated *time.Time
		pUpdated *time.Time

		// settings
		isPublic           *bool
		allowNotifications *bool
		theme              *string
		sCreated           *time.Time
		sUpdated           *time.Time
	)

	err := row.Scan(
		&id,
		&username,
		&email,
		&firstName,
		&lastName,
		&avatarURL,
		&status,
		&lastSeenAt,
		&createdAt,
		&updatedAt,
		&deletedAt,

		&bio,
		&location,
		&website,
		&birthday,
		&pCreated,
		&pUpdated,

		&isPublic,
		&allowNotifications,
		&theme,
		&sCreated,
		&sUpdated,
	)
	if err != nil {
		return nil, err
	}

	// ===== value objects
	uName, err := userDomain.NewUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid username in db: %w", err)
	}

	uEmail, err := userDomain.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email in db: %w", err)
	}

	// ===== profile
	profile := userDomain.RestoreProfile(
		bio,
		location,
		website,
		birthday,
		derefTime(pCreated),
		derefTime(pUpdated),
	)

	// ===== settings
	settings := userDomain.RestoreSettings(
		derefBool(isPublic),
		derefBool(allowNotifications),
		derefString(theme),
		derefTime(sCreated),
		derefTime(sUpdated),
	)

	// ===== aggregate

	user := userDomain.RestoreUser(
		id,
		uName,
		uEmail,
		firstName,
		lastName,
		avatarURL,
		userDomain.UserStatus(status),
		profile,
		settings,
		lastSeenAt,
		createdAt,
		updatedAt,
		deletedAt,
	)

	return user, nil
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func derefTime(v *time.Time) time.Time {
	if v == nil {
		return time.Time{}
	}
	return *v
}
