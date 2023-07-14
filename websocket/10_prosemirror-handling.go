package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	domain "dkfbasel.ch/orca/collaboration/src/domain"
	"dkfbasel.ch/orca/collaboration/src/internal/environment"
	image "dkfbasel.ch/orca/image/src/domain"
	"dkfbasel.ch/orca/pkg/logger"
	"go.uber.org/zap"
)

// ProsemirrorInitMessage is used when initializing a new room
type ProsemirrorInitMessage struct {
	DocumentID      string          `json:"documentid,omitempty"`
	DocumentSchema  json.RawMessage `json:"schema,omitempty"`
	DocumentVersion int64           `json:"version,omitempty"`
}

// handleProsemirrorInitMessage will handle all messages used to initialize
// a prosemirror document
func handleProsemirrorInitMessage(srv *environment.Services, room *WebsocketRoom,
	message *Message) {

	var payload ProsemirrorInitMessage
	err := json.Unmarshal(message.Payload, &payload)
	if err != nil {
		logger.DebugError("could not decode load message payload", err)
		return
	}

	// set the room version to the document version of the first
	// registering client
	if room.DocumentVersion == -1 {

		room.DocumentSchema = payload.DocumentSchema

		// check if we do already have a room starting id from redis
		cmd := srv.Redis.Get(message.DocumentID + "-starting-version")
		roomStartVersion, err := cmd.Int64()

		if err != nil {
			// use the version from the client if we do not have a start version yet

			room.DocumentVersion = payload.DocumentVersion
			srv.Redis.Set(message.DocumentID+"-starting-version", room.DocumentVersion, roomExpiration)

		} else {
			// use the redis document version
			// calculate the version of the room by adding the number
			// of steps currently in the redis buffer
			cmd2 := srv.Redis.LLen(message.DocumentID + "-steps")
			stepCount, err := cmd2.Result()
			if err != nil {
				logger.DebugError("could not fetch step count from redis", err,
					logger.String("documentid", message.DocumentID))
			}

			room.DocumentVersion = roomStartVersion + stepCount
		}
	}

	// reset the redis version to the client version if the client
	// version is newer than the redis version. Also reset the redis
	// cache for the given document version
	if room.DocumentVersion < payload.DocumentVersion {
		logger.Debug("client version is newer than redis version",
			zap.Int64("message-version", payload.DocumentVersion),
			zap.Int64("room-version", room.DocumentVersion))

		// set the room version to the document version received
		// from the client
		room.DocumentVersion = payload.DocumentVersion

		// reset the starting version
		srv.Redis.Set(message.DocumentID+"-starting-version", room.DocumentVersion, roomExpiration)

		// empty the saved -steps, -userids and -clientids
		srv.Redis.Del(message.DocumentID+"-steps", message.DocumentID+"-userids", message.DocumentID+"-clientids")

		logger.Debug("Redis room version got set to message version",
			zap.Int64("message-version", payload.DocumentVersion),
			zap.Int64("room-version", room.DocumentVersion),
			zap.String("documentId", message.DocumentID))
	}

}

// ProsemirrorStepMessage information
type ProsemirrorStepMessage struct {
	DocumentID      string            `json:"documentid,omitempty"`
	DocumentVersion int64             `json:"version,omitempty"`
	ClientID        int               `json:"clientID,omitempty"`
	Steps           []json.RawMessage `json:"steps,omitempty"`
	SaveImmediate   bool              `json:"save_immediate"`
}

type ProsemirrorStepResponse struct {
	Type    MessageType `json:"type"`
	Payload struct {
		BaseVersion   int64             `json:"base_version"`
		Version       int64             `json:"version"`
		ClientIDs     []int             `json:"clientIds,omitempty"`
		Steps         []json.RawMessage `json:"steps,omitempty"`
		FromInit      bool              `json:"from_init"`
		SaveImmediate bool              `json:"save_immediate"`
	} `json:"payload"`
}

type ProsemirrorInfoResponse struct {
	Type    MessageType `json:"type"`
	Payload struct {
		BaseVersion int64 `json:"base_version"`
		Version     int64 `json:"version"`
	} `json:"payload"`
}

