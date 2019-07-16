package softether

import (
	"errors"
	"strconv"
	"strings"
)

const Head = "SecureNAT"

type session struct {
	name     string
	username string
	kind     string
	seq      int
}

func (s *session) String() string {
	return s.name
}

type HubSessions struct {
	hub      string
	sessions map[string]session
}

func NewHubSessions(hubname string) (*HubSessions, error) {
	S := &HubSessions{
		hub:      hubname,
		sessions: make(map[string]session),
	}
	err := S.Sync()
	return S, err
}
func (S *HubSessions) GetSessionBySid(username string) (session, error) {
	s, ok := S.sessions[username]
	if ok {
		return s, nil
	} else {
		return session{}, errors.New("username is not existed")
	}
}
func (S *HubSessions) DeleteSessionBySid(username string) error {

	API, _ := PoolGetConn()
	session, ok := S.sessions[username]
	if ok {
		if _, err := API.DeleteSession(S.hub, session.name); err != nil {
			return err
		}
		delete(S.sessions, username)
	} else {
		return errors.New("username is not existed")
	}
	return nil
}
func (S *HubSessions) Sync() error {
	API, _ := PoolGetConn()
	out, err := API.ListSessions(S.hub)
	if err != nil {
		return err
	}
	names, ok := out["Username"].([]interface{})
	if ok && len(names) > 1 {
		for index, name := range names {
			if name != Head {
				sessionname := out["Name"].([]interface{})[index].(string)
				selflogger.Info("SessionName[%s] UserName[%s]", sessionname, names)
				str := strings.Split(sessionname, "-")
				seqid, _ := strconv.Atoi(str[3])
				session := session{
					name:     sessionname,
					username: name.(string),
					kind:     str[2],
					seq:      seqid,
				}
				S.sessions[name.(string)] = session
			}
		}
	}
	return nil
}
func (S *HubSessions) Size() int {
	return len(S.sessions)
}
