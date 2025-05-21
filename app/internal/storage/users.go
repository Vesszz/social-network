package storage

import (
	"database/sql"
	"social-network/internal/models"
	"time"
)

func (s *Storage) GetUsers() ([]models.User, error) {
	const query = `
		SELECT id, username, password_hash, created_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		var createdAt time.Time
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}
		
		user.CreatedAt = createdAt
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Storage) CreateUser(u *models.User) (int, error) {
	var id int
	query := `
		INSERT INTO users (username, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		u.Username,
		u.PasswordHash,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Storage) GetUserByName(name string) (models.User, error) {
	var user User
	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1
	`
	err := s.db.QueryRow(query, name).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Storage) DeleteUser(u *User) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`
	result, err := s.db.Exec(query, u.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
