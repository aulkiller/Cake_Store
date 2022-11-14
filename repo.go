package main

import (
	"database/sql"
	"log"
	"time"
)

type Cake struct {
	Id            int       `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Rating        float64   `json:"rating"`
	Image         string    `json:"image"`
	RawCreated_at time.Time `json:"-"`
	RawUpdated_at time.Time `json:"-"`
	Created_at    string    `json:"created_at"`
	Updated_at    string    `json:"updated_at"`
}

type CakeList struct {
	Id     int     `json:"id"`
	Title  string  `json:"title"`
	Rating float64 `json:"rating"`
	Image  string  `json:"image"`
}

func (c *Cake) getCake(db *sql.DB) error {
	query := `
	SELECT 
		id, title, description, rating, image, created_at, updated_at
	FROM 
		cakes
	WHERE
		id = ?
	`

	return db.QueryRow(query, c.Id).Scan(&c.Id, &c.Title, &c.Description, &c.Rating, &c.Image, &c.RawCreated_at, &c.RawUpdated_at)
}

func (c *Cake) updateCake(db *sql.DB) error {
	query := `
	Update
		cakes
	SET
		title = ?,
		description = ?,
		rating = ?,
		image = ?,
		updated_at = NOW()
	WHERE
		id = ?
	`

	_, err := db.Exec(query, c.Title, c.Description, c.Rating, c.Image, c.Id)
	return err
}

func (c *Cake) deleteCake(db *sql.DB) error {
	query := `
	DELETE FROM
		cakes
	WHERE
		id = ?
	`

	_, err := db.Exec(query, c.Id)
	return err
}

func (c *Cake) createCake(db *sql.DB) error {
	query := `
	INSERT INTO
		cakes(title, description,rating,image,created_at,updated_at)
	VALUES
		(?, ?, ?, ? , NOW(), NOW())
	`

	_, err := db.Exec(query, c.Title, c.Description, c.Rating, c.Image)
	if err != nil {
		log.Println("error db: ", err)
		return err
	}

	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&c.Id)
	if err != nil {
		log.Println("error db: ", err)
		return err
	}

	return nil
}

func getCakes(db *sql.DB) ([]CakeList, error) {
	cakes := []CakeList{}
	query := `
	SELECT 
		id, title, rating, image
	FROM 
		cakes
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("error db: ", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cake CakeList

		err = rows.Scan(&cake.Id, &cake.Title, &cake.Rating, &cake.Image)
		if err != nil {
			log.Println("error db: ", err)
			return nil, err
		}

		cakes = append(cakes, cake)
	}

	return cakes, nil
}
