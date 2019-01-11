# Bartlett

[![Go Report Card](https://goreportcard.com/badge/github.com/royallthefourth/bartlett)](https://goreportcard.com/report/github.com/royallthefourth/bartlett)
[![Build Status](https://travis-ci.org/royallthefourth/bartlett.svg?branch=master)](https://travis-ci.org/royallthefourth/bartlett)
[![codecov](https://codecov.io/gh/royallthefourth/bartlett/branch/master/graph/badge.svg)](https://codecov.io/gh/royallthefourth/bartlett)

*Bartlett* is a library that automatically adds API routes corresponding to your database tables to your Go web application.

## Usage

Invoke Bartlett by providing a database connection, a Bartlett driver, a slice of tables, and a function that returns a userID.
Bartlett will return a slice of routes corresponding to your table names and a slice of HTTP request handlers.

### Server Setup

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    _ "github.com/go-sql-driver/mysql"
    "github.com/royallthefourth/bartlett"
    "github.com/royallthefourth/bartlett/mariadb"
)

func indexPage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, `Welcome to your Bartlett application! The interesting parts are mounted under /api`)
}

func dummyUserProvider(_ *http.Request) (interface{}, error) {
    return 0, nil // In a real application, use a closure that includes your session handler to generate a user ID. 
}

func main() {
    http.HandleFunc(`/`, indexPage)
    
    // The students table will be available from the API, but the rest of the database will not.
    tables := []bartlett.Table{
    	{
            Name: `students`,
            UserID: `student_id`, // Requests will only return rows corresponding to their ID for this table.
    	},
    }
    db, err := sql.Open("mysql", ":@/school")
    if err != nil {
        log.Fatal(err)
    }
    
    // Bartlett is not a web application.
    // Instead, it is a tool that allows you to quickly add an API to your existing application.
    routes, handlers := bartlett.Bartlett{db, mariadb.MariaDB{}, tables, dummyUserProvider}.Routes()
    for i, route := range routes {
    	http.HandleFunc(`/api` + route, handlers[i]) // Adds /api/students to the server.
    }
    
    log.Fatal(http.ListenAndServe(`:8080`, nil))
}
```

### Querying

#### SELECT

To `SELECT` from a table, make a `GET` request to its corresponding URL.
For example, `SELECT * FROM students;` may be achieved by `curl -XGET http://localhost:8080/students`
The result set will be emitted as a JSON array:
```json
[
    {
    "student_id": 1,
    "age": 18,
    "grade": 85
    },
    {
      "student_id": 2,
      "age": 20,
      "grade": 91
    }
]
```
Note that all results are emitted as an array, even if there is only one row.

Requests may filter columns by the `select=` query parameter to cut out irrelevant data.
Separate column names by `,`: `/students?select=student_id,grade`


## Status

This project is under heavy development.
Bartlett currently supports SQLite3 and MariaDB.
Postgres support is planned once support for existing databases is more robust.
Most data types are not yet under test and may not produce good results.
Some MariaDB types do not have a clear JSON representation. These types are marshaled as `[]byte`.

`WHERE` clauses, `INSERT`, `UPDATE`, `DELETE`, and `JOIN` are all planned for future development.

## Security

Taking user input from the web to paste into a SQL query does prevent some hazards.
The only place where user input is placed into queries is by parameter placeholders.
All other dynamic SQL strings generated from the strings passed into the arguments at startup time, never from the URL.

To restrict access per-row, specify a column name in your `Table.UserID`.
Tables with a `UserID` set will always filter according to `Table.UserID = ?` with the result of the userProvider function.

## Prior Art

This project is inspired by [Postgrest](https://www.postgrest.org/).
Instead of something that runs everything on its own, though, I prefer a tool that integrates with my existing application.
