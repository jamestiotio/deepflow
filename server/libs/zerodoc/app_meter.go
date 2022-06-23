package zerodoc

import (
	"server/libs/ckdb"
	"server/libs/zerodoc/pb"
)

type AppMeter struct {
	AppTriffic
	AppLatency
	AppAnomaly
}

func (m *AppMeter) Reverse() {
	m.AppTriffic.Reverse()
	m.AppLatency.Reverse()
	m.AppAnomaly.Reverse()
}

func (m *AppMeter) ID() uint8 {
	return APP_ID
}

func (m *AppMeter) Name() string {
	return MeterVTAPNames[m.ID()]
}

func (m *AppMeter) VTAPName() string {
	return MeterVTAPNames[m.ID()]
}

func (m *AppMeter) SortKey() uint64 {
	return m.RRTSum
}

func (m *AppMeter) WriteToPB(p *pb.AppMeter) {
	if p.AppTriffic == nil {
		p.AppTriffic = &pb.AppTriffic{}
	}
	m.AppTriffic.WriteToPB(p.AppTriffic)

	if p.AppLatency == nil {
		p.AppLatency = &pb.AppLatency{}
	}
	m.AppLatency.WriteToPB(p.AppLatency)

	if p.AppAnomaly == nil {
		p.AppAnomaly = &pb.AppAnomaly{}
	}
	m.AppAnomaly.WriteToPB(p.AppAnomaly)
}

func (m *AppMeter) ReadFromPB(p *pb.AppMeter) {
	m.AppTriffic.ReadFromPB(p.AppTriffic)
	m.AppLatency.ReadFromPB(p.AppLatency)
	m.AppAnomaly.ReadFromPB(p.AppAnomaly)
}

func (m *AppMeter) ConcurrentMerge(other Meter) {
	if pm, ok := other.(*AppMeter); ok {
		m.AppTriffic.ConcurrentMerge(&pm.AppTriffic)
		m.AppLatency.ConcurrentMerge(&pm.AppLatency)
		m.AppAnomaly.ConcurrentMerge(&pm.AppAnomaly)
	}
}

func (m *AppMeter) SequentialMerge(other Meter) {
	if pm, ok := other.(*AppMeter); ok {
		m.AppTriffic.SequentialMerge(&pm.AppTriffic)
		m.AppLatency.SequentialMerge(&pm.AppLatency)
		m.AppAnomaly.SequentialMerge(&pm.AppAnomaly)
	}
}

func (m *AppMeter) ToKVString() string {
	buffer := make([]byte, MAX_STRING_LENGTH)
	size := m.MarshalTo(buffer)
	return string(buffer[:size])
}

func (m *AppMeter) MarshalTo(b []byte) int {
	offset := 0

	offset += m.AppTriffic.MarshalTo(b[offset:])
	if offset > 0 && b[offset-1] != ',' {
		b[offset] = ','
		offset++
	}

	offset += m.AppLatency.MarshalTo(b[offset:])
	if offset > 0 && b[offset-1] != ',' {
		b[offset] = ','
		offset++
	}
	offset += m.AppAnomaly.MarshalTo(b[offset:])
	if offset > 0 && b[offset-1] != ',' {
		b[offset] = ','
		offset++
	}

	return offset
}

func AppMeterColumns() []*ckdb.Column {
	columns := []*ckdb.Column{}
	columns = append(columns, AppTrifficColumns()...)
	columns = append(columns, AppLatencyColumns()...)
	columns = append(columns, AppAnomalyColumns()...)
	return columns
}

func (m *AppMeter) WriteBlock(block *ckdb.Block) error {
	if err := m.AppTriffic.WriteBlock(block); err != nil {
		return err
	}
	if err := m.AppLatency.WriteBlock(block); err != nil {
		return err
	}
	if err := m.AppAnomaly.WriteBlock(block); err != nil {
		return err
	}
	return nil
}

type AppTriffic struct {
	Request  uint32 `db:"request"`
	Response uint32 `db:"response"`
}

func (_ *AppTriffic) Reverse() {
	// 异常统计量以客户端、服务端为视角，无需Reverse
}

func (t *AppTriffic) WriteToPB(p *pb.AppTriffic) {
	p.Request = t.Request
	p.Response = t.Response
}

func (t *AppTriffic) ReadFromPB(p *pb.AppTriffic) {
	t.Request = p.Request
	t.Response = p.Response
}

func (t *AppTriffic) ConcurrentMerge(other *AppTriffic) {
	t.Request += other.Request
	t.Response += other.Response
}

func (t *AppTriffic) SequentialMerge(other *AppTriffic) {
	t.ConcurrentMerge(other)
}

func (t *AppTriffic) MarshalTo(b []byte) int {
	fields := []string{"request=", "response="}
	values := []uint64{uint64(t.Request), uint64(t.Response)}
	return marshalKeyValues(b, fields, values)
}

const (
	AppTRIFFIC_RRT_MAX = iota
	AppTRIFFIC_RRT_SUM
	AppTRIFFIC_RRT_COUNT
)

// Columns列和WriteBlock的列需要按顺序一一对应
func AppTrifficColumns() []*ckdb.Column {
	columns := []*ckdb.Column{}
	columns = append(columns, ckdb.NewColumn("request", ckdb.UInt32).SetComment("累计请求次数").SetIndex(ckdb.IndexNone))
	columns = append(columns, ckdb.NewColumn("response", ckdb.UInt32).SetComment("累计响应次数").SetIndex(ckdb.IndexNone))
	return columns
}

// WriteBlock和LatencyColumns的列需要按顺序一一对应
func (t *AppTriffic) WriteBlock(block *ckdb.Block) error {
	if err := block.WriteUInt32(t.Request); err != nil {
		return err
	}
	if err := block.WriteUInt32(t.Response); err != nil {
		return err
	}
	return nil
}

