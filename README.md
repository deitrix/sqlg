# SQLg

Add sqlg as a dependency to your project
```sh
go get github.com/deitrix/sqlg
```

Then import it
```go
import "github.com/deitrix/sqlg"
```

## Usage

### Selecting multiple rows

First, define your scan function. This should be in the signature `func(row sqlg.Row) (T, error)`
```go
func scanUser(row sqlg.Row) (u User, err error) {
	return u, row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}
```

Then, build your goqu query and call `sqlg.SelectAll`, ensuring to pass a `sqlg.Queryable` (implemented by `*sql.DB`)
```go
import _ "github.com/doug-martin/goqu/v9/dialect/mysql"

var mysql = goqu.Dialect("mysql")

type Store struct {
	db sqlg.Queryable
}

var usersTable = goqu.T("users")

var selectUsers = mysql.
	Select("id", "name", "email", "password", "created_at", "updated_at").
	From(usersTable)

func (s *Store) UsersCreatedSince(ctx context.Context, since time.Time) ([]User, error) {
	return sqlg.SelectAll(ctx, s.db, scanUser, selectUsers.
		Where(goqu.C("created_at").Gt(since)))
}
```

### Selecting a single row
```go
func (s *Store) UserByID(ctx context.Context, id int) (User, error) {
	return sqlg.Select(ctx, s.db, scanUser, selectUsers.
		Where(goqu.C("id").Eq(id)).
		Limit(1))
}
```

### Inserting, updating or deleting rows
```go
func (s *Store) CreateUser(ctx context.Context, user User) (int, error) {
	return sqlg.ExecID(ctx, s.db, mysql.
		Insert(usersTable).
		Rows(goqu.Record{
			"name":       user.Name,
			"email":      user.Email,
			"password":   user.Password,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}))
}

func (s *Store) UpdateUser(ctx context.Context, id int, user User) error {
	return sqlg.Exec(ctx, s.db, mysql.
		Update(usersTable).
		Set(goqu.Record{
			"name":       user.Name,
			"email":      user.Email,
			"password":   user.Password,
			"updated_at": time.Now(),
		}).
		Where(goqu.C("id").Eq(id)))
}

func (s *Store) DeleteUser(ctx context.Context, id int) error {
	return sqlg.Exec(ctx, s.db, mysql.
		Delete(usersTable).
		Where(goqu.C("id").Eq(id)))
}
```
