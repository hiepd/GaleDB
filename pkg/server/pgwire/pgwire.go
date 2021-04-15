package pgwire

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/hiepd/galedb/pkg/entity"

	"github.com/hiepd/galedb/pkg/index"

	"github.com/hiepd/galedb/pkg/sql/parser"
	"github.com/hiepd/galedb/pkg/sql/planner"
	"github.com/hiepd/galedb/pkg/storage"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Listener  net.Listener
	Parser    *parser.Parser
	DefaultDb *storage.Database
	closeCh   chan struct{}
	doneCh    chan struct{}
}

func NewServer(port int, host string, p *parser.Parser, defaultDb *storage.Database) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	return &Server{
		Listener:  l,
		Parser:    p,
		DefaultDb: defaultDb,
		closeCh:   make(chan struct{}, 1),
		doneCh:    make(chan struct{}, 1),
	}, nil
}

func (s *Server) Start() error {
	wg := &sync.WaitGroup{}
	for {
		select {
		case sig := <-s.closeCh:
			s.closeCh <- sig
			wg.Wait()
			close(s.closeCh)
			s.doneCh <- struct{}{}
			return nil
		default:
			// Wait for a connection.
			c, err := s.Listener.Accept()
			// TODO: Close remaining conn
			if err != nil {
				continue
			}
			wg.Add(1)

			sc := &sessionConn{
				netConn: c,
				parser:  s.Parser,
			}

			// Handle the connection in a new goroutine.
			// The loop then returns to accepting, so that
			// multiple connections may be served concurrently.
			go sc.serveConn(s.DefaultDb, s.closeCh, wg)
		}
	}
}

func (s *Server) Close() error {
	if err := s.Listener.Close(); err != nil {
		return err
	}
	s.closeCh <- struct{}{}
	<-s.doneCh
	close(s.doneCh)
	return nil
}

type sessionConn struct {
	netConn net.Conn
	parser  *parser.Parser
}

func (sc *sessionConn) serveConn(db *storage.Database, closeCh chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		sc.netConn.Close()
		wg.Done()
	}()

	// ctx := context.Background()

	// get startup message
	reader := bufio.NewReader(sc.netConn)
	go func() {
		for {
			select {
			case sig := <-closeCh:
				sc.netConn.Close()
				closeCh <- sig
			}
		}
	}()
	// get length
	lenBytes, err := readBytes(reader, 4)
	if err != nil {
		logrus.Error(err)
		return
	}
	length := binary.BigEndian.Uint32(lenBytes)
	// get protocol
	protocolBytes, err := readBytes(reader, 4)
	if err != nil {
		logrus.Error(err)
		return
	}
	protocol := binary.BigEndian.Uint32(protocolBytes)
	// get payload
	payload, err := readBytes(reader, int(length-8))

	startupMsg := &startupMessage{
		length:   length,
		protocol: protocol,
		payload:  payload,
	}

	logrus.WithField("source", "pgwire").Info("---RECEIVING STARTUP MSG---")
	logrus.WithField("source", "pgwire").Info(startupMsg.string())
	if err := authOk.writeConn(sc.netConn); err != nil {
		logrus.Error(err)
		return
	}
	if err := readyForQuery.writeConn(sc.netConn); err != nil {
		logrus.Error(err)
		return
	}

	for {
		// get tag
		tag, err := reader.ReadByte()
		if err != nil {
			logrus.Error(err)
			break
		}
		// get length
		lenBytes, err := readBytes(reader, 4)
		if err != nil {
			logrus.Error(err)
			break
		}
		length := binary.BigEndian.Uint32(lenBytes)
		// get payload
		payload, err := readBytes(reader, int(length)-4)

		if err != nil {
			logrus.Error(err)
			break
		}
		req := &message{
			tag:     tag,
			payload: payload,
		}
		logrus.WithField("source", "pgwire").Info("---RECEIVING MSG---")
		logrus.WithField("source", "pgwire").Info(req.string())
		cc := &commandComplete{}
		if err := sc.handle(db, req); err != nil {
			logrus.WithError(err).Error("failed to handle query")
			cc.value = err.Error()
		}
		if err := cc.message().writeConn(sc.netConn); err != nil {
			logrus.WithError(err).Error("failed to send command complete")
			break

		}
		if err := readyForQuery.writeConn(sc.netConn); err != nil {
			logrus.WithError(err).Error("failed to send ready for query")
			break
		}
	}
}

