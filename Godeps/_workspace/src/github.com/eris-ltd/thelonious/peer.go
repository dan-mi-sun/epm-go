package thelonious

import (
	"bytes"
	"container/list"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monkchain"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monklog"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monktrie"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monkutil"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monkwire"
)

var peerlogger = monklog.NewLogger("PEER")

const (
	// The size of the output buffer for writing messages
	outputBufferSize = 50
	// Current protocol version
	ProtocolVersion = 33
	// Current P2P version
	P2PVersion = 0
	// Thelonious network version
	NetVersion = 0
	// Interval for ping/pong message
	pingPongTimer = 2 * time.Second
)

type DiscReason byte

const (
	// Values are given explicitly instead of by iota because these values are
	// defined by the wire protocol spec; it is easier for humans to ensure
	// correctness when values are explicit.
	DiscReRequested  = 0x00
	DiscReTcpSysErr  = 0x01
	DiscBadProto     = 0x02
	DiscBadPeer      = 0x03
	DiscTooManyPeers = 0x04
	DiscConnDup      = 0x05

	DiscGenesisErr = 0x06
	DiscProtoErr   = 0x07
	DiscQuitting   = 0x08
)

var discReasonToString = []string{
	"requested",
	"TCP sys error",
	"bad protocol",
	"useless peer",
	"too many peers",
	"already connected",
	"wrong genesis block",
	"incompatible network",
	"quitting",
}

func (d DiscReason) String() string {
	if len(discReasonToString) < int(d) {
		return "Unknown"
	}

	return discReasonToString[d]
}

// Peer capabilities
type Caps byte

const (
	CapPeerDiscTy Caps = 1 << iota
	CapTxTy
	CapChainTy

	CapDefault = CapChainTy | CapTxTy | CapPeerDiscTy
)

var capsToString = map[Caps]string{
	CapPeerDiscTy: "Peer discovery",
	CapTxTy:       "Transaction relaying",
	CapChainTy:    "Block chain relaying",
}

func (c Caps) IsCap(cap Caps) bool {
	return c&cap > 0
}

func (c Caps) String() string {
	var caps []string
	if c.IsCap(CapPeerDiscTy) {
		caps = append(caps, capsToString[CapPeerDiscTy])
	}
	if c.IsCap(CapChainTy) {
		caps = append(caps, capsToString[CapChainTy])
	}
	if c.IsCap(CapTxTy) {
		caps = append(caps, capsToString[CapTxTy])
	}

	return strings.Join(caps, " | ")
}

type Peer struct {
	// Thelonious interface
	thelonious *Thelonious
	// Net connection
	conn net.Conn
	// Output queue which is used to communicate and handle messages
	outputQueue chan *monkwire.Msg
	// Quit channel
	quit chan bool
	// Determines whether it's an inbound or outbound peer
	inbound bool
	// Flag for checking the peer's connectivity state
	connected  int32
	disconnect int32
	// Last known message send
	lastSend time.Time
	// Indicated whether a verack has been send or not
	// This flag is used by writeMessage to check if messages are allowed
	// to be send or not. If no version is known all messages are ignored.
	versionKnown bool
	statusKnown  bool

	// Last received pong message
	lastPong           int64
	lastBlockReceived  time.Time
	doneFetchingHashes bool

	host             []byte
	port             uint16
	caps             Caps
	td               *big.Int
	bestHash         []byte
	lastReceivedHash []byte
	requestedHashes  [][]byte

	// This peer's public key
	pubkey []byte

	// Indicated whether the node is catching up or not
	catchingUp      bool
	diverted        bool
	blocksRequested int

	version string

	// We use this to give some kind of pingtime to a node, not very accurate, could be improved.
	pingTime      time.Duration
	pingStartTime time.Time

	lastRequestedBlock *monkchain.Block

	protocolCaps *monkutil.Value

	mut sync.RWMutex
}

