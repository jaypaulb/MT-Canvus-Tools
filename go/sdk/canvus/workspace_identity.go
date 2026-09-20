package canvus

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// UnmarshalJSON preserves index presence without breaking the legacy integer
// field. A sparse/malformed identity must never become a command for index zero.
func (w *Workspace) UnmarshalJSON(data []byte) error {
	type plain Workspace
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("decode Workspace: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("decode workspace presence: %w", err)
	}
	raw, present := fields["index"]
	value.IndexPresent = present
	value.IndexNull = present && bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
	*w = Workspace(value)
	return nil
}

func (w Workspace) validIndex() bool { return w.IndexPresent && !w.IndexNull && w.Index >= 0 }