func (sc *sessionConn) handle(db *storage.Database, msg *message) error {
	// ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(30)*time.Second))
	parsed, err := sc.parser.Parse(string(msg.payload))
	if err != nil {
		return err
	}
	logrus.Debugf("parsed tree:\n%s", parsed.String())
	pl := planner.New(db)
	plan, err := pl.Prepare(parsed)
	if err != nil {
		return err
	}
	logrus.WithField("plan", plan).Debugf("query plan")
	// execute plan
	iter := plan.Iter()
	cols := plan.Columns()
	fields := make([]*field, len(cols))
	for i, col := range cols {
		fields[i] = &field{
			name: col + "\x00",
		}
	}
	rd := &rowDescription{
		fields: fields,
	}
	if err := rd.message().writeConn(sc.netConn); err != nil {
		return err
	}
	for {
		row, err := iter.Next()
		if err == index.EndOfIterator {
			break
		} else if err != nil {
			logrus.WithError(err).Error("failed to iterate results")
			break
		}
		dr := convertRowToDataRow(&row)
		if err := dr.message().writeConn(sc.netConn); err != nil {
			return err
		}
	}
	return nil
}

type message struct {
	tag     byte
	payload []byte
}

type startupMessage struct {
	length   uint32
	protocol uint32
	payload  []byte
}

var (
	authOk = &message{
		tag:     'R',
		payload: int32ToBytes(0),
	}

	readyForQuery = &message{
		tag:     'Z',
		payload: []byte{'I'},
	}
)

func (msg *message) string() string {
	return fmt.Sprintf("Tag: %c\nPayload: %s", msg.tag, string(msg.payload))
}

func (msg *message) bytes() []byte {
	size := len(msg.payload) + 5
	res := make([]byte, size)
	res[0] = msg.tag
	// write length
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(size)-1)
	for i := 1; i <= 4; i++ {
		res[i] = sizeBytes[i-1]
	}
	// write payload
	for i := 5; i < size; i++ {
		res[i] = msg.payload[i-5]
	}
	return res
}

func (msg *message) writeConn(c net.Conn) error {
	fmt.Println("---WRITING MSG---")
	fmt.Println(msg.bytes())
	_, err := c.Write(msg.bytes())
	return err
}

func (msg *startupMessage) string() string {
	return fmt.Sprintf("Length: %d\nProtocol: %d\nPayload: %s", msg.length, msg.protocol, string(msg.payload))
}

func readBytes(reader *bufio.Reader, n int) ([]byte, error) {
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		res[i] = b
	}
	return res, nil
}

func int32ToBytes(n int32) []byte {
	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, uint32(n))
	return res
}

func int16ToBytes(n int16) []byte {
	res := make([]byte, 2)
	binary.BigEndian.PutUint16(res, uint16(n))
	return res
}

type rowDescription struct {
	fields []*field
}

type field struct {
	name     string
	tableOid int32
	colNo    int16
	typeOid  int32
	typeLen  int16
	typeMod  int32
	format   int16
}

func (rd *rowDescription) message() *message {
	res := make([]byte, 0)
	res = append(res, int16ToBytes(int16(len(rd.fields)))...)
	for _, fd := range rd.fields {
		res = append(res, fd.bytes()...)
	}
	return &message{
		tag:     'T',
		payload: res,
	}
}

func (fd *field) bytes() []byte {
	res := make([]byte, 0)
	res = append(res, []byte(fd.name)...)
	res = append(res, int32ToBytes(fd.tableOid)...)
	res = append(res, int16ToBytes(fd.colNo)...)
	res = append(res, int32ToBytes(fd.typeOid)...)
	res = append(res, int16ToBytes(fd.typeLen)...)
	res = append(res, int32ToBytes(fd.typeMod)...)
	res = append(res, int16ToBytes(fd.format)...)
	return res
}

type dataRow struct {
	colNo int16
	cols  []col
}

type col struct {
	dataLen int32
	data    []byte
}

func (dr *dataRow) message() *message {
	res := make([]byte, 0)
	res = append(res, int16ToBytes(dr.colNo)...)
	for _, c := range dr.cols {
		res = append(res, c.bytes()...)
	}
	return &message{
		tag:     'D',
		payload: res,
	}
}

func (c *col) bytes() []byte {
	res := make([]byte, 0)
	res = append(res, int32ToBytes(int32(len(c.data)))...)
	res = append(res, c.data...)
	return res
}

func convertRowToDataRow(row *entity.Row) *dataRow {
	cols := make([]col, len(row.Values))
	for i, val := range row.Values {
		res := ""
		switch v := val.(type) {
		case int, int16, int32, int64:
			res = fmt.Sprintf("%d", v)
		case string:
			res = v
		default:
		}
		cols[i] = col{
			dataLen: int32(len(res)),
			data:    []byte(res),
		}
	}
	return &dataRow{
		colNo: int16(len(row.Values)),
		cols:  cols,
	}
}

type commandComplete struct {
	value string
}

func (cc *commandComplete) message() *message {
	return &message{
		tag:     'C',
		payload: []byte(cc.value + "\x00"),
	}
}
