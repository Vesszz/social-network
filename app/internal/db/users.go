package db

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
	
}

func (s *Storage) DeleteUser(u *models.User) error {
	
}