func NewPeer(conn net.Conn, th *Thelonious, inbound bool) *Peer {
	pubkey := th.KeyManager().PublicKey()[1:]

	return &Peer{
		outputQueue:        make(chan *monkwire.Msg, outputBufferSize),
		quit:               make(chan bool),
		thelonious:         th,
		conn:               conn,
		inbound:            inbound,
		disconnect:         0,
		connected:          1,
		port:               30303,
		pubkey:             pubkey,
		blocksRequested:    10,
		caps:               th.ServerCaps(),
		version:            th.ClientIdentity().String(),
		protocolCaps:       monkutil.NewValue(nil),
		td:                 big.NewInt(0),
		doneFetchingHashes: true,
	}
}

func NewOutboundPeer(addr string, th *Thelonious, caps Caps) *Peer {
	p := &Peer{
		outputQueue:        make(chan *monkwire.Msg, outputBufferSize),
		quit:               make(chan bool),
		thelonious:         th,
		inbound:            false,
		connected:          0,
		disconnect:         0,
		port:               30303,
		caps:               caps,
		version:            th.ClientIdentity().String(),
		protocolCaps:       monkutil.NewValue(nil),
		td:                 big.NewInt(0),
		doneFetchingHashes: true,
	}

	// Set up the connection in another goroutine so we don't block the main thread
	go func() {
		conn, err := p.Connect(addr)
		if err != nil {
			peerlogger.Debugln("Connection to peer failed. Giving up.", err)
			p.Stop()
			return
		}
		p.conn = conn

		// Atomically set the connection state
		atomic.StoreInt32(&p.connected, 1)
		atomic.StoreInt32(&p.disconnect, 0)

		p.Start()
	}()

	return p
}

func (self *Peer) Connect(addr string) (conn net.Conn, err error) {
	const maxTries = 3
	for attempts := 0; attempts < maxTries; attempts++ {
		conn, err = net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			time.Sleep(time.Duration(attempts*20) * time.Second)
			continue
		}

		// Success
		return
	}

	return
}

// Getters
func (p *Peer) PingTime() string {
	return p.pingTime.String()
}
func (p *Peer) Inbound() bool {
	return p.inbound
}
func (p *Peer) LastSend() time.Time {
	return p.lastSend
}
func (p *Peer) LastPong() int64 {
	return p.lastPong
}
func (p *Peer) Host() []byte {
	return p.host
}
func (p *Peer) Port() uint16 {
	return p.port
}
func (p *Peer) Version() string {
	return p.version
}
func (p *Peer) Connected() *int32 {
	return &p.connected
}
func (p *Peer) StatusKnown() bool {
	p.mut.Lock()
	defer p.mut.Unlock()
	return p.statusKnown
}

// Setters
func (p *Peer) SetVersion(version string) {
	p.version = version
}

// Outputs any RLP encoded data to the peer
func (p *Peer) QueueMessage(msg *monkwire.Msg) {
	if atomic.LoadInt32(&p.connected) != 1 {
		return
	}
	p.outputQueue <- msg
}

func (p *Peer) writeMessage(msg *monkwire.Msg) {
	// Ignore the write if we're not connected
	if atomic.LoadInt32(&p.connected) != 1 {
		return
	}

	if !p.versionKnown {
		switch msg.Type {
		case monkwire.MsgHandshakeTy: // Ok
		default: // Anything but ack is allowed
			return
		}
	} else {
		/*
			if !p.statusKnown {
				switch msg.Type {
				case monkwire.MsgStatusTy: // Ok
				default: // Anything but ack is allowed
					return
				}
			}
		*/
	}

	peerlogger.DebugDetailf("(%v) <= %v\n", p.conn.RemoteAddr(), formatMessage(msg))

	err := monkwire.WriteMessage(p.conn, msg)
	if err != nil {
		peerlogger.Debugln(" Can't send message:", err)
		// Stop the client if there was an error writing to it
		p.Stop()
		return
	}
}

