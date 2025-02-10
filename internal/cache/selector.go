package cache

import (
	"hash"
	"net"
	"strings"
	"sync"

	"github.com/dhaifley/apigo/internal/errors"
	"github.com/google/gomemcache/memcache"
	"github.com/google/uuid"
)

var _ memcache.ServerSelector = ServerList{}

// ServerList is a simple ServerSelector. Its zero value is usable.
type ServerList struct {
	addrs []net.Addr
}

// staticAddr caches the Network() and String() values from any net.Addr.
type staticAddr struct {
	ntw, str string
}

func newStaticAddr(a net.Addr) net.Addr {
	return &staticAddr{
		ntw: a.Network(),
		str: a.String(),
	}
}

func (s *staticAddr) Network() string { return s.ntw }
func (s *staticAddr) String() string  { return s.str }

// NewServerList creates a new ServerList selector.
func NewServerList(servers ...string) (*ServerList, error) {
	ss := &ServerList{}

	nAddr := make([]net.Addr, len(servers))

	for i, server := range servers {
		if strings.Contains(server, "/") {
			addr, err := net.ResolveUnixAddr("unix", server)
			if err != nil {
				return nil, err
			}

			nAddr[i] = newStaticAddr(addr)
		} else {
			tcpAddr, err := net.ResolveTCPAddr("tcp", server)
			if err != nil {
				return nil, err
			}

			nAddr[i] = newStaticAddr(tcpAddr)
		}
	}

	ss.addrs = nAddr

	return ss, nil
}

// Each iterates over each server calling the given function.
func (ss ServerList) Each(f func(net.Addr) error) error {
	for _, a := range ss.addrs {
		if err := f(a); nil != err {
			return err
		}
	}

	return nil
}

// keyBufPool returns []byte buffers to avoid allocations.
var keyBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 256)

		return &b
	},
}

// PickServer selects a server based on the cache key.
// It uses a selection mechanism that mirrors what libmemcached does
// using the MEMCACHED_DISTRIBUTION_MODULA behavior and the
// MEMCACHED_HASH_DEFAULT hash algorithm.
func (ss ServerList) PickServer(key string) (net.Addr, error) {
	if len(ss.addrs) == 0 {
		return nil, memcache.ErrNoServers
	}

	if len(ss.addrs) == 1 {
		return ss.addrs[0], nil
	}

	buf, ok := keyBufPool.Get().(*[]byte)
	if !ok {
		return nil, errors.New(errors.ErrCache,
			"invalid buffer from pool")
	}

	n := copy(*buf, key)
	jh := newJenkinsHash()
	jh.Write((*buf)[:n])
	keyBufPool.Put(buf)

	return ss.addrs[jh.Sum32()%uint32(len(ss.addrs))], nil
}

// PickAnyServer selects any active server randomly.
func (ss ServerList) PickAnyServer() (net.Addr, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCache,
			"unable to generate uuid")
	}

	return ss.PickServer(u.String())
}

// jenkinsHash values are used to implement the same algorithm used by
// libmemcached as MEMCACHED_HASH_DEFAULT.
type jenkinsHash uint32

const (
	blockSize = 1
	hashSize  = 4
)

func newJenkinsHash() hash.Hash32 {
	var j jenkinsHash

	return &j
}

func (j *jenkinsHash) Write(key []byte) (int, error) {
	hash := *j

	for _, b := range key {
		hash += jenkinsHash(b)
		hash += (hash << 10)
		hash ^= (hash >> 6)
	}

	hash += (hash << 3)
	hash ^= (hash >> 11)
	hash += (hash << 15)

	*j = hash

	return len(key), nil
}

func (j *jenkinsHash) Reset() {
	*j = 0
}

func (j *jenkinsHash) Size() int {
	return hashSize
}

func (j *jenkinsHash) BlockSize() int {
	return blockSize
}

func (j *jenkinsHash) Sum32() uint32 {
	return uint32(*j)
}

func (j *jenkinsHash) Sum(in []byte) []byte {
	v := j.Sum32()

	return append(in, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
