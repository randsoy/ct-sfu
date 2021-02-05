package meet

import "github.com/pion/webrtc/v3"

// Join message sent when initializing a peer connection
type Join struct {
	Sid   string                    `json:"sid"` // SID room id
	Offer webrtc.SessionDescription `json:"offer"`
}

// Trickle message sent when renegotiating the peer connection
type Trickle struct {
	Target    int                     `json:"target"`
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

// Negotiation message sent when renegotiating the peer connection
type Negotiation struct {
	Desc webrtc.SessionDescription `json:"desc"`
}

// Leave message sent when leave the peer connection
type Leave struct {
}
