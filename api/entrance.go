package api

import (
	"database/sql"
	"net/http"

	"github.com/keydotcat/backend/managers"
	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

var TEST_MODE = false

type apiHandler struct {
	db   *sql.DB
	sm   managers.SessionMgr
	mail *mailer
	csrf csrf
}

func NewAPIHandler(c Conf) (http.Handler, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}
	ah := apiHandler{}
	ah.db, err = sql.Open("postgres", c.DB)
	if err != nil {
		return nil, util.NewErrorf("Could not connect to db '%s': %s", c.DB, err)
	}
	switch {
	case c.MailSMTP != nil:
		ah.mail, err = newMailer(c.Url, TEST_MODE, managers.NewMailMgrSMTP(c.MailSMTP.Server, c.MailSMTP.User, c.MailSMTP.Password, c.MailFrom))
	case c.MailSparkpost != nil:
		ah.mail, err = newMailer(c.Url, TEST_MODE, managers.NewMailMgrSparkpost(c.MailSparkpost.Key, c.MailFrom))
	default:
		return nil, util.NewErrorf("No mail manager defined in the configuration")
	}
	if err != nil {
		return nil, util.NewErrorf("Could not create mailer: %s", err)
	}
	ah.sm, err = managers.NewSessionMgrRedis(c.SessionRedis.Server, c.SessionRedis.DBId)
	if err != nil {
		return nil, util.NewErrorf("Could not connect to redis at %s: %s", c.SessionRedis.Server, err)
	}
	var blockKey []byte
	if len(c.Csrf.BlockKey) > 0 {
		blockKey = []byte(c.Csrf.BlockKey)
	}
	ah.csrf = newCsrf([]byte(c.Csrf.HashKey), blockKey)
	return ah, nil
}

func (ah apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(models.AddDBToContext(r.Context(), ah.db))
	head := ""
	head, r.URL.Path = shiftPath(r.URL.Path)
	if head == "auth" {
		ah.authRoot(w, r)
		return
	}
	r = ah.authorizeRequest(w, r)
	//From here on you need to be authenticated
	switch head {
	case "session":
		ah.sessionRoot(w, r)
	case "user":
		ah.userRoot(w, r)
	case "team":
		ah.teamRoot(w, r)
	default:
		httpErr(w, util.NewErrorFrom(ErrNotFound))
	}
}