// Outbound message handler. Outbound messages are handled here
func (p *Peer) HandleOutbound() {
	// The ping timer. Makes sure that every 2 minutes a ping is send to the peer
	pingTimer := time.NewTicker(pingPongTimer)
	serviceTimer := time.NewTicker(10 * time.Second)

out:
	for {
	skip:
		select {
		// Main message queue. All outbound messages are processed through here
		case msg := <-p.outputQueue:
			if !p.StatusKnown() {
				switch msg.Type {
				case monkwire.MsgGetTxsTy, monkwire.MsgTxTy, monkwire.MsgGetBlockHashesTy, monkwire.MsgBlockHashesTy, monkwire.MsgGetBlocksTy, monkwire.MsgBlockTy:
					break skip
				}
			}

			p.writeMessage(msg)
			p.setLastSend()

		// Ping timer
		case <-pingTimer.C:
			/*
				timeSince := time.Since(time.Unix(p.lastPong, 0))
				if !p.pingStartTime.IsZero() && p.lastPong != 0 && timeSince > (pingPongTimer+30*time.Second) {
					peerlogger.Infof("Peer did not respond to latest pong fast enough, it took %s, disconnecting.\n", timeSince)
					p.Stop()
					return
				}
			*/
			p.writeMessage(monkwire.NewMessage(monkwire.MsgPingTy, ""))
			p.setPingStartTime()

		// Service timer takes care of peer broadcasting, transaction
		// posting or block posting
		case <-serviceTimer.C:
			p.QueueMessage(monkwire.NewMessage(monkwire.MsgGetPeersTy, ""))

		case <-p.quit:
			// Break out of the for loop if a quit message is posted
			break out
		}
	}

clean:
	// This loop is for draining the output queue and anybody waiting for us
	for {
		select {
		case <-p.outputQueue:
			// TODO
		default:
			break clean
		}
	}
}

func formatMessage(msg *monkwire.Msg) (ret string) {
	ret = fmt.Sprintf("%v %v", msg.Type, msg.Data)

	/*
		XXX Commented out because I need the log level here to determine
		if i should or shouldn't generate this message
	*/
	/*
		switch msg.Type {
		case monkwire.MsgPeersTy:
			ret += fmt.Sprintf("(%d entries)", msg.Data.Len())
		case monkwire.MsgBlockTy:
			b1, b2 := monkchain.NewBlockFromRlpValue(msg.Data.Get(0)), monkchain.NewBlockFromRlpValue(msg.Data.Get(msg.Data.Len()-1))
			ret += fmt.Sprintf("(%d entries) %x - %x", msg.Data.Len(), b1.Hash()[0:4], b2.Hash()[0:4])
		case monkwire.MsgBlockHashesTy:
			h1, h2 := msg.Data.Get(0).Bytes(), msg.Data.Get(msg.Data.Len()-1).Bytes()
			ret += fmt.Sprintf("(%d entries) %x - %x", msg.Data.Len(), h1, h2)
		}
	*/

	return
}

