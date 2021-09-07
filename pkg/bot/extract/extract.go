package extract

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ErrIndexNotFound indicates the requested index does not exist
type ErrIndexNotFound error

// NewErrIndexNotFound builds an ErrIndexNotFound from a field name and index
func NewErrIndexNotFound(field string, index int) error {
	return ErrIndexNotFound(fmt.Errorf("could not find %s with index %d", field, index))
}

// MessageString extracts strings from messages
type MessageString func(m *discordgo.Message) (string, error)

// EmbedFieldValue returns a MessageString extracting the value found in the
// embed field at the respective indices
func EmbedFieldValue(embedIndex int, fieldIndex int) MessageString {
	return func(m *discordgo.Message) (string, error) {
		embeds := m.Embeds
		if len(embeds) < embedIndex+1 {
			return "", NewErrIndexNotFound("embed", embedIndex)
		}
		fields := embeds[embedIndex].Fields
		if len(fields) < fieldIndex+1 {
			return "", NewErrIndexNotFound("field", fieldIndex)
		}
		return fields[fieldIndex].Value, nil
	}
}
