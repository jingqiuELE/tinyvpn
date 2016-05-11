package main

import (
	"fmt"
)

type SessionKey [6]byte

type Session struct {
	conn   Connection
	secret []byte
}

type SessionMap struct {
	//lock sync.Mutex
	Map map[SessionKey]Session
}

type IpSessionMap struct {
	ipToSession map[string]SessionKey
	sessionToIp map[SessionKey]string
}

func (m *IpSessionMap) getSession(ip string) SessionKey {
	return m.ipToSession[ip]
}

func (m *IpSessionMap) getIp(sessionKey SessionKey) string {
	return m.sessionToIp[sessionKey]
}

func (m *IpSessionMap) Add(ip string, sessionKey SessionKey) {
	m.ipToSession[ip] = sessionKey
	m.sessionToIp[sessionKey] = ip
}

func NewSessionMap() (s *SessionMap) {
	s = new(SessionMap)
	s.Map = make(map[SessionKey]Session)
	return s
}

func (m *SessionMap) Update(conn Connection, k SessionKey) {
	session, ok := m.Map[k]
	if ok == true {
		session.conn = conn
		fmt.Println("Created New connection for connection")
	}
}
