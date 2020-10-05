package anycable

import "encoding/json"

type Socket struct {
	transmissions []string
}

func (s *Socket) Write(t interface{}) error {
	json, err := json.Marshal(t)
	if err != nil {
		return err
	}
	s.transmissions = append(s.transmissions, string(json))
	return nil
}
