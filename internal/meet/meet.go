package meet

import (
	"errors"
	"sync"

	log "github.com/pion/ion-log"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"github.com/randsoy/ct-sfu/internal/meet/conf"
)

// Peer p
type Peer struct {
	*sfu.Peer
	ch Chan
}

// Meet server
type Meet struct {
	c   *conf.Config
	sfu *sfu.SFU

	peers map[string]*Peer
	mu    sync.RWMutex
}

// New new server
func New(c *conf.Config) *Meet {
	s := &Meet{
		c:     c,
		sfu:   sfu.NewSFU(c.Config),
		peers: make(map[string]*Peer),
	}
	dc := s.sfu.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)
	return s
}

func (m *Meet) getPeer(ch Chan) *Peer {
	m.mu.Lock()
	defer m.mu.Unlock()
	mid := ch.MID()
	p := m.peers[mid]
	if p == nil {
		p = &Peer{
			Peer: sfu.NewPeer(m.sfu),
		}
		m.peers[mid] = p
	}
	if ch != nil {
		p.ch = ch
	}
	return p
}

func (m *Meet) delPeer(mid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.peers, mid)
}

// Close peer
func (m *Meet) Close() {

}

// Join peer
func (m *Meet) Join(ch Chan, join Join) (interface{}, error) {
	peer := m.getPeer(ch)

	// notify user of new offer
	peer.OnOffer = func(offer *webrtc.SessionDescription) {
		if err := peer.ch.Notify("offer", offer); err != nil {
			log.Errorf("error sending offer %s", err)
		}
		log.Infof("reply OnOffer: %v", offer)
	}

	// notify user of new ice candidate
	peer.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
		if err := peer.ch.Notify("trickle", Trickle{
			Candidate: *candidate,
			Target:    target,
		}); err != nil {
			log.Errorf("error sending ice candidate %s", err)
		}
		log.Infof("reply OnIceCandidate: %v", candidate)
	}

	peer.OnICEConnectionStateChange = func(state webrtc.ICEConnectionState) {
		// log.Infof("send ice connection state to [%s]: %v", ch.MID(), state)
	}

	answer, err := peer.Join(join.Sid, join.Offer)
	if err != nil {
		log.Errorf("join error: %v", err)
		return nil, err
	}

	// return answer
	log.Infof("reply join: %v", answer)
	return answer, nil
}

// Offer peer
func (m *Meet) Offer(ch Chan, negotiation Negotiation) (interface{}, error) {
	peer := m.getPeer(ch)
	if peer == nil {
		log.Warnf("peer not found, mid=%s", ch.MID())
		return nil, errors.New("peer not found")
	}

	answer, err := peer.Answer(negotiation.Desc)
	if err != nil {
		log.Errorf("peer.Answer: %v", err)
		return nil, err
	}
	log.Infof("reply Offer: %v", answer)
	return answer, nil
}

// Answer peer
func (m *Meet) Answer(ch Chan, negotiation Negotiation) (interface{}, error) {
	peer := m.getPeer(ch)
	if peer == nil {
		log.Warnf("peer not found, mid=%s", ch.MID())
		return nil, errors.New("peer not found")
	}

	if err := peer.SetRemoteDescription(negotiation.Desc); err != nil {
		log.Errorf("set remote description error: %v", err)
		return nil, err
	}
	return nil, nil
}

// Trickle peer
func (m *Meet) Trickle(ch Chan, trickle Trickle) (interface{}, error) {
	peer := m.getPeer(ch)
	if peer == nil {
		log.Warnf("peer not found, mid=%s", ch.MID())
		return nil, errors.New("peer not found")
	}

	if err := peer.Trickle(trickle.Candidate, trickle.Target); err != nil {
		return nil, err
	}
	return nil, nil
}

// Leave peer
func (m *Meet) Leave(ch Chan, level Leave) (interface{}, error) {
	peer := m.getPeer(ch)
	if peer == nil {
		log.Warnf("peer not found, mid=%s", ch.MID())
		return nil, errors.New("peer not found")
	}
	m.delPeer(ch.MID())
	if err := peer.Peer.Close(); err != nil {
		return nil, err
	}
	return nil, nil
}
