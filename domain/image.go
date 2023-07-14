package prosemirror

// --- IMAGE EVENTS ---

type ImageCopy struct {
	ImageID    string `json:"imageId"`    // id for the new image
	OriginalID string `json:"originalId"` // id of the original image
}