// Inbound handler. Inbound messages are received here and passed to the appropriate methods
func (p *Peer) HandleInbound() {
	for atomic.LoadInt32(&p.disconnect) == 0 {

		// HMM?
		time.Sleep(50 * time.Millisecond)
		// Wait for a message from the peer
		msgs, err := monkwire.ReadMessages(p.conn)
		if err != nil {
			peerlogger.Debugln(err)
		}
		for _, msg := range msgs {
			peerlogger.DebugDetailf("(%v) => %v\n", p.conn.RemoteAddr(), formatMessage(msg))

			switch msg.Type {
			case monkwire.MsgHandshakeTy:
				// Version message
				p.handleHandshake(msg)

				//if p.caps.IsCap(CapPeerDiscTy) {
				p.QueueMessage(monkwire.NewMessage(monkwire.MsgGetPeersTy, ""))
				//}

			case monkwire.MsgDiscTy:
				p.Stop()
				peerlogger.Infoln("Disconnect peer: ", DiscReason(msg.Data.Get(0).Uint()))
			case monkwire.MsgPingTy:
				// Respond back with pong
				p.QueueMessage(monkwire.NewMessage(monkwire.MsgPongTy, ""))
			case monkwire.MsgPongTy:
				// If we received a pong back from a peer we set the
				// last pong so the peer handler knows this peer is still
				// active.
				p.setLastPong()
				p.setPingTime()
			case monkwire.MsgTxTy:
				// If the message was a transaction queue the transaction
				// in the TxPool where it will undergo validation and
				// processing when a new block is found
				for i := 0; i < msg.Data.Len(); i++ {
					tx := monkchain.NewTransactionFromValue(msg.Data.Get(i))
					p.thelonious.TxPool().QueueTransaction(tx)
				}
			case monkwire.MsgGetPeersTy:
				// Peer asked for list of connected peers
				p.pushPeers()
			case monkwire.MsgPeersTy:
				// Received a list of peers (probably because MsgGetPeersTy was send)
				data := msg.Data
				// Create new list of possible peers for the thelonious to process
				peers := make([]string, data.Len())
				// Parse each possible peer
				for i := 0; i < data.Len(); i++ {
					value := data.Get(i)
					peers[i] = unpackAddr(value.Get(0), value.Get(1).Uint())
				}
				// Connect to the list of peers
				p.thelonious.ProcessPeerList(peers)

			case monkwire.MsgStatusTy:
				// Handle peer's status msg
				p.handleStatus(msg)
			}

			// TMP
			if p.statusKnown {
				switch msg.Type {
				case monkwire.MsgGetTxsTy:
					// Get the current transactions of the pool
					txs := p.thelonious.TxPool().CurrentTransactions()
					// Get the RlpData values from the txs
					txsInterface := make([]interface{}, len(txs))
					for i, tx := range txs {
						txsInterface[i] = tx.RlpData()
					}
					// Broadcast it back to the peer
					p.QueueMessage(monkwire.NewMessage(monkwire.MsgTxTy, txsInterface))

				case monkwire.MsgGetBlockHashesTy:
					if msg.Data.Len() < 2 {
						peerlogger.Debugln("err: argument length invalid ", msg.Data.Len())
					}

					hash := msg.Data.Get(0).Bytes()
					amount := msg.Data.Get(1).Uint()

					hashes := p.thelonious.ChainManager().GetChainHashesFromHash(hash, amount)

					p.QueueMessage(monkwire.NewMessage(monkwire.MsgBlockHashesTy, monkutil.ByteSliceToInterface(hashes)))

				case monkwire.MsgGetBlocksTy:
					// Limit to max 300 blocks
					max := int(math.Min(float64(msg.Data.Len()), 300.0))
					var blocks []interface{}

					for i := 0; i < max; i++ {
						hash := msg.Data.Get(i).Bytes()
						block := p.thelonious.ChainManager().GetBlock(hash)
						if block != nil {
							blocks = append(blocks, block.Value().Raw())
						}
					}

					p.QueueMessage(monkwire.NewMessage(monkwire.MsgBlockTy, blocks))

				case monkwire.MsgBlockHashesTy:
					p.setCatchingUp(true)

					blockPool := p.thelonious.blockPool
					//waiting := p.thelonious.ChainManager().WaitingForCheckpoint()
					// add hashes to pool until found common
					foundCommonHash := false
					it := msg.Data.NewIterator()
					for it.Next() {
						hash := it.Value().Bytes()
						p.lastReceivedHash = hash
						if blockPool.HasCommonHash(hash) {
							foundCommonHash = true
							break
						}
						blockPool.AddHash(hash, p)
					}

					if !foundCommonHash && msg.Data.Len() != 0 {
						p.FetchHashes()
					} else {
						peerlogger.Infof("Found common hash (%x...)\n", p.lastReceivedHash[0:4])
						p.setDoneFetchingHashes(true)
					}

				case monkwire.MsgBlockTy:
					p.setCatchingUp(true)

					blockPool := p.thelonious.blockPool

					it := msg.Data.NewIterator()
					for it.Next() {
						block := monkchain.NewBlockFromRlpValue(it.Value())
						//fmt.Printf("%v %x - %x\n", block.Number, block.Hash()[0:4], block.PrevHash[0:4])

						blockPool.Add(block, p)

						p.setLastBlockReceived()
					}

				case monkwire.MsgGetStateTy:
					data := msg.Data.Get(0)
					bb := p.thelonious.ChainManager().GetBlock(data.Bytes())
					tr := bb.State().Trie
					poollogger.Infoln("root is", tr.Root)
					trIt := tr.NewIterator()
					response := []interface{}{}
					trIt.Each(func(key string, val *monkutil.Value) {
						pair := []interface{}{[]byte(key), val.Bytes()}
						response = append(response, pair)
					})

					p.QueueMessage(monkwire.NewMessage(monkwire.MsgStateTy, response))

				case monkwire.MsgStateTy:
					poollogger.Infoln("Catching up on state!")
					newTrie := monktrie.New(monkutil.Config.Db, "")
					for i := 0; i < msg.Data.Len(); i++ {
						n := msg.Data.Get(i)
						newTrie.Update(string(n.Get(0).Bytes()), string(n.Get(1).Bytes()))
					}
					newTrie.Sync()
					p.thelonious.Reactor().Post("chainReady", nil)
				}

			}
		}
	}

	p.Stop()
}

