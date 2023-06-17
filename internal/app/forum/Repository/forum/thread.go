package Postgres

import (
	f "TP2023_DBMS_Project/internal/app/forum/Model"
	"fmt"
	"github.com/go-openapi/strfmt"
	"time"
)

func (p *RepStruct) CreateThread(thread f.ThreadForum) (f.ThreadForum, error) {
	query := `INSERT INTO thread(
    slug,
    author,
    created,
    message,
    title,
	forum)
	VALUES (NULLIF($1, ''), $2, $3, $4, $5, $6) RETURNING *`

	forum, err := p.ForumBySlug(thread.Forum)
	if err != nil {
		return f.ThreadForum{}, err
	}

	var threadAns f.ThreadForum
	var created time.Time

	if thread.Created != "" {
		err = p.Connection.QueryRow(query, thread.Slug, thread.Author,
			thread.Created, thread.Message, thread.Title, forum.Slug).Scan(&threadAns.Author,
			&created, &threadAns.Forum, &threadAns.Id, &threadAns.Message, &threadAns.Slug,
			&threadAns.Title, &threadAns.Votes)

	} else {
		err = p.Connection.QueryRow(query, thread.Slug, thread.Author,
			time.Time{}, thread.Message, thread.Title, forum.Slug).Scan(&threadAns.Author,
			&created, &threadAns.Forum, &threadAns.Id, &threadAns.Message, &threadAns.Slug,
			&threadAns.Title, &threadAns.Votes)
	}
	threadAns.Created = strfmt.DateTime(created.UTC()).String()
	return threadAns, err
}

func (p *RepStruct) GetThreadsWithParams(slug string, lim int, from string, desc bool) ([]f.ThreadForum, error) {
	var whereExp string
	var orderExp string

	if from != "" && desc {
		whereExp = fmt.Sprintf(`LOWER(forum)=LOWER('%s') AND created <= '%s'`, slug, from)
	} else if from != "" && !desc {
		whereExp = fmt.Sprintf(`LOWER(forum)=LOWER('%s') AND created >= '%s'`, slug, from)
	} else {
		whereExp = fmt.Sprintf(`LOWER(forum)=LOWER('%s')`, slug)
	}
	if desc {
		orderExp = `DESC`
	} else {
		orderExp = `ASC`
	}

	query := fmt.Sprintf("SELECT * FROM thread WHERE %s ORDER BY created %s LIMIT NULLIF(%d, 0)",
		whereExp, orderExp, lim)

	data := make([]f.ThreadForum, 0, 0)
	row, err := p.Connection.Query(query)

	if err != nil {
		return nil, err
	}

	defer func() {
		if row != nil {
			row.Close()
		}
	}()

	for row.Next() {
		var threadAns f.ThreadForum
		var created time.Time

		err = row.Scan(&threadAns.Author, &created, &threadAns.Forum, &threadAns.Id, &threadAns.Message, &threadAns.Slug, &threadAns.Title, &threadAns.Votes)

		if err != nil {
			return nil, err
		}

		threadAns.Created = strfmt.DateTime(created.UTC()).String()

		data = append(data, threadAns)
	}

	return data, err
}

func (p *RepStruct) IsThread(slug string) (bool, error) {
	query := `select exists(select 1 from thread where LOWER(forum)=LOWER($1))`

	var exists bool

	err := p.Connection.QueryRow(query, slug).Scan(&exists)
	return exists, err
}

func (p *RepStruct) ThreadBySlug(slug string) (f.ThreadForum, error) {
	query := `SELECT * FROM thread WHERE LOWER(slug)=LOWER($1)`

	var threadAns f.ThreadForum
	var created time.Time

	err := p.Connection.QueryRow(query, slug).Scan(&threadAns.Author, &created, &threadAns.Forum,
		&threadAns.Id, &threadAns.Message, &threadAns.Slug, &threadAns.Title, &threadAns.Votes)

	threadAns.Created = strfmt.DateTime(created.UTC()).String()
	return threadAns, err
}

func (p *RepStruct) ThreadByID(id int) (f.ThreadForum, error) {
	query := `SELECT * FROM thread WHERE id=$1`

	var threadAns f.ThreadForum
	var created time.Time

	err := p.Connection.QueryRow(query, id).Scan(&threadAns.Author, &created, &threadAns.Forum,
		&threadAns.Id, &threadAns.Message, &threadAns.Slug, &threadAns.Title, &threadAns.Votes)
	threadAns.Created = strfmt.DateTime(created.UTC()).String()

	return threadAns, err
}

func (p *RepStruct) ThreadIDBySlug(slug string) (int, error) {
	query := `SELECT id FROM thread WHERE LOWER(slug)=LOWER($1)`

	var id int
	err := p.Connection.QueryRow(query, slug).Scan(&id)
	return id, err
}

func (p *RepStruct) ThreadSlugByID(id int) (string, error) {
	query := `SELECT slug FROM thread WHERE id=$1`

	var slug string
	err := p.Connection.QueryRow(query, id).Scan(&slug)
	return slug, err
}

func (p *RepStruct) UpdateThread(updated f.ThreadForum) (f.ThreadForum, error) {
	query := `UPDATE thread SET message=COALESCE(NULLIF($1, ''), message), title=COALESCE(NULLIF($2, ''), title) WHERE `

	if updated.Id > 0 {
		query += `id = $3 RETURNING *`
		var threadAns f.ThreadForum
		var created time.Time
		err := p.Connection.QueryRow(query, updated.Message, updated.Title, updated.Id).Scan(
			&threadAns.Author, &created, &threadAns.Forum, &threadAns.Id, &threadAns.Message, &threadAns.Slug,
			&threadAns.Title, &threadAns.Votes)
		threadAns.Created = strfmt.DateTime(created.UTC()).String()
		return threadAns, err
	} else {
		query += `LOWER(slug) = LOWER($3) RETURNING *`
		var threadAns f.ThreadForum
		var created time.Time
		err := p.Connection.QueryRow(query, updated.Message, updated.Title, updated.Slug).Scan(
			&threadAns.Author, &created, &threadAns.Forum, &threadAns.Id, &threadAns.Message, &threadAns.Slug,
			&threadAns.Title, &threadAns.Votes)
		threadAns.Created = strfmt.DateTime(created.UTC()).String()
		return threadAns, err
	}
}