// handleProsemirrorStepsMessage will handle prosemirror step transactions
func handleProsemirrorStepsMessage(srv *environment.Services, room *WebsocketRoom,
	message *Message, fromInit bool) {

	var payload ProsemirrorStepMessage
	err := json.Unmarshal(message.Payload, &payload)
	if err != nil {
		logger.DebugError("could not decode steps message payload", err)
		return
	}

	logger.Debug("message",
		zap.Int64("message-version", payload.DocumentVersion),
		zap.Int64("room-version", room.DocumentVersion))

	// inform client that a page reload is required
	// if the client version is newer
	if payload.DocumentVersion > room.DocumentVersion {
		logger.Debug("client version is newer, nothing to do",
			zap.Int64("message-version", payload.DocumentVersion),
			zap.Int64("room-version", room.DocumentVersion))

		// create a response message if the client version is newer than
		// the room version, to inform the client to reload the page
		response := ProsemirrorInfoResponse{}
		response.Type = MessageTypeProssemirrorReload

		// add the current server version
		response.Payload.BaseVersion = payload.DocumentVersion
		response.Payload.Version = room.DocumentVersion

		// encode the message for sending
		msg, err := json.Marshal(&response)
		if err != nil {
			logger.DebugError("could not encode reload page response", err)
			return
		}

		// send the info to reload the page back to the client
		message.Reply <- msg
		return
	}

	// save new steps, if client has the same version as the server
	// new steps are only counted for the client if they have been acknowledged
	// by the central authority (i.e. the room on the server)
	if payload.DocumentVersion == room.DocumentVersion {

		stepCount := len(payload.Steps)
		if stepCount == 0 {
			return
		}

		// log some information about the transactions
		logger.Debug("transactions received", zap.Int("count", stepCount),
			zap.Int64("message-version", payload.DocumentVersion),
			zap.Int64("room-version", room.DocumentVersion))

		// push all new steps to redis
		for _, step := range payload.Steps {

			// handle link and comment steps and check permissions
			err := handleSpecialSteps(srv, message.DocumentID,
				message.UserID, message.Permission, step)
			if err != nil {
				logger.DebugError("permission missmatch", err)
				return
			}

			asString := fmt.Sprintf("%s", step)

			// push the step to our document steps list
			cmd := srv.Redis.RPush(message.DocumentID+"-steps", asString)
			err = cmd.Err()
			if err != nil {
				logger.DebugError("could not store step in redis", err)
				return
			}

			// push the client id into our document clientids list
			cmd = srv.Redis.RPush(message.DocumentID+"-clientids", payload.ClientID)
			err = cmd.Err()
			if err != nil {
				logger.DebugError("could not store clientid in redis", err)
				return
			}

			// push the client id into our document clientids list
			cmd = srv.Redis.RPush(message.DocumentID+"-userids", message.UserID)
			err = cmd.Err()
			if err != nil {
				logger.DebugError("could not store userid in redis", err)
				return
			}

		}

		// expire keys after a certain time of inactivity
		cmd := srv.Redis.Expire(message.DocumentID+"-steps", roomExpiration)
		err := cmd.Err()
		if err != nil {
			logger.DebugError("could not set expiration time on step list", err)
			return
		}

		// expire keys after a certain time of inactivity
		cmd = srv.Redis.Expire(message.DocumentID+"-clientids", roomExpiration)
		err = cmd.Err()
		if err != nil {
			logger.DebugError("could not set expiration time on clientid list", err)
			return
		}

		// expire keys after a certain time of inactivity
		cmd = srv.Redis.Expire(message.DocumentID+"-userids", roomExpiration)
		err = cmd.Err()
		if err != nil {
			logger.DebugError("could not set expiration time on userid list", err)
			return
		}

		// expire keys after a certain time of inactivity
		cmd = srv.Redis.Expire(message.DocumentID+"-starting-version", roomExpiration)
		err = cmd.Err()
		if err != nil {
			logger.DebugError("could not set expiration time on start version", err)
			return
		}

		// initialize the response message
		stepMessage := ProsemirrorStepResponse{}
		stepMessage.Type = MessageTypeProsemirrorSteps

		// set an flag if the steps was send after an prosemirror-init event
		stepMessage.Payload.FromInit = fromInit

		// save the version, that the client must at least provide, to integrate
		// the new steps
		stepMessage.Payload.BaseVersion = room.DocumentVersion

		// save the new version of the room (current version plus steps applied)
		room.DocumentVersion = room.DocumentVersion + int64(stepCount)

		// add the new version number to the payload
		stepMessage.Payload.Version = room.DocumentVersion
		stepMessage.Payload.Steps = payload.Steps

		// add flag to notify if the save function on client side should be
		// executed immediately
		stepMessage.Payload.SaveImmediate = payload.SaveImmediate

		// add the client ids for all steps
		stepMessage.Payload.ClientIDs = make([]int, len(payload.Steps))
		for i := 0; i < len(payload.Steps); i++ {
			stepMessage.Payload.ClientIDs[i] = payload.ClientID
		}

		// send the new steps to all clients
		broadcast, err := json.Marshal(stepMessage)
		if err != nil {
			logger.DebugError("could not marshal steps broadcast", err)
			return
		}

		room.Broadcast <- broadcast
		return
	}

	// send missing steps to the client, if client version is
	// before server version
	if payload.DocumentVersion < room.DocumentVersion {
		logger.Debug("missing some steps", zap.Int64("room-version", room.DocumentVersion),
			zap.Int64("client version", payload.DocumentVersion))

		// nothing to do, if room version is on zero
		if room.DocumentVersion == 0 {
			return
		}

		// adapt the starting point from which we need to fetch steps
		// (redis room might not have all steps from the beginning of the document)
		fetchFrom := payload.DocumentVersion - room.DocumentVersion

		// get the last x steps. note that redis will return
		// the steps in inverse order and we need to resort it
		// again
		cmd2 := srv.Redis.LRange(message.DocumentID+"-steps", fetchFrom, 50000)
		steps, err := cmd2.Result()

		if err != nil {
			logger.Debug("could not fetch steps from redis", logger.Err(err))
			return
		}

		// send a response message to inform the client to reload the page
		// if no steps are in the redis cache. The steps got cleared during
		// initialisation as the current version was higher than
		// the room version
		if len(steps) == 0 {
			logger.Debug("no steps found, got probably reset",
				zap.Int64("room-version", room.DocumentVersion),
				zap.Int64("client-version", payload.DocumentVersion),
				zap.String("documentID", message.DocumentID))

			response := ProsemirrorInfoResponse{}
			response.Type = MessageTypeProssemirrorReload

			// add the current server version
			response.Payload.BaseVersion = payload.DocumentVersion
			response.Payload.Version = room.DocumentVersion

			// encode the message for sending
			msg, err := json.Marshal(&response)
			if err != nil {
				logger.DebugError("could not encode reload page response", err)
				return
			}

			// send the info to reload the page back to the client
			message.Reply <- msg
			return

		}

		// get the corresponding client ids
		cmd3 := srv.Redis.LRange(message.DocumentID+"-clientids", fetchFrom, 50000)
		clientIDs, err := cmd3.Result()
		if err != nil {
			logger.Debug("could not fetch clientids from redis", logger.Err(err))
			return
		}

		// create a response message. the client id should be
		// different from the effective client id, so that prosemirror
		// knows that this is from someone else
		response := ProsemirrorStepResponse{}
		response.Type = MessageTypeProsemirrorSteps

		// add the current server version
		response.Payload.BaseVersion = payload.DocumentVersion
		response.Payload.Version = room.DocumentVersion

		// add the steps to the response
		response.Payload.Steps = make([]json.RawMessage, len(steps))
		for i := range steps {
			response.Payload.Steps[i] = json.RawMessage(steps[i])
		}

		// add the client ids to the response
		response.Payload.ClientIDs = make([]int, len(clientIDs))
		for i := range clientIDs {
			response.Payload.ClientIDs[i], err = strconv.Atoi(clientIDs[i])
			if err != nil {
				logger.DebugError("could not convert client id to integer", err,
					logger.String("clientid", clientIDs[i]))
			}
		}

		// encode the message for sending
		msg, err := json.Marshal(&response)
		if err != nil {
			logger.DebugError("could not encode steps response", err)
			return
		}

		// send the missing steps back to the client
		message.Reply <- msg
		return
	}
}

