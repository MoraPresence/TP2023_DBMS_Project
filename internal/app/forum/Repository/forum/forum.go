package Postgres

import (
	f "TP2023_DBMS_Project/internal/app/forum/Model"
)

func (p *RepStruct) CreateForum(forum f.ForumProj) (f.ForumProj, error) {
	query := `INSERT INTO forum(
    "user",
    slug,
    title)
	VALUES ($1, $2, $3) RETURNING *`
	user, err := p.Rep.GetUserByNick(forum.User)
	if err != nil {
		return f.ForumProj{}, err
	}
	var newF f.ForumProj
	err = p.Connection.QueryRow(query, user.Nick, forum.Slug, forum.Title).Scan(&newF.User, &newF.Post, &newF.Slug, &newF.Thread, &newF.Title)
	return newF, err
}

func (p *RepStruct) ForumBySlug(slug string) (f.ForumProj, error) {
	query := `SELECT * FROM forum WHERE LOWER(slug)=LOWER($1)`
	var forum f.ForumProj
	err := p.Connection.QueryRow(query, slug).Scan(&forum.User, &forum.Post, &forum.Slug, &forum.Thread, &forum.Title)
	return forum, err
}

func (p *RepStruct) getForumSlug(ID int) (string, error) {
	query := `SELECT forum FROM thread WHERE id=$1`

	var slug string
	err := p.Connection.QueryRow(query, ID).Scan(&slug)
	return slug, err
}
