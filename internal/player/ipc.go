package player

import "encoding/json"

type command struct {
	Command   []any `json:"command"`
	RequestID int   `json:"request_id"`
}

// frame is a parsed inbound line from mpv (event or property-change).
type frame struct {
	Event  string `json:"event"`
	Name   string `json:"name"`
	Data   any    `json:"data"`
	Reason string `json:"reason"` // set on end-file: eof|stop|error|…
}

func encodeCommand(id int, args ...any) ([]byte, error) {
	b, err := json.Marshal(command{Command: args, RequestID: id})
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func parseLine(line []byte) (frame, error) {
	var f frame
	if err := json.Unmarshal(line, &f); err != nil {
		return frame{}, err
	}
	return f, nil
}
