package influx

// Measurement names. These are the InfluxDB measurement identifiers
// produced by the exporter.
//
// Naming convention is InfluxDB-conventional snake_case. This is a
// deliberate breaking change from paregupt/nexus_traffic_monitor (NTM)
// in cases where NTM used camelCase or mixed-case names (e.g. NTM's
// "ucCurrQueueDepth" becomes this package's "current_depth").
const (
	// MeasurementDevice carries device-level metadata and gauges.
	MeasurementDevice = "device"
	// MeasurementInterface carries per-interface counters and state.
	MeasurementInterface = "interface"
	// MeasurementQueue carries per-queue traffic-management counters.
	MeasurementQueue = "queue"
	// MeasurementBuffer carries per-ASIC-instance buffer occupancy.
	MeasurementBuffer = "buffer"
	// MeasurementBurst carries discrete burst events with PeakTime
	// timestamps.
	MeasurementBurst = "burst"
)

// Tag keys reused across measurements.
const (
	TagHost           = "host"
	TagLocation       = "location"
	TagVendor         = "vendor"
	TagRole           = "role"
	TagInterface      = "interface"
	TagAdminState     = "admin_state"
	TagOperState      = "oper_state"
	TagOperMode       = "oper_mode"
	TagClassification = "classification"
	TagPeerName       = "peer_name"
	TagPeerType       = "peer_type"
	TagQueueID        = "queue_id"
	TagQueueName      = "queue_name"
	TagInstanceID     = "instance_id"
	TagInstanceName   = "instance_name"
)

// Device measurement field keys.
const (
	FieldModel               = "model"
	FieldOSVersion           = "os_version"
	FieldCPUKernel           = "cpu_kernel"
	FieldCPUUser             = "cpu_user"
	FieldMemTotal            = "mem_total"
	FieldMemUsed             = "mem_used"
	FieldKernelUptimeSeconds = "kernel_uptime_seconds"
	FieldResponseTimeMs      = "response_time_ms"
)

// Interface measurement field keys. The packet-size buckets are
// emitted as separate rx_ and tx_ prefixed fields because the
// histograms are independent.
const (
	FieldRxBytes      = "rx_bytes"
	FieldTxBytes      = "tx_bytes"
	FieldRxUcastPkts  = "rx_ucast_pkts"
	FieldTxUcastPkts  = "tx_ucast_pkts"
	FieldRxMcastPkts  = "rx_mcast_pkts"
	FieldTxMcastPkts  = "tx_mcast_pkts"
	FieldRxBcastPkts  = "rx_bcast_pkts"
	FieldTxBcastPkts  = "tx_bcast_pkts"
	FieldRxCRC        = "rx_crc"
	FieldRxCRCStomped = "rx_crc_stomped"
	FieldRxJumbo      = "rx_jumbo"
	FieldTxJumbo      = "tx_jumbo"
	FieldOperSpeed    = "oper_speed"
	FieldDescription  = "description"
	FieldDownReason   = "down_reason"

	FieldRxPkts64         = "rx_pkts_64"
	FieldRxPkts65to127    = "rx_pkts_65_127"
	FieldRxPkts128to255   = "rx_pkts_128_255"
	FieldRxPkts256to511   = "rx_pkts_256_511"
	FieldRxPkts512to1023  = "rx_pkts_512_1023"
	FieldRxPkts1024to1518 = "rx_pkts_1024_1518"
	FieldRxPkts1519to2047 = "rx_pkts_1519_2047"
	FieldRxPkts2048to4095 = "rx_pkts_2048_4095"
	FieldRxPkts4096to9216 = "rx_pkts_4096_9216"

	FieldTxPkts64         = "tx_pkts_64"
	FieldTxPkts65to127    = "tx_pkts_65_127"
	FieldTxPkts128to255   = "tx_pkts_128_255"
	FieldTxPkts256to511   = "tx_pkts_256_511"
	FieldTxPkts512to1023  = "tx_pkts_512_1023"
	FieldTxPkts1024to1518 = "tx_pkts_1024_1518"
	FieldTxPkts1519to2047 = "tx_pkts_1519_2047"
	FieldTxPkts2048to4095 = "tx_pkts_2048_4095"
	FieldTxPkts4096to9216 = "tx_pkts_4096_9216"
)

// Queue measurement field keys. Some names overlap with interface
// fields by design; the InfluxDB measurement scope keeps them
// distinct downstream.
const (
	FieldQueueTxBytes         = "tx_bytes"
	FieldQueueTxPkts          = "tx_pkts"
	FieldQueueDropBytes       = "drop_bytes"
	FieldQueueDropPkts        = "drop_pkts"
	FieldQueueRandomDropBytes = "random_drop_bytes"
	FieldQueueRandomDropPkts  = "random_drop_pkts"
	FieldQueueCurrentDepth    = "current_depth"
	FieldQueuePeakDepth       = "peak_depth"
	FieldQueueRxPause         = "rx_pause"
	FieldQueueTxPause         = "tx_pause"
	FieldQueueWatchdogEvents  = "watchdog_events"
	FieldQueueECNMarkedPkts   = "ecn_marked_pkts"
)

// Buffer measurement field keys.
const (
	FieldBufferPeakCellDropPG    = "peak_cell_drop_pg"
	FieldBufferPeakCellNoDrop    = "peak_cell_no_drop"
	FieldBufferCurrentCellDropPG = "current_cell_drop_pg"
	FieldBufferCurrentCellNoDrop = "current_cell_no_drop"
)

// Burst measurement field keys.
const (
	FieldBurstStartDepth = "start_depth"
	FieldBurstEndDepth   = "end_depth"
	FieldBurstPeakDepth  = "peak_depth"
	FieldBurstDurationUs = "duration_us"
)