// ProsemirrorMarkStep information
type ProsemirrorMarkStep struct {
	Type string          `json:"stepType"`
	From int             `json:"from"`
	To   int             `json:"to"`
	Mark ProsemirrorMark `json:"mark"`
}

type ProsemirrorMark struct {
	Type  string `json:"type"`
	Attrs struct {
		ID   string `json:"id"`
		File bool   `json:"file"`
		Link bool   `json:"link"`

		Name string `json:"name"`
		Url  string `json:"url"`

		Process   bool   `json:"process"`
		ProcessID string `json:"processId"`
	} `json:"attrs"`
}

// ProsemirrorCustomStep information
type ProsemirrorCustomStep struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ProsemirrorImageStep struct {
	Slice struct {
		Content []struct {
			Type  string `json:"type"`
			Attrs struct {
				ImageID string `json:"imageId"`
				Width   int    `json:"width"`
				Height  int    `json:"height"`
				Copy    bool   `json:"copy"`
			} `json:"attrs"`
		}
	}
}

// ProseMirrorReplaceStep is used when inserting pdf documents directly
type ProsemirrorReplaceStep struct {
	Slice struct {
		Content []ProsemirrorStepContent `json:"content"`
	}
}

// ProsemirrorStepContent is used to parse subcontent of steps
// attributes are not parsed, as we are mainly interested in the marks
type ProsemirrorStepContent struct {
	Type  string `json:"type"`
	Attrs struct {
		// attributes for pdf files
		DocumentID string `json:"documentId"`
		FileName   string `json:"fileName"`
		Src        string `json:"src"`

		// attributes for images
		ImageID string `json:"imageId"`
		Width   int    `json:"width"`
		Height  int    `json:"height"`
		Copy    bool   `json:"copy"`
	} `json:"attrs"`
	Marks   []*ProsemirrorMark        `json:"marks,omitempty"`
	Content []*ProsemirrorStepContent `json:"content,omitempty"`
}

