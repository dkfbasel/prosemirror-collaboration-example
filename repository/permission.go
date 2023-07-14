package repository

import (
	"errors"

	domain "dkfbasel.ch/orca/collaboration/src/domain"
	"dkfbasel.ch/orca/pkg/database"
	"dkfbasel.ch/orca/pkg/logger"
)

// FetchPermission will check for permissions for the given user on the document
func (db *DB) FetchPermission(documentVersionId, userID string) (domain.Permission, error) {

	if documentVersionId == "" || userID == "" {
		return domain.None, errors.New("missing parameters")
	}

	// check if document id exists
	var isDocumentVersionID bool
	stmt := `[SQL-STATEMENT]`
	err := db.Session.Get(&isDocumentVersionID, stmt, documentVersionId)
	if database.NotNoResultsError(database.NewError(err)) {
		logger.Debug("verify if document id failed")
		return domain.None, err
	}

	// do not allow access if the document is not in draft status
	if !isDocumentVersionID {
		return domain.None, nil
	}

	// check if user is a sysadmin who is allowed to do everything
	var isSysadmin bool
	stmt = `[SQL-STATEMENT]`
	err = db.Session.Get(&isSysadmin, stmt, userID)

	if database.NotNoResultsError(database.NewError(err)) {
		logger.Debug("error on sys admin check")
		return domain.None, err
	}

	if isSysadmin {
		return domain.Edit, nil
	}

	// check if the user has folder manage permissions to allow them to write
	// the process
	var hasManagePermissions bool
	stmt = `[SQL-STATEMENT]`
	err = db.Session.Get(&hasManagePermissions, stmt, documentVersionId, userID)

	// allow users with folder manage permissions to edit the document
	// continue on error and no results error, since we also need to check
	// if the user is a contributor
	if err == nil && hasManagePermissions {
		return domain.Edit, nil
	}

	// check if user is a contributor of the given document
	var isContributor bool
	stmt = `[SQL-STATEMENT]`
	err = db.Session.Get(&isContributor, stmt, documentVersionId, userID)
	if database.NotNoResultsError(database.NewError(err)) {
		return domain.None, err
	}

	if isContributor {
		return domain.Edit, nil
	}

	return domain.None, nil
}
