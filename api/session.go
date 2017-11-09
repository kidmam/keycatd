package api

import (
	"net/http"

	"github.com/keydotcat/backend/managers"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

func (ah apiHandler) sessionRoot(w http.ResponseWriter, r *http.Request) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		//TODO: list all sessions
		return util.NewErrorFrom(ErrNotFound)
	} else {
		switch r.Method {
		case "GET":
			return ah.sessionGetToken(w, r, head)
		case "DELETE":
			return ah.sessionDeleteToken(w, r, head)
		}
	}
	return util.NewErrorFrom(ErrNotFound)
}

type sessionResponse struct {
	managers.Session
	Csrf       string `json:"csrf,omitempty"`
	StoreToken string `json:"store_token,omitempty"`
}

// GET /session/:token
func (ah apiHandler) sessionGetToken(w http.ResponseWriter, r *http.Request, tid string) error {
	currentSession := ctxGetSession(r.Context())
	if currentSession.Id == tid {
		return jsonResponse(w, sessionResponse{currentSession, ctxGetCsrf(r.Context()), currentSession.StoreToken})
	}
	currentUser := ctxGetUser(r.Context())
	s, err := ah.sm.GetSession(tid)
	if err != nil {
		return util.NewErrorFrom(models.ErrDoesntExist)
	}
	if s.UserId != currentUser.Id {
		return util.NewErrorFrom(models.ErrDoesntExist)
	}
	return jsonResponse(w, s)
}

// DELETE /session/:token
func (ah apiHandler) sessionDeleteToken(w http.ResponseWriter, r *http.Request, tid string) error {
	currentSession := ctxGetSession(r.Context())
	if err := ah.sm.DeleteSession(currentSession.Id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}