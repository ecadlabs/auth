package users

import (
	"context"
	"database/sql"
	"fmt"
	"git.ecadlabs.com/ecad/auth/jsonpatch"
	"git.ecadlabs.com/ecad/auth/query"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	//log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type userModel struct {
	ID            uuid.UUID      `db:"id"`
	Email         string         `db:"email"`
	PasswordHash  []byte         `db:"password_hash"`
	Name          sql.NullString `db:"name"`
	Added         time.Time      `db:"added"`
	Modified      time.Time      `db:"modified"`
	EmailVerified bool           `db:"email_verified"`
	SortedBy      string         `db:"sorted_by"` // Output only
	Roles         pq.StringArray `db:"roles"`
}

func (u *userModel) toUser() *User {
	ret := &User{
		ID:            u.ID,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		Name:          u.Name.String,
		Added:         u.Added,
		Modified:      u.Modified,
		Roles:         make(map[string]interface{}),
		EmailVerified: u.EmailVerified,
	}

	for _, r := range u.Roles {
		ret.Roles[r] = true
	}

	return ret
}

type Storage struct {
	DB *sqlx.DB
}

func (s *Storage) getUser(ctx context.Context, col string, val interface{}) (*User, error) {
	var u userModel

	q := "SELECT users.*, ra.roles FROM users LEFT JOIN (SELECT user_id, array_agg(role) AS roles FROM roles GROUP BY user_id) AS ra ON ra.user_id = users.id WHERE users." + pq.QuoteIdentifier(col) + " = $1"
	if err := s.DB.GetContext(ctx, &u, q, val); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

func (s *Storage) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.getUser(ctx, "id", id)
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.getUser(ctx, "email", email)
}

var queryColumns = map[string]struct{}{
	"id":             struct{}{},
	"email":          struct{}{},
	"name":           struct{}{},
	"added":          struct{}{},
	"modified":       struct{}{},
	"roles":          struct{}{},
	"email_verified": struct{}{},
}

