package bartlett

// A Table represents a table in the database.
// Name is required.
// Writable determines whether the table allows INSERT, UPDATE, and DELETE queries. Default is read-only.
// UserID is the name of column containing user IDs. It should match the output of the UserIDProvider passed to Bartlett.
// If UserID is left blank, all rows will be available regardless of the UserIDProvider.
type Table struct {
	columns  []string
	Name     string
	Writable bool
	UserID   string
}
