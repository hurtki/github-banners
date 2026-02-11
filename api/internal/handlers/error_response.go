package handlers

import (
	"encoding/json"
	"net/http"
)

// errror is used to write error in json
// if error, when marshaling appears, handles and logs it
func (h *BannersHandler) error(rw http.ResponseWriter, statusCode int, message string) {
	resJson := make(map[string]string)
	resJson["error"] = message
	res, err := json.Marshal(resJson)
	if err != nil {
		h.logger.Error("canl't marshal error to response")
		rw.WriteHeader(http.StatusInternalServerError)
		_, err := rw.Write([]byte("{\"error\": \"server error occured\"}"))
		if err != nil {
			h.logger.Warn("can't write error response")
			return
		}
	}
	rw.WriteHeader(statusCode)
	_, err = rw.Write(res)
	if err != nil {
		h.logger.Warn("can't write error response")
		return
	}
}
