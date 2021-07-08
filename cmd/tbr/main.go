package main

import (
	"database/sql"
	"fmt"
	"github.com/e-zk/subc"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type Book struct {
	name     string
	author   string
	year     int
	bookType string
	url      string
}

const (
	dbPath = "./test.db"
)

func setupDb(db *sql.DB) error {
	sqlStmt := `
	create table if not exists books(name text not null primary key, author text not null, year integer not null, type text default null, hyperlink type text default null);
	create table if not exists read(bookName text not null primary key, dateRead datetime not null,
	foreign key(bookName) references books(name));
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}

func insertBook(db *sql.DB, book Book) error {
	insertBookSQL := `replace into books(name, author, year) values (?,?,?)`
	stmt, err := db.Prepare(insertBookSQL)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(book.name, book.author, book.year)
	if err != nil {
		return err
	}

	return nil
}

func removeBook(db *sql.DB, bookTitle string) error {
	sql := `
	delete from books
	where books.name = ? ;
	`
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(bookTitle)
	if err != nil {
		return err
	}

	return nil
}

func bookRead(db *sql.DB, bookTitle string) error {
	sql := `insert into read(bookName, dateRead) values (?,?)`
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(bookTitle, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func getToBeRead(db *sql.DB) ([]Book, error) {
	var tbr []Book
	getTBRSQL := `
	select books.name, books.author, books.year
	from books
	where books.name not in (select read.bookName from read);
	`

	rows, err := db.Query(getTBRSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thisBook Book
		err := rows.Scan(&thisBook.name, &thisBook.author, &thisBook.year)
		if err != nil {
			return nil, err
		}
		tbr = append(tbr, thisBook)
	}

	return tbr, nil
}

func getRead(db *sql.DB) ([]Book, error) {
	var read []Book

	getReadSQL := `
	select books.name, books.author, books.year
	from read
	join books on read.bookName = books.name;
	`

	rows, err := db.Query(getReadSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thisBook Book
		err := rows.Scan(&thisBook.name, &thisBook.author, &thisBook.year)
		if err != nil {
			return nil, err
		}
		read = append(read, thisBook)
	}

	return read, nil
}

func (b Book) String() string {
	return fmt.Sprintf("%s, by %s (%d)", b.name, b.author, b.year)
}

func (b Book) HTML() string {
	var output string

	if b.url != "" {
		output = fmt.Sprintf("<a href=\"%s\"><em>%s</em></a>, by  %s (%d)", b.url, b.name, b.author, b.year)
	} else {
		output = fmt.Sprintf("<em>%s</em>, by %s (%d)", b.name, b.author, b.year)
	}

	if b.bookType != "" && b.bookType != "book" {
		output = fmt.Sprintf("%s [%s]", output, b.bookType)
	}

	return output
}

func main() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	err = setupDb(db)
	if err != nil {
		log.Fatal(err)
	}

	var (
		title  string
		author string
		year   int
	)

	subc.Sub("add").IntVar(&year, "y", 0, "year book was released")
	subc.Sub("add").StringVar(&title, "n", "default", "book title")
	subc.Sub("add").StringVar(&author, "a", "default", "book author")
	subc.Sub("remove").StringVar(&title, "n", "default", "title of book to remove")
	subc.Sub("read").StringVar(&title, "n", "default", "title of book to mark as read")

	subcommand, err := subc.Parse()
	if err == subc.ErrSubcNotExist {
		panic(err)
	} else if err == subc.ErrUsage {
		return
	} else if err != nil {
		panic(err)
	}

	switch subcommand {
	case "add":
		err = insertBook(db, Book{name: title, author: author, year: year})
	case "remove":
		err = removeBook(db, title)
	case "read":
		err = bookRead(db, title)
	}

	if err != nil {
		panic(err)
	}
}
