package Postgres

import (
	"TP2023_DBMS_Project/internal/app/forum/Model"
)

func (p *RepStruct) CreateVote(vote Model.VotePost) error {
	query := `INSERT INTO vote(
				nickname,  
				voice,     
				idThread)
				VALUES ($1, $2, NULLIF($3, 0))`

	_, err := p.Connection.Exec(query, vote.Nickname, vote.Voice, vote.IdThread)
	return err
}

func (p *RepStruct) UpdateVote(vote Model.VotePost) error {
	query := `UPDATE vote SET voice=$1 WHERE LOWER(nickname) = LOWER($2) AND idThread = $3`
	_, err := p.Connection.Exec(query, vote.Voice, vote.Nickname, vote.IdThread)
	return err
}
