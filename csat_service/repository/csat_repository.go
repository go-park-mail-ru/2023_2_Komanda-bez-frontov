package repository

import (
//  "database/sql"
 "go-form-hub/csat_service/db"
)

// PassageService представляет сервис для работы с таблицей csat_passage
type CSATService struct {
 DB *db.Database
}

type Raiting struct {
	Raiting int `json:"rating"`
}

type Results struct {
	Results []Raiting `json:"ratings"`
}

// NewPassageService возвращает новый экземпляр PassageService
func NewCSATService(db *db.Database) *CSATService {
 	return &CSATService{DB: db}
}

// CheckPassageByID проверяет наличие строки в таблице csat_passage по user_id
func (s *CSATService) CheckPassageByUserID(userID int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM csat_passage WHERE user_id = $1)"
	err := s.DB.Conn.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// AddPassageAndForm добавляет строки в таблицы csat_form и csat_passage с данными userID и rating
func (s *CSATService) AddPassage(userID int64, rating int) error {
	query := "INSERT INTO csat_passage (user_id, rating) VALUES ($1, $2)"
	// err := s.DB.Conn.QueryRow(query, userID, rating).Scan(&result)
	// fmt.Println(err, result)
	// if err != nil {
	// 	return err
	// }
	// return nil
	_, err := s.DB.Conn.Exec(query, userID, rating)
    if err != nil {
        return err
    }

    return nil
}

func (s *CSATService) Results() (Results, error) {
	query := "SELECT rating FROM csat_passage"
    rows, err := s.DB.Conn.Query(query)
    if err != nil {
        return Results{}, err
    }
    defer rows.Close()

    var raitings []Raiting
    for rows.Next() {
        var raiting Raiting
        if err := rows.Scan(&raiting.Raiting); err != nil {
            return Results{}, err
        }
        raitings = append(raitings, raiting)
    }
    if err := rows.Err(); err != nil {
        return Results{}, err
    }

    results := Results{Results: raitings}
    return results, nil
}