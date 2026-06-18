package main

import (
	"encoding/json"

	"github.com/LH-10/TermBuzz/Client/models"
)

// modify later for better performance(maybe some binary encoding method directly instead of json)
func SerializeMessage(a models.ClientMessageFormat) ([]byte, error) {
	return json.Marshal(a)
}

func DeserializeMessage(obj []byte, a *models.ClientMessageFormat) error {
	return json.Unmarshal(obj, a)
}
