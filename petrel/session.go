package main

import (
	"crypto/rand"
	"fmt"
)

type Session struct {
	conn   Connection
	secret []byte
}

type SessionKey [6]byte

func NewSessionKey() (m *SessionKey, err error) {
	m = new(SessionKey)
	_, err = rand.Read(m[:])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	return
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

func (m *SessionMap) Update(k SessionKey, conn Connection) {
	session, ok := m.Map[k]
	if ok == true {
		session.conn = conn
		fmt.Println("Created New connection for connection")
	}
}
