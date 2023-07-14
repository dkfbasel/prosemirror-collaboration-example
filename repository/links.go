package repository

import "dkfbasel.ch/orca/pkg/database"

// SaveLink will save the given link information
func (db *DB) SaveLink(blockID, linkType, linkID, url, title string) error {

	// avoid duplicate links
	stmt := `[SQL-STATEMENT]`

	var exists bool
	err := db.Session.Get(&exists, stmt, blockID, linkID, linkType)
	if database.NotNoResultsError(database.NewError(err)) {
		return err
	}

	if exists {
		return nil
	}

	stmt = `[SQL-STATEMENT]`

	_, err = db.Session.Exec(stmt, blockID, linkType, linkID, url, title)
	return err
}

// DeleteLink will mark the given link as archived. Note that we need to
// preserve the links to allow users to access links from restored document versions
func (db *DB) DeleteLink(blockID, linkID, url string) error {

	stmt := `[SQL-STATEMENT]`

	_, err := db.Session.Exec(stmt, blockID, linkID, url)
	return err
}
