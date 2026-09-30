package constants

const (
	RequestID = 9800 + iota
	RequestPeerConnection
	RequestPeerList //ask for list of clients connected to the signaling server
	PeerChat
	Blank
	SDPExchange
	SDPAnswer
	Candidate
)

const (
	ClientIDResponse = 8800 + iota
)
