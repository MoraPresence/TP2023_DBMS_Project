package repository

import (
	"TP2023_DBMS_Project/internal/app/user"
	u "TP2023_DBMS_Project/internal/app/user/models"
	"fmt"
	"github.com/jackc/pgx"
)

type postgresUserRepository struct {
	Conn *pgx.ConnPool
}

func New(conn *pgx.ConnPool) user.Repository {
	return &postgresUserRepository{
		Conn: conn,
	}
}

func (p *postgresUserRepository) AddUser(user u.UserInfo) error {
	query := `INSERT INTO users(
    about,
    email,
    fullname,
    nickname)
	VALUES ($1, $2, $3, $4)`

	_, err := p.Conn.Exec(query, user.About, user.Email, user.Full, user.Nick)
	return err
}

func (p *postgresUserRepository) GetUserByNickAndEmail(nickname, email string) ([]u.UserInfo, error) {
	query := `SELECT * FROM users WHERE LOWER(Nickname)=LOWER($1) OR Email=$2`

	var data []u.UserInfo
	row, err := p.Conn.Query(query, nickname, email)
	if err != nil {
		return nil, err
	}
	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		var use u.UserInfo
		err = row.Scan(&use.About, &use.Email, &use.Full, &use.Nick)
		if err != nil {
			return nil, err
		}
		data = append(data, use)
	}

	return data, err
}

func (p *postgresUserRepository) GetUserByNick(nickname string) (u.UserInfo, error) {
	query := `SELECT * FROM users WHERE LOWER(Nickname)=LOWER($1)`

	var userObj u.UserInfo
	err := p.Conn.QueryRow(query, nickname).Scan(&userObj.About, &userObj.Email, &userObj.Full, &userObj.Nick)
	return userObj, err
}

func (p *postgresUserRepository) UpdateUser(user u.UserInfo) (u.UserInfo, error) {
	query := `UPDATE users SET 
                 about=COALESCE(NULLIF($1, ''), about),
                 email=COALESCE(NULLIF($2, ''), email),
                 fullname=COALESCE(NULLIF($3, ''), fullname) 
	WHERE LOWER(nickname) = LOWER($4) RETURNING *`

	var userObj u.UserInfo
	err := p.Conn.QueryRow(query, user.About, user.Email, user.Full, user.Nick).Scan(&userObj.About, &userObj.Email, &userObj.Full, &userObj.Nick)
	return userObj, err
}

func (p *postgresUserRepository) GetUsersByForumSlug(slug string, lim int, from string, desc bool) ([]u.UserInfo, error) {
	var query string
	if desc {
		if from != "" {
			query = fmt.Sprintf(`SELECT users.about, users.Email, users.FullName, users.Nickname FROM users
    	inner join users_forum uf on users.Nickname = uf.nickname
        WHERE uf.slug =$1 AND uf.nickname < '%s'
        ORDER BY lower(users.Nickname) DESC LIMIT NULLIF($2, 0)`, from)
		} else {
			query = `SELECT users.about, users.Email, users.FullName, users.Nickname FROM users
    	inner join users_forum uf on users.Nickname = uf.nickname
        WHERE uf.slug =$1
        ORDER BY lower(users.Nickname) DESC LIMIT NULLIF($2, 0)`
		}
	} else {
		query = fmt.Sprintf(`SELECT users.about, users.Email, users.FullName, users.Nickname FROM users
    	inner join users_forum uf on users.Nickname = uf.nickname
        WHERE uf.slug =$1 AND uf.nickname > '%s'
        ORDER BY lower(users.Nickname) LIMIT NULLIF($2, 0)`, from)
	}
	var data []u.UserInfo
	row, err := p.Conn.Query(query, slug, lim)

	if err != nil {
		return data, nil
	}

	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		var use u.UserInfo
		err = row.Scan(&use.About, &use.Email, &use.Full, &use.Nick)
		if err != nil {
			return data, err
		}
		data = append(data, use)
	}
	return data, err
}