func (self *Peer) FetchBlocks(hashes [][]byte) {
	if len(hashes) > 0 {
		peerlogger.Debugf("Fetching blocks (%d)\n", len(hashes))

		self.QueueMessage(monkwire.NewMessage(monkwire.MsgGetBlocksTy, monkutil.ByteSliceToInterface(hashes)))
	}
}

func (self *Peer) FetchHashes() {
	self.doneFetchingHashes = false

	blockPool := self.thelonious.blockPool

	// if this peer has higher TD than other peers, fetch hashes
	if self.td.Cmp(self.thelonious.HighestTDPeer()) >= 0 {
		blockPool.td = self.td

		if !blockPool.HasLatestHash() {
			self.QueueMessage(monkwire.NewMessage(monkwire.MsgGetBlockHashesTy, []interface{}{self.lastReceivedHash, uint32(256)}))
		}
	}
}

func (self *Peer) FetchingHashes() bool {
	return !self.doneFetchingHashes
}

// General update method
func (self *Peer) update() {
	serviceTimer := time.NewTicker(100 * time.Millisecond)

out:
	for {
		select {
		case <-serviceTimer.C:
			if self.IsCap("eth") {
				self.mut.Lock()
				var (
					sinceBlock = time.Since(self.lastBlockReceived)
				)
				self.mut.Unlock()

				if sinceBlock > 5*time.Second {
					self.setCatchingUp(false)
				}
			}
		case <-self.quit:
			break out
		}
	}

	serviceTimer.Stop()
}

func (p *Peer) Start() {
	servHost, servPort, _ := net.SplitHostPort(p.conn.RemoteAddr().String())

	// this peers addr is the remote host
	// its port is either the remote port or, if its inbound,
	// will be fixed when we handle handshake
	p.host, p.port = packAddr(servHost, servPort)

	err := p.pushHandshake()
	if err != nil {
		peerlogger.Debugln("Peer can't send outbound version ack", err)

		p.Stop()

		return
	}

	go p.HandleOutbound()
	// Run the inbound handler in a new goroutine
	go p.HandleInbound()
	// Run the general update handler
	go p.update()

	// Wait a few seconds for startup and then ask for an initial ping
	time.Sleep(2 * time.Second)
	p.writeMessage(monkwire.NewMessage(monkwire.MsgPingTy, ""))
	p.setPingStartTime()

}

func (p *Peer) setPingStartTime() {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.pingStartTime = time.Now()
}

func (p *Peer) setLastBlockReceived() {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.lastBlockReceived = time.Now()
}

func (p *Peer) setCatchingUp(b bool) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.catchingUp = b
}

func (p *Peer) setLastPong() {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.lastPong = time.Now().Unix()
}

func (p *Peer) setLastSend() {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.lastSend = time.Now()
}

func (p *Peer) setPingTime() {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.pingTime = time.Since(p.pingStartTime)
}