type AppLatency struct {
	RRTMax   uint32 `db:"rrt_max"` // us
	RRTSum   uint64 `db:"rrt_sum"` // us
	RRTCount uint32 `db:"rrt_count"`
}

func (_ *AppLatency) Reverse() {
	// 异常统计量以客户端、服务端为视角，无需Reverse
}

func (l *AppLatency) WriteToPB(p *pb.AppLatency) {
	p.RRTMax = l.RRTMax
	p.RRTSum = l.RRTSum
	p.RRTCount = l.RRTCount
}

func (l *AppLatency) ReadFromPB(p *pb.AppLatency) {
	l.RRTMax = p.RRTMax
	l.RRTSum = p.RRTSum
	l.RRTCount = p.RRTCount
}

func (l *AppLatency) ConcurrentMerge(other *AppLatency) {
	if l.RRTMax < other.RRTMax {
		l.RRTMax = other.RRTMax
	}
	l.RRTSum += other.RRTSum
	l.RRTCount += other.RRTCount
}

func (l *AppLatency) SequentialMerge(other *AppLatency) {
	l.ConcurrentMerge(other)
}

func (l *AppLatency) MarshalTo(b []byte) int {
	fields := []string{"rrt_sum=", "rrt_count=", "rrt_max="}
	values := []uint64{l.RRTSum, uint64(l.RRTCount), uint64(l.RRTMax)}
	return marshalKeyValues(b, fields, values)
}

const (
	APPLATENCY_RRT_MAX = iota
	APPLATENCY_RRT_SUM
	APPLATENCY_RRT_COUNT
)

// Columns列和WriteBlock的列需要按顺序一一对应
func AppLatencyColumns() []*ckdb.Column {
	columns := []*ckdb.Column{}
	columns = append(columns, ckdb.NewColumn("rrt_max", ckdb.UInt32).SetComment("所有请求响应时延最大值(us)").SetIndex(ckdb.IndexNone))
	columns = append(columns, ckdb.NewColumn("rrt_sum", ckdb.Float64).SetComment("累计所有请求响应时延(us)"))
	columns = append(columns, ckdb.NewColumn("rrt_count", ckdb.UInt64).SetComment("请求响应时延计算次数"))
	return columns
}

// WriteBlock和LatencyColumns的列需要按顺序一一对应
func (l *AppLatency) WriteBlock(block *ckdb.Block) error {
	if err := block.WriteUInt32(l.RRTMax); err != nil {
		return err
	}
	if err := block.WriteFloat64(float64(l.RRTSum)); err != nil {
		return err
	}
	if err := block.WriteUInt64(uint64(l.RRTCount)); err != nil {
		return err
	}
	return nil
}

type AppAnomaly struct {
	ClientError uint32 `db:"client_error"`
	ServerError uint32 `db:"server_error"`
	Timeout     uint32 `db:"timeout"`
}

func (_ *AppAnomaly) Reverse() {
	// 异常统计量以客户端、服务端为视角，无需Reverse
}

func (a *AppAnomaly) WriteToPB(p *pb.AppAnomaly) {
	p.ClientError = a.ClientError
	p.ServerError = a.ServerError
	p.Timeout = a.Timeout
}

func (a *AppAnomaly) ReadFromPB(p *pb.AppAnomaly) {
	a.ClientError = p.ClientError
	a.ServerError = p.ServerError
	a.Timeout = p.Timeout
}

func (a *AppAnomaly) ConcurrentMerge(other *AppAnomaly) {
	a.ClientError += other.ClientError
	a.ServerError += other.ServerError
	a.Timeout += other.Timeout
}

func (a *AppAnomaly) SequentialMerge(other *AppAnomaly) {
	a.ConcurrentMerge(other)
}

func (a *AppAnomaly) MarshalTo(b []byte) int {
	fields := []string{
		"client_error=", "server_error=", "timeout=", "error=",
	}
	values := []uint64{
		uint64(a.ClientError), uint64(a.ServerError), uint64(a.Timeout), uint64(a.ClientError + a.ServerError),
	}
	return marshalKeyValues(b, fields, values)
}

const (
	APPANOMALY_CLIENT_ERROR = iota
	APPANOMALY_SERVER_ERROR
	APPANOMALY_TIMEOUT
	APPANOMALY_ERROR
)

// Columns列和WriteBlock的列需要按顺序一一对应
func AppAnomalyColumns() []*ckdb.Column {
	columns := ckdb.NewColumnsWithComment(
		[][2]string{
			APPANOMALY_CLIENT_ERROR: {"client_error", "客户端异常次数"},
			APPANOMALY_SERVER_ERROR: {"server_error", "服务端异常次数"},
			APPANOMALY_TIMEOUT:      {"timeout", "请求超时次数"},
			APPANOMALY_ERROR:        {"error", "异常次数"},
		}, ckdb.UInt64)
	for _, v := range columns {
		v.SetIndex(ckdb.IndexNone)
	}
	return columns
}

// WriteBlock的列和AnomalyColumns需要按顺序一一对应
func (a *AppAnomaly) WriteBlock(block *ckdb.Block) error {
	values := []uint64{
		APPANOMALY_CLIENT_ERROR: uint64(a.ClientError),
		APPANOMALY_SERVER_ERROR: uint64(a.ServerError),
		APPANOMALY_TIMEOUT:      uint64(a.Timeout),
		APPANOMALY_ERROR:        uint64(a.ClientError + a.ServerError),
	}
	for _, v := range values {
		if err := block.WriteUInt64(v); err != nil {
			return err
		}
	}
	return nil
}
