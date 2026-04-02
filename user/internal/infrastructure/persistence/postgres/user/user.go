// Package userpg
package userpg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
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
			status, role,
			last_seen_at,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		user.ID(),
		user.Username().String(),
		user.Email().String(),
		user.FirstName(),
		user.LastName(),
		user.AvatarURL(),
		user.Status(),
		user.Role(),
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
		u.role,
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
	WHERE u.id = $1 AND u.status == 'active'
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
	// Implementation of the FindByEmail method
	return nil, nil
}

func (r *userRepository) Update(ctx context.Context, user *userDomain.User) error {
	// Implementation of the Update method
	return nil
}

func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	// Implementation of the Delete method
	return nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username userDomain.Username) (bool, error) {
	// Implementation of the ExistsByUsername method
	return false, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email userDomain.Email) (bool, error) {
	// Implementation of the ExistsByEmail method
	return false, nil
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
		role       string
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
		&role,
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
		userDomain.UserRole(role),
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
