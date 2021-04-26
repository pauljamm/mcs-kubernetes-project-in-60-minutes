package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mcs-kubernetes-project/models"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func createConnection() *sql.DB {
	dbport, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	os.Getenv("DB_HOST"), dbport, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected to the database!")

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		userid SERIAL PRIMARY KEY NOT NULL,
                name TEXT NOT NULL,
                location TEXT NOT NULL,
                age INT NOT NULL
	);`); err != nil {
		log.Fatal("Unable to create `users` table. $v", err)
	}

	return db
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	res := response{
		Message: "OK",
	}
	json.NewEncoder(w).Encode(res)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {

	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}

	insertID := insertUser(user)

	res := response{
		ID:      insertID,
		Message: "User created successfully",
	}

	json.NewEncoder(w).Encode(res)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	user, err := getUser(int64(id))

	if err != nil {
		log.Fatalf("Unable to get user. %v", err)
	}

	json.NewEncoder(w).Encode(user)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {

	users, err := getAllUsers()

	if err != nil {
		log.Fatalf("Unable to get all user. %v", err)
	}

	json.NewEncoder(w).Encode(users)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	var user models.User

	err = json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}

	updatedRows := updateUser(int64(id), user)

	msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", updatedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	deletedRows := deleteUser(int64(id))

	msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", deletedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func insertUser(user models.User) int64 {

	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO users (name, location, age) VALUES ($1, $2, $3) RETURNING userid`

	var id int64

	err := db.QueryRow(sqlStatement, user.Name, user.Location, user.Age).Scan(&id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

        log.Printf("Inserted a single record %v", id)

	return id
}

func getUser(id int64) (models.User, error) {
	db := createConnection()

	defer db.Close()

	var user models.User

	sqlStatement := `SELECT * FROM users WHERE userid=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&user.ID, &user.Name, &user.Age, &user.Location)

	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	return user, err
}

func getAllUsers() ([]models.User, error) {
	db := createConnection()

	defer db.Close()

	var users []models.User

	sqlStatement := `SELECT * FROM users`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var user models.User

		err = rows.Scan(&user.ID, &user.Name, &user.Location, &user.Age)

		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}

		users = append(users, user)

	}

	return users, err
}

func updateUser(id int64, user models.User) int64 {

	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE users SET name=$2, location=$3, age=$4 WHERE userid=$1`

	res, err := db.Exec(sqlStatement, id, user.Name, user.Location, user.Age)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	log.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

func deleteUser(id int64) int64 {

	db := createConnection()

	defer db.Close()

	sqlStatement := `DELETE FROM users WHERE userid=$1`

	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	log.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}
