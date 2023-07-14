package prosemirror

import (
	"github.com/dkfbasel/protobuf/types/nullstring"
	"github.com/dkfbasel/protobuf/types/timestamp"
)

// CommentAdd ...
type CommentAdd struct {
	DocumentVersionID string                 `json:"-"`
	ID                string                 `json:"id"`
	AuthorID          string                 `json:"authorId"`
	Message           string                 `json:"message"`
	Origin            string                 `json:"origin"`
	Done              *timestamp.Timestamp   `json:"done"`
	DoneBy            *nullstring.NullString `json:"doneBy"`
	Archived          *timestamp.Timestamp   `json:"archived"`
	ArchivedBy        *nullstring.NullString `json:"archivedBy"`
	Timestamp         *timestamp.Timestamp   `json:"timestamp"`
}

// CommentDelete ...
type CommentDelete struct {
	ID        string              `json:"id"`
	UserID    string              `json:"userId"`
	Timestamp timestamp.Timestamp `json:"timestamp"`
}

// CommentSetDone ...
type CommentDone struct {
	ID        string              `json:"id"`
	UserID    string              `json:"userId"`
	Timestamp timestamp.Timestamp `json:"done"`
}

// CommentReply ...
type CommentReply struct {
	ReplyID   string              `json:"id"`
	CommentID string              `json:"commentId"`
	AuthorID  string              `json:"authorId"`
	Message   string              `json:"message"`
	Timestamp timestamp.Timestamp `json:"timestamp"`
}

// CommentDeleteReply ...
type CommentDeleteReply struct {
	CommentID string              `json:"commentId"`
	ReplyID   string              `json:"replyId"`
	UserID    string              `json:"userId"`
	Timestamp timestamp.Timestamp `json:"timestamp"`
}
