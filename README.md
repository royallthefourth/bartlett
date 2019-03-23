# Bartlett

[![GoDoc](https://godoc.org/github.com/royallthefourth/bartlett?status.svg)](https://godoc.org/github.com/royallthefourth/bartlett)
[![Go Report Card](https://goreportcard.com/badge/github.com/royallthefourth/bartlett)](https://goreportcard.com/report/github.com/royallthefourth/bartlett)
[![Build Status](https://travis-ci.org/royallthefourth/bartlett.svg?branch=master)](https://travis-ci.org/royallthefourth/bartlett)
[![codecov](https://codecov.io/gh/royallthefourth/bartlett/branch/master/graph/badge.svg)](https://codecov.io/gh/royallthefourth/bartlett)

*Bartlett* is a library that automatically generates a CRUD API for your Go web application.

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
    b := bartlett.Bartlett{DB: db, Driver: &mariadb.MariaDB{}, Tables: tables, Users: dummyUserProvider}
    routes, handlers := b.Routes()
    for i, route := range routes {
    	http.HandleFunc(`/api` + route, handlers[i]) // Adds /api/students to the server.
    }
    
    log.Fatal(http.ListenAndServe(`:8080`, nil))
}
```

See the [todo list demo application](https://github.com/royallthefourth/bartlett-todo) for a bigger example.

### Querying

#### `SELECT`

To `SELECT` from a table, make a `GET` request to its corresponding URL.
For example, `SELECT * FROM students;` against the example above may be achieved by `curl -XGET http://localhost:8080/students`
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

Requests may filter columns by the `select=` query parameter, eg `/students?select=student_id,grade`

##### `WHERE`

To filter on simple `WHERE` conditions, specify a column name as a query string parameter and the conditions as the value.
For example: `/students?age=eq.20` produces `WHERE age = 20`.

| Operator  | SQL       | Note                      |
| --------- | --------- | ------------------------- |
|   `eq`    |   `=`     |                           |
|   `neq`   |   `!=`    |                           |
|   `gt`    |   `>`     |                           |
|   `gte`   |   `>=`    |                           |
|   `lt`    |   `<`     |                           |
|   `lte`   |   `<=`    |                           |
|   `like`  |   `LIKE`  | use `*` in place of `%`   |
|   `is`    |   `IS`    | eg `is.true` or `is.null` |
|   `in`    |   `IN`    | eg `in."hi, there","bye"` |

Any of these conditions can be negated by prefixing it with `not.` eg `/students?age=not.eq.20`

##### `ORDER BY`

To order results, add `order` to the query: `/students?order=student_id`

Order by mutliple columns by separating them with `,`: `/students?order=age,grade`

Choose `ASC` or `DESC` by appending `.asc` or `.desc` to the field name: `/students?order=age.asc,grade.desc`

##### `LIMIT` and `OFFSET`

To restrict result output, add `limit`. The request `/students?limit=10` will return 10 results.

To add an offset, use `offset` in your query: `/students?limit=10&offset=2` will return 10 after skipping the first 2 results.

#### `INSERT`

To write rows to a table, make a `POST` request to the corresponding table's URL.
Your request should include a payload in the form of a JSON array of rows to insert.

To generate your own surrogate key for each row, identify in your `Table` struct an `IDColumn`.
Provide a function that returns a new ID each time it's invoked.
This column will be protected from tampering by users. The `UserID` column is also filtered out incoming `POST` requests.

#### `UPDATE`

To run an `UPDATE` query, issue a `PATCH` request.
Set your `WHERE` params on the URL exactly the way you do with a `SELECT`.
Any `PATCH` requests that do not have a `WHERE` will be rejected for your safety.

`PATCH` requests must include a JSON payload body with the fields to be updated and their values:
```json
{
  "age": 71,
  "name": "Alex"
}
```

#### `DELETE`

To delete rows from a table, make a `DELETE` request to the corresponding table's URL.

You _must_ specify at least one `WHERE` clause, otherwise the request will return an error.
This is a design feature to prevent users from deleting everything by mistake.
 
## Status

This project is under heavy development.
Bartlett currently supports SQLite3 and MariaDB.
Postgres support is planned once support for existing databases is more robust.
Most data types are not yet under test and may not produce useful results.
Some MariaDB types do not have a clear JSON representation. These types are marshaled as `[]byte`.

## Security

Taking user input from the web to paste into a SQL query does present some hazards.
The only place where user input is placed into queries is by parameter placeholders.
All other dynamic SQL strings are generated from the strings passed into the arguments at startup time, never from the URL.

To restrict access per-row, specify a column name in your `Table.UserID`.
Tables with a `UserID` set will always filter according to `Table.UserID = ?` with the result of the userProvider function.

## Prior Art

This project is inspired by [Postgrest](https://www.postgrest.org/).
Instead of something that runs everything on its own, though, I prefer a tool that integrates with my existing application.
