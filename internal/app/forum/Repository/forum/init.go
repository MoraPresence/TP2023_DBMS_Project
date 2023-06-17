package Postgres

import (
	"TP2023_DBMS_Project/internal/app/forum/Repository"
	"TP2023_DBMS_Project/internal/app/user"
	"errors"
	"github.com/jackc/pgx"
)

type RepStruct struct {
	Connection *pgx.ConnPool
	Rep        user.Repository
}

func New(conn *pgx.ConnPool, repository user.Repository) Repository.Repository {
	return &RepStruct{
		Connection: conn,
		Rep:        repository,
	}
}

func (p *RepStruct) Status() (map[string]int, error) {
	query := `SELECT * FROM (SELECT COUNT(*) FROM forum) as fC, (SELECT COUNT(*) FROM post) as pC,
              (SELECT COUNT(*) FROM thread) as tC, (SELECT COUNT(*) FROM users) as uC;`

	a, err := p.Connection.Query(query)
	if err != nil {
		return nil, err
	}

	if a.Next() {
		forumCount, postCount, threadCount, usersCount := 0, 0, 0, 0
		err := a.Scan(&forumCount, &postCount, &threadCount, &usersCount)
		if err != nil {
			return nil, err
		}
		return map[string]int{
			"forum":  forumCount,
			"post":   postCount,
			"thread": threadCount,
			"user":   usersCount,
		}, nil
	}
	return nil, errors.New("no info available")
}

func (p *RepStruct) Clear() error {
	query := `TRUNCATE users, forum, thread, post, vote, users_forum;`

	_, err := p.Connection.Exec(query)
	return err
}