func (p *Peer) setDoneFetchingHashes(b bool) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.doneFetchingHashes = b
}

func (p *Peer) Stop() {
	if atomic.AddInt32(&p.disconnect, 1) != 1 {
		return
	}

	close(p.quit)
	if atomic.LoadInt32(&p.connected) != 0 {
		p.writeMessage(monkwire.NewMessage(monkwire.MsgDiscTy, ""))
		p.conn.Close()
	}

	// Pre-emptively remove the peer; don't wait for reaping. We already know it's dead if we are here
	p.thelonious.RemovePeer(p)
}

func (p *Peer) peersMessage() *monkwire.Msg {
	outPeers := make([]interface{}, len(p.thelonious.InOutPeers()))
	// Serialise each peer
	for i, peer := range p.thelonious.InOutPeers() {
		// Don't return localhost as valid peer
		if !net.ParseIP(peer.conn.RemoteAddr().String()).IsLoopback() {
			outPeers[i] = peer.RlpData()
		}
	}

	// Return the message to the peer with the known list of connected clients
	return monkwire.NewMessage(monkwire.MsgPeersTy, outPeers)
}

// Pushes the list of outbound peers to the client when requested
func (p *Peer) pushPeers() {
	p.QueueMessage(p.peersMessage())
}

func (self *Peer) pushStatus() {
	msg := monkwire.NewMessage(monkwire.MsgStatusTy, []interface{}{
		uint32(ProtocolVersion),
		uint32(NetVersion),
		self.thelonious.ChainManager().TD,
		self.thelonious.ChainManager().CurrentBlock().Hash(),
		//self.thelonious.ChainManager().Genesis().Hash(),
		self.thelonious.ChainManager().ChainID(),
	})

	self.QueueMessage(msg)
}

func (self *Peer) handleStatus(msg *monkwire.Msg) {
	c := msg.Data

	var (
		protoVersion = c.Get(0).Uint()
		netVersion   = c.Get(1).Uint()
		td           = c.Get(2).BigInt()
		bestHash     = c.Get(3).Bytes()
		chainId      = c.Get(4).Bytes()
	)

	if bytes.Compare(self.thelonious.ChainManager().ChainID(), chainId) != 0 {
		monklogger.Warnf("Invalid chainId %x. Disabling [eth]\n", chainId)
		return
	}

	if netVersion != NetVersion {
		monklogger.Warnf("Invalid network version %d. Disabling [eth]\n", netVersion)
		return
	}

	if protoVersion != ProtocolVersion {
		monklogger.Warnf("Invalid protocol version %d. Disabling [eth]\n", protoVersion)
		return
	}

	self.mut.Lock()
	// Get the td and last hash
	self.td = td
	self.bestHash = bestHash
	self.lastReceivedHash = bestHash

	self.statusKnown = true
	self.mut.Unlock()

	// Compare the total TD with the blockchain TD. If remote is higher
	// fetch hashes from highest TD node.
	if self.td.Cmp(self.thelonious.ChainManager().TD) > 0 {
		self.thelonious.blockPool.AddHash(self.lastReceivedHash, self)
		self.FetchHashes()
	}

	monklogger.Infof("Peer is [eth] capable. (TD = %v ~ %x) %d / %d", self.td, self.bestHash, protoVersion, netVersion)

}

func (p *Peer) pushHandshake() error {
	pubkey := p.thelonious.KeyManager().PublicKey()
	// we need to send our listening port
	port, _ := strconv.Atoi(p.thelonious.Port)
	msg := monkwire.NewMessage(monkwire.MsgHandshakeTy, []interface{}{
		P2PVersion, []byte(p.version), []interface{}{"eth"}, port, pubkey[1:],
	})

	p.QueueMessage(msg)

	return nil
}

