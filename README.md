# Bartlett

[![Go Report Card](https://goreportcard.com/badge/github.com/royallthefourth/bartlett)](https://goreportcard.com/report/github.com/royallthefourth/bartlett)
[![Build Status](https://travis-ci.org/royallthefourth/bartlett.svg?branch=master)](https://travis-ci.org/royallthefourth/bartlett)

*Bartlett* is a library that automatically adds API routes corresponding to your database tables to your Go web application.

## Usage

Invoke it by providing a function that returns a userID, a slice of tables, and a database connection.

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
    return 0, nil
}

func main() {
    http.HandleFunc(`/`, indexPage)
    
    // The students table will be available from the API, but the rest of the database will not.
    tables := []bartlett.Table{
    	{
            Name: `students`,
    	},
    }
    db, err := sql.Open("mysql", ":@/school")
    if err != nil {
        log.Fatal(err)
    }
    
    // Bartlett is not a web application.
    // Instead, it is a tool that allows you to quickly add an API to your existing application.
    routes, handlers := mariadb.New(db, tables, dummyUserProvider).Routes()
    for i, route := range routes {
    	http.HandleFunc(`/api` + route, handlers[i]) // Adds /api/students to the server.
    }
    
    log.Fatal(http.ListenAndServe(`:8080`, nil))
}
```

## Status

Bartlett currently supports SQLite3 and MariaDB.
Most data types are not yet under test and may not produce good results.
Some MariaDB types do not have a clear JSON representation. These types are marshaled as `[]byte`.
The current behavior consists of returning a JSON body corresponding to the query `SELECT * FROM $TABLE;`.
`WHERE` clauses, column selection, authentication, and joins are all planned for future development.

## Security

Taking user input from the web to paste into a SQL query does prevent some hazards.
I mitigate the risk of this by using a whitelist for each type of input and parameter queries for everything else.
The only tables that can be queried with Bartlett are the ones you specify.

## Prior Art

This project is inspired by [Postgrest](https://www.postgrest.org/).
Instead of something that runs everything on its own, though, I prefer a tool that integrates with my existing application.
