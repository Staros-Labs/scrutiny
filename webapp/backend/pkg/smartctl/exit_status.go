// Package smartctl holds shared knowledge about the smartctl command-line tool
// that both the collector and the web server need to agree on.
package smartctl

// smartctl reports its result through a bitmask exit status. The bits below are
// defined by the smartctl manpage, "RETURN VALUES":
// http://www.linuxguide.it/command_line/linux-manpage/do.php?file=smartctl#sect7
const (
	ExitCommandLineError      = 0x01 // command line did not parse
	ExitDeviceOpenFailed      = 0x02 // device open failed, or device did not return an IDENTIFY DEVICE structure
	ExitChecksumError         = 0x04 // some SMART command failed, or there was a checksum error in a SMART data structure
	ExitDiskFailing           = 0x08 // SMART status check returned "DISK FAILING"
	ExitPrefailBelowThreshold = 0x10 // found prefail attributes at or below threshold
	ExitUsageBelowThreshold   = 0x20 // some attributes have been at or below threshold at some time in the past
	ExitErrorLogHasRecords    = 0x40 // the device error log contains records of errors
	ExitSelfTestLogHasErrors  = 0x80 // the device self-test log contains records of errors
)

// FatalMask is the set of exit-status bits that mean the JSON payload smartctl
// printed cannot be trusted: the command was never understood (0x01), or the
// device was never opened and so no IDENTIFY structure was read (0x02).
//
// Every other bit describes a condition of the *disk*, not of the output.
// smartctl still emits complete, well-formed JSON alongside them. Treating any
// of those as fatal throws away usable data - see issue 663, where a QNAP TR-004
// enclosure reports "Read Device Statistics page 0x00 failed" and exits 4 while
// returning a full identity and attribute set.
const FatalMask = ExitCommandLineError | ExitDeviceOpenFailed

// IsFatal reports whether a smartctl exit status means the accompanying output
// must be discarded. This single rule is applied at every point that inspects an
// exit status: device detection, metrics collection, and server-side upload
// validation.
func IsFatal(exitStatus int) bool {
	return exitStatus&FatalMask != 0
}