func (s *Storage) GetUsers(ctx context.Context, q *query.Query) ([]*User, *query.Query, error) {
	if q.SortBy == "" {
		q.SortBy = DefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "users.*, ra.roles, users." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   "users LEFT JOIN (SELECT user_id, array_agg(role) AS roles FROM roles GROUP BY user_id) AS ra ON ra.user_id = users.id",
		IDColumn:   "id",
		ValidateColumn: func(col string) bool {
			_, ok := queryColumns[col]
			return ok
		},
	}

	stmt, args, err := q.SelectStmt(&selOpt)
	if err != nil {
		return nil, nil, &Error{err, http.StatusBadRequest}
	}
	//log.Println(stmt)

	rows, err := s.DB.QueryxContext(ctx, stmt, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	users := []*User{}
	var lastItem *userModel

	for rows.Next() {
		var user userModel
		if err := rows.StructScan(&user); err != nil {
			return nil, nil, err
		}

		lastItem = &user
		users = append(users, user.toUser())
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	ret := *q

	if lastItem != nil {
		// Update query
		ret.LastID = lastItem.ID.String()
		ret.Last = lastItem.SortedBy
	}

	return users, &ret, nil
}

func isUniqueViolation(err error, constraint string) bool {
	e, ok := err.(*pq.Error)
	return ok && e.Code.Name() == "unique_violation" && e.Constraint == constraint
}

func (s *Storage) NewUser(ctx context.Context, user *User) (res *User, err error) {
	// TODO Check allowed roles
	model := userModel{
		ID:           uuid.NewV4(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         sql.NullString{String: user.Name, Valid: user.Name != ""},
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	// Create user
	rows, err := sqlx.NamedQueryContext(ctx, tx, "INSERT INTO users (id, email, password_hash, name) VALUES (:id, :email, :password_hash, :name) RETURNING added, modified, email_verified", &model)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			err = ErrEmail
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.StructScan(&model); err != nil {
			return
		}
	}

	if err = rows.Err(); err != nil {
		return
	}

	res = model.toUser()

	if len(user.Roles) == 0 {
		return
	}

	res.Roles = user.Roles

	// Create roles
	valuesExprs := make([]string, len(user.Roles))
	args := make([]interface{}, len(user.Roles)+1)

	args[0] = model.ID
	var i int

	for r := range user.Roles {
		valuesExprs[i] = fmt.Sprintf("($1, $%d)", i+2)
		args[i+1] = r
		i++
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO roles (user_id, role) VALUES "+strings.Join(valuesExprs, ", "), args...)
	return
}

func errPatchOp(o *jsonpatch.Op) error {
	return &Error{fmt.Errorf("Incorrect JSON patch op `%s' for path `%s'", o.Op, o.Path), http.StatusBadRequest}
}

var updatePaths = map[string]struct{}{
	"/email": struct{}{},
	"/name":  struct{}{},
}

func (s *Storage) PatchUser(ctx context.Context, id uuid.UUID, patch jsonpatch.Patch) (user *User, err error) {
	updateCols := make([]string, 0, len(patch))
	updateArgs := make([]interface{}, 0, len(patch)+1)

	addRolesArgs := make([]interface{}, 0, len(patch)+1)
	removeRolesArgs := make([]interface{}, 0, len(patch)+1)

	addRolesArgs = append(addRolesArgs, id)
	removeRolesArgs = append(removeRolesArgs, id)

	// Verify everything
	for _, o := range patch {
		if _, ok := updatePaths[o.Path]; ok {
			if o.Op != "replace" {
				return nil, errPatchOp(o)
			}

			if o.Value == nil {
				return nil, ErrPatchValue
			}

			updateCols = append(updateCols, o.Path[1:])
			updateArgs = append(updateArgs, o.Value)
		} else if o.Path == "/roles" {
			switch o.Op {
			case "add":
				if o.Value == nil {
					return nil, ErrPatchValue
				}
				addRolesArgs = append(addRolesArgs, o.Value)
			case "remove":
				removeRolesArgs = append(removeRolesArgs, o.Value)
			default:
				return nil, errPatchOp(o)
			}
		}
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	updateExpr := "UPDATE users SET "
	for i, c := range updateCols {
		if i != 0 {
			updateExpr += ", "
		}
		updateExpr += fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(c), i+1)
	}

	if len(updateCols) != 0 {
		updateExpr += ", "
	}
	updateExpr += fmt.Sprintf("modified = DEFAULT WHERE id = $%d RETURNING *", len(updateCols)+1)
	updateArgs = append(updateArgs, id)

	//log.Println(updateExpr)

	var u userModel
	if err = tx.GetContext(ctx, &u, updateExpr, updateArgs...); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
		}
		return nil, err
	}

	// Update roles
	if len(addRolesArgs) > 1 {
		insertExpr := "INSERT INTO roles (user_id, role) VALUES "

		for i := 0; i < len(addRolesArgs)-1; i++ {
			if i != 0 {
				insertExpr += ", "
			}
			insertExpr += fmt.Sprintf("($1, $%d)", i+2)
		}

		//log.Println(insertExpr)

		if _, err = tx.ExecContext(ctx, insertExpr, addRolesArgs...); err != nil {
			if isUniqueViolation(err, "roles_pkey") {
				err = ErrRoleExists
			}
			return nil, err
		}
	}

	if len(removeRolesArgs) > 1 {
		deleteExpr := "DELETE FROM roles WHERE user_id = $1 AND ("

		for i := 0; i < len(removeRolesArgs)-1; i++ {
			if i != 0 {
				deleteExpr += " OR "
			}
			deleteExpr += fmt.Sprintf("role = $%d", i+2)
		}

		deleteExpr += ")"

		//log.Println(deleteExpr)
		//log.Println(removeRolesArgs)

		if _, err = tx.ExecContext(ctx, deleteExpr, removeRolesArgs...); err != nil {
			return nil, err
		}
	}

	// Get roles back
	if err = tx.GetContext(ctx, &u, "SELECT array_agg(role) AS roles FROM roles WHERE user_id = $1 GROUP BY user_id", id); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	return u.toUser(), nil
}

func (s *Storage) DeleteUser(ctx context.Context, id uuid.UUID) (err error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	res, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM roles WHERE user_id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if err := s.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
