package websocket

// Link contains information on links in documents
type Link struct {
	ID   string
	Type string
	URL  string
	Name string
}

// extractLinks goes through the given content recursively extracts all link marks
// into the given slice of marks
func extractLinks(content *ProsemirrorStepContent, links map[string]Link) {

	// extract links from marks
	for _, mark := range content.Marks {
		if mark.Type == "file" || mark.Type == "weblink" || mark.Type == "process" {
			_, ok := links[mark.Attrs.ID]
			if !ok {
				link := Link{
					ID:   mark.Attrs.ID,
					Type: mark.Type,
					URL:  mark.Attrs.Url,
					Name: mark.Attrs.Name,
				}
				links[link.ID] = link
			}
		}
	}

	// do the same thing recursively for all childrens with content
	for i := range content.Content {
		extractLinks(content.Content[i], links)
	}
}
