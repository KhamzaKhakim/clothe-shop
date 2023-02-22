package data

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"time"
)

type Roles []string

func (p Roles) Include(role string) bool {
	for i := range p {
		if role == p[i] {
			return true
		}
	}
	return false
}

type RolesModel struct {
	DB *sql.DB
}

func (m RolesModel) GetAllRolesForUser(userID int64) (Roles, error) {
	query := `
SELECT roles.role
FROM roles
INNER JOIN users_roles ON users_roles.roles_id = roles.id
INNER JOIN users ON users_roles.user_id = users.id
WHERE users.id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles Roles
	for rows.Next() {
		var role string
		err := rows.Scan(&role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (m RolesModel) AddRolesForUser(userID int64, roles ...string) error {
	query := `
INSERT INTO users_roles
SELECT $1, roles.id FROM roles WHERE roles.role = ANY($2)`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(roles))
	return err
}
