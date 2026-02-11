package handlers

import (
	"encoding/json"
	"net/http"
)

// errror is used to write error in json
// if error, when marshaling appears, handles and logs it
func (h *BannersHandler) error(rw http.ResponseWriter, statusCode int, message string) {
	fn := "internal.handlers.BannersHandler.error"
	rw.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(map[string]string{"error": message})

	if err != nil {
		h.logger.Error("can't marshal error response", "err", err, "source", fn)
		rw.WriteHeader(http.StatusInternalServerError)
		_, err := rw.Write([]byte("{\"error\": \"server error occured\"}"))
		if err != nil {
			h.logger.Warn("can't write error response", "err", err, "source", fn)
		}
		return
	}

	rw.WriteHeader(statusCode)
	_, err = rw.Write(res)
	if err != nil {
		h.logger.Warn("can't write error response", "err", err, "source", fn)
		return
	}
}
