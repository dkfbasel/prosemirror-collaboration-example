package repository

import (
	"fmt"
	"strings"

	domain "dkfbasel.ch/orca/collaboration/src/domain"
	"github.com/pkg/errors"
)

// SaveComment will add the given comment information to the database and
// assign it to the given process
func (db *DB) SaveComment(comment *domain.CommentAdd) error {

	// do not handle prelimiary comments
	if strings.HasPrefix(comment.ID, "preliminary") {
		return nil
	}

	// insert the comment information into the database
	stmt := `[SQL-STATEMENT]`

	_, err := db.Session.Exec(stmt, comment.ID, comment.AuthorID, comment.Message,
		comment.Origin, comment.DocumentVersionID)
	if err != nil {
		return errors.Wrap(err, "could not add process comment")
	}

	return nil
}

// DeleteComment will flag the given comment as archived
func (db *DB) DeleteComment(comment *domain.CommentDelete) error {

	// do not handle prelimiary comments
	if strings.HasPrefix(comment.ID, "preliminary") {
		return nil
	}

	stmt := `[SQL-STATEMENT]`

	_, err := db.Session.Exec(stmt, comment.ID, comment.UserID, comment.Timestamp)
	if err != nil {
		return errors.Wrap(err, "could not remove process comment")
	}

	return nil
}

// SetCommentDone will flag the given comment as done
func (db *DB) SetCommentDone(comment *domain.CommentDone) error {

	// do not handle prelimiary comments
	if strings.HasPrefix(comment.ID, "preliminary") {
		return nil
	}

	stmt := `[SQL-STATEMENT]`

	_, err := db.Session.Exec(stmt, comment.ID, comment.UserID, comment.Timestamp)
	if err != nil {
		return errors.Wrap(err, "could not flag comment as done")
	}

	return nil
}

// SaveCommentReply will add the given comment reply
func (db *DB) SaveCommentReply(reply *domain.CommentReply) error {

	stmt := `[SQL-STATEMENT]`

	_, err := db.Session.Exec(stmt, reply.ReplyID, reply.CommentID, reply.AuthorID, reply.Message)
	if err != nil {
		return errors.Wrap(err, "could not save comment reply")
	}

	return nil
}

// DeleteCommentReply will flag the given comment as archived
func (db *DB) DeleteCommentReply(reply *domain.CommentDeleteReply) error {

	stmt := `[SQL-STATEMENT]`

	result, err := db.Session.Exec(stmt, reply.ReplyID, reply.CommentID,
		reply.UserID, reply.Timestamp)
	if err != nil {
		return errors.Wrap(err, "could not remove process comment reply")
	}

	rowCount, _ := result.RowsAffected()
	if rowCount != 1 {
		return fmt.Errorf("users may only delete their own replies")
	}

	return nil
}
