package main

import (
	"encoding/json"

	"github.com/LH-10/TermBuzz/shared/models"
)

// modify later for better performance(maybe some binary encoding method directly instead of json)
func SerializeMessage(a models.ClientMessageFormatFut) ([]byte, error) {
	return json.Marshal(a)
}

func DeserializeMessage(obj []byte, a *models.ClientMessageFormatFut) error {
	return json.Unmarshal(obj, a)
}
