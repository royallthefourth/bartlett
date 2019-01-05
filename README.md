# Bartlett

*Bartlett* is a library that automatically adds API routes corresponding to your database tables to your Go web application.

## Usage

Invoke it by providing an http server, a function that returns a userID, a list of tables, and a database connection.

## Security

Taking user input from the web to paste into a SQL query does prevent some hazards.
I mitigate the risk of this up front by using a whitelist for each type of input and parameter queries for everything else.
The only tables that can be queried with this are the ones you specify.

## Prior Art

This project is inspired by [Postgrest](https://www.postgrest.org/).
Instead of something that runs everything on its own, though, I prefer a tool that integrates with my existing application.
