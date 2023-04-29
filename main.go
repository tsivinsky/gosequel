package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var pool *sql.DB

type User struct {
	FirstName string `json:"firstName" sql:"first_name"`
	LastName  string `json:"lastName" sql:"last_name"`
}

func findUsers() ([]User, error) {
	rows, err := pool.Query("select * from users;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		user := new(User)
		err = rows.Scan(&user.FirstName, &user.LastName)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return users, nil
}

func findByName(firstName string) (*User, error) {
	rows, err := pool.Query("select * from users where first_name = $1;", firstName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()

	var f, l string
	err = rows.Scan(&f, &l)
	if err != nil {
		return nil, err
	}

	user := &User{
		FirstName: f,
		LastName:  l,
	}

	return user, nil
}

func createUser(firstName, lastName string) error {
	_, err := pool.Exec("insert into users (first_name, last_name) values ($1, $2);", firstName, lastName)
	if err != nil {
		return err
	}

	return nil
}

func prompt(qs string) (string, error) {
	r := bufio.NewReader(os.Stdin)
	fmt.Print(qs)
	ans, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	ans = strings.TrimSpace(ans)

	return ans, nil
}

func main() {
	flag.Parse()

	cmd := flag.Arg(0)

	godotenv.Load()

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("DB_HOST")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)

	pool, _ = sql.Open("postgres", dsn)
	defer pool.Close()

	pool.SetConnMaxLifetime(0)
	pool.SetMaxIdleConns(3)
	pool.SetMaxOpenConns(3)

	switch cmd {
	case "create":
		firstName, _ := prompt("First name: ")
		lastName, _ := prompt("Last name: ")

		err := createUser(firstName, lastName)
		if err != nil {
			panic(err)
		}
		break

	case "find":
		firstName, _ := prompt("First name: ")

		user, err := findByName(firstName)
		if err != nil {
			fmt.Printf("No user found\n")
			os.Exit(1)
		}
		fmt.Printf("Hello, %s %s!\n", user.FirstName, user.LastName)
		break

	default:
		users, err := findUsers()
		if err != nil {
			panic(err)
		}

		for _, user := range users {
			fmt.Printf("Hello, %s %s!\n", user.FirstName, user.LastName)
		}
		break
	}
}