// handleSpecialSteps is used to handle comment and link steps
func handleSpecialSteps(srv *environment.Services, documentId string, userId string,
	permission domain.Permission, step json.RawMessage) error {

	// user needs edit permissions to change anything
	if permission != domain.Edit {
		return fmt.Errorf("no permission to edit the document")
	}

	// parse links from marks
	if bytes.Contains(step, []byte(`"stepType":"addMark"`)) {

		// parse the mark content
		var stp ProsemirrorMarkStep
		err := json.Unmarshal(step, &stp)
		if err != nil {
			logger.DebugError("could not marshal mark step", err)
			return err
		}

		switch stp.Mark.Type {
		case "file", "weblink":
			err = srv.Postgres.SaveLink(documentId, stp.Mark.Type,
				stp.Mark.Attrs.ID, stp.Mark.Attrs.Url, stp.Mark.Attrs.Name)
			if err != nil {
				logger.DebugError("could not save link", err)
			}
			return err

		case "process":
			err = srv.Postgres.SaveLink(documentId, stp.Mark.Type,
				stp.Mark.Attrs.ID, stp.Mark.Attrs.Url, stp.Mark.Attrs.ProcessID)
			if err != nil {
				logger.DebugError("could not save link", err)
			}
			return err

		case "comment":
			return nil

		default:
			logger.Debug("mark handling not yet defined",
				logger.String("mark-type", stp.Mark.Type))
			return err
		}

	} else if bytes.Contains(step, []byte(`"stepType":"removeMark"`)) {

		// parse the mark content
		var stp ProsemirrorMarkStep
		err := json.Unmarshal(step, &stp)
		if err != nil {
			logger.DebugError("could not marshal mark step", err)
			return err
		}

		switch stp.Mark.Type {
		case "file", "weblink", "process":
			err = srv.Postgres.DeleteLink(documentId, stp.Mark.Attrs.ID,
				stp.Mark.Attrs.Url)
			if err != nil {
				logger.DebugError("could not save file link", err)
			}
			return err

		case "comment":
			// do not delete comments at the moment, if the comment
			// mark is removed
			return nil

		default:
			logger.Debug("mark handling not yet defined",
				logger.String("mark-type", stp.Mark.Type))
			return err
		}

	} else if bytes.Contains(step, []byte(`"stepType":"comment"`)) {

		// handle comments
		var stp ProsemirrorCustomStep
		err := json.Unmarshal(step, &stp)
		if err != nil {
			logger.DebugError("could not parse mark step", err)
			return err
		}

		switch stp.Type {
		case "addComment":
			var comment domain.CommentAdd
			err := json.Unmarshal(stp.Payload, &comment)
			if err != nil {
				logger.DebugError("could not parse add comment step", err)
				return err
			}
			comment.DocumentVersionID = documentId
			comment.AuthorID = userId
			err = srv.Postgres.SaveComment(&comment)
			if err != nil {
				logger.DebugError("could not add comment", err)
			}
			return err

		case "setCommentDone":
			var comment domain.CommentDone
			err := json.Unmarshal(stp.Payload, &comment)
			if err != nil {
				logger.DebugError("could not parse set comment as done step", err)
				return err
			}
			comment.UserID = userId
			err = srv.Postgres.SetCommentDone(&comment)
			if err != nil {
				logger.DebugError("could not set comment as done", err)
			}
			return err

		case "delete":
			var comment domain.CommentDelete
			err := json.Unmarshal(stp.Payload, &comment)
			if err != nil {
				logger.DebugError("could not parse delete comment step", err)
				return err
			}
			comment.UserID = userId
			err = srv.Postgres.DeleteComment(&comment)
			if err != nil {
				logger.DebugError("could not delete comment", err)
			}
			return err

		case "replyComment":
			var reply domain.CommentReply
			err := json.Unmarshal(stp.Payload, &reply)
			if err != nil {
				logger.DebugError("could not parse replyComment step", err)
				return err
			}
			reply.AuthorID = userId
			err = srv.Postgres.SaveCommentReply(&reply)
			if err != nil {
				logger.DebugError("could not reply to comment", err)
			}
			return err

		case "deleteCommentReply":
			var reply domain.CommentDeleteReply
			err := json.Unmarshal(stp.Payload, &reply)
			if err != nil {
				logger.DebugError("could not parse deleteCommentReply step", err)
				return err
			}
			reply.UserID = userId
			err = srv.Postgres.DeleteCommentReply(&reply)
			if err != nil {
				logger.DebugError("could not add comment", err)
			}
			return err

		default:
			logger.Debug("comment type handling not yet defined",
				logger.String("comment-type", stp.Type))
		}

	} else if bytes.Contains(step, []byte(`"stepType":"picture"`)) {

		// handle pictures
		var stp ProsemirrorCustomStep
		err := json.Unmarshal(step, &stp)
		if err != nil {
			logger.DebugError("could not parse custom step", err)
			return err
		}

		switch stp.Type {
		// tell the image service to create a copy of the given image
		case "copyPicture":
			var copyData domain.ImageCopy
			err := json.Unmarshal(stp.Payload, &copyData)
			if err != nil {
				logger.DebugError("could not parse picture step", err)
				return err
			}

			request := image.DuplicateImageRequest{
				Id:    copyData.OriginalID,
				NewId: copyData.ImageID,
			}
			_, err = srv.Image.DuplicateImage(context.Background(), &request)
			if err != nil {
				logger.DebugError("could not copy picture", err)
			}

		default:
			break
		}

	} else if bytes.Contains(step, []byte(`"stepType":"replace"`)) {

		// parse the step content
		var stp ProsemirrorReplaceStep
		err := json.Unmarshal(step, &stp)
		if err != nil {
			logger.DebugError("could not unmarshal replace step", err)
			return err
		}

		go func() {

			links := make(map[string]Link)

			// find pdf block and extract document id to save link for later
			// access to the document
			for _, content := range stp.Slice.Content {

				switch content.Type {
				case "pdf":
					// ignore pdf blocks without document id
					if content.Attrs.DocumentID == "" {
						continue
					}

					downloadUrl := fmt.Sprintf("/download/process/%s", content.Attrs.DocumentID)
					links["pdf-"+content.Attrs.DocumentID] = Link{
						ID:   content.Attrs.DocumentID,
						Type: "pdf",
						URL:  downloadUrl,
						Name: content.Attrs.FileName,
					}
					continue

				// handle picture blocks
				case "picture":
					// save the image link to the database
					imageURL := fmt.Sprintf("/image/process/%s", content.Attrs.ImageID)
					links["image-"+content.Attrs.ImageID] = Link{
						ID:   content.Attrs.ImageID,
						Type: "image",
						URL:  imageURL,
						Name: "",
					}
					continue

				default:
					// extract all link marks from the content
					extractLinks(&content, links)
				}

			}

			// save all links in the database
			for _, link := range links {
				err = srv.Postgres.SaveLink(documentId, link.Type,
					link.ID, link.URL, link.Name)
				if err != nil {
					logger.Error("could not save link", err, zap.Any("link", link))
				}
			}
		}()
	}

	return nil

}