func (p *Peer) handleHandshake(msg *monkwire.Msg) {
	c := msg.Data

	var (
		p2pVersion = c.Get(0).Uint()
		clientId   = c.Get(1).Str()
		caps       = c.Get(2)
		port       = c.Get(3).Uint()
		pub        = c.Get(4).Bytes()
	)

	// Check correctness of p2p protocol version
	if p2pVersion != P2PVersion {
		peerlogger.Debugf("Invalid P2P version. Require protocol %d, received %d\n", P2PVersion, p2pVersion)
		p.Stop()
		return
	}

	// Handle the pub key (validation, uniqueness)
	if len(pub) == 0 {
		peerlogger.Warnln("Pubkey required, not supplied in handshake.")
		p.Stop()
		return
	}

	// Self connect detection
	pubkey := p.thelonious.KeyManager().PublicKey()
	if bytes.Compare(pubkey[1:], pub) == 0 {
		p.Stop()

		return
	}

	usedPub := 0
	// This peer is already added to the peerlist so we expect to find a double pubkey at least once
	eachPeer(p.thelonious.Peers(), func(peer *Peer, e *list.Element) {
		if bytes.Compare(pub, peer.pubkey) == 0 {
			usedPub++
		}
	})

	if usedPub > 0 {
		peerlogger.Debugf("Pubkey %x found more then once. Already connected to client.", p.pubkey)
		p.Stop()
		return
	}
	p.pubkey = pub

	// If this is an inbound connection send an ack back
	if p.inbound {
		p.port = uint16(port)
	}

	p.SetVersion(clientId)

	p.versionKnown = true

	p.thelonious.PushPeer(p)
	p.thelonious.reactor.Post("peerList", p.thelonious.Peers())

	p.mut.Lock()
	p.protocolCaps = caps
	p.mut.Unlock()

	capsIt := caps.NewIterator()
	var capsStrs []string
	for capsIt.Next() {
		cap := capsIt.Value().Str()
		switch cap {
		case "eth":
			p.pushStatus()
		}

		capsStrs = append(capsStrs, cap)
	}

	monklogger.Infof("Added peer (%s) %d / %d (%v)\n", p.conn.RemoteAddr(), p.thelonious.Peers().Len(), p.thelonious.MaxPeers, capsStrs)

	peerlogger.Debugln(p)
}

func (self *Peer) IsCap(cap string) bool {
	self.mut.Lock()
	capsIt := self.protocolCaps.NewIterator()
	self.mut.Unlock()

	for capsIt.Next() {
		if capsIt.Value().Str() == cap {
			return true
		}
	}

	return false
}

func (self *Peer) Caps() *monkutil.Value {
	return self.protocolCaps
}

func (p *Peer) String() string {
	var strBoundType string
	if p.inbound {
		strBoundType = "inbound"
	} else {
		strBoundType = "outbound"
	}
	var strConnectType string
	if atomic.LoadInt32(&p.disconnect) == 0 {
		strConnectType = "connected"
	} else {
		strConnectType = "disconnected"
	}

	return fmt.Sprintf("[%s] (%s) %v %s [%s]", strConnectType, strBoundType, p.conn.RemoteAddr(), p.version, p.caps)

}

func (p *Peer) RlpData() []interface{} {
	return []interface{}{p.host, p.port, p.pubkey}
}

func packAddr(address, port string) ([]byte, uint16) {
	addr := strings.Split(address, ".")
	a, _ := strconv.Atoi(addr[0])
	b, _ := strconv.Atoi(addr[1])
	c, _ := strconv.Atoi(addr[2])
	d, _ := strconv.Atoi(addr[3])
	host := []byte{byte(a), byte(b), byte(c), byte(d)}
	prt, _ := strconv.Atoi(port)

	return host, uint16(prt)
}

func unpackAddr(value *monkutil.Value, p uint64) string {
	byts := value.Bytes()
	a := strconv.Itoa(int(byts[0]))
	b := strconv.Itoa(int(byts[1]))
	c := strconv.Itoa(int(byts[2]))
	d := strconv.Itoa(int(byts[3]))
	host := strings.Join([]string{a, b, c, d}, ".")
	port := strconv.Itoa(int(p))
	return net.JoinHostPort(host, port)
}
