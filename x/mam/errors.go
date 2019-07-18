package mam

import "github.com/pkg/errors"

const (
	rcSeverityShift    = 6
	rcSeverityMinor    = 0x0 << rcSeverityShift
	rcSeverityModerate = 0x1 << rcSeverityShift
	rcSeverityMajor    = 0x2 << rcSeverityShift
	rcSeverityFatal    = 0x3 << rcSeverityShift
)

const (
	rcModuleShift             = 8
	rcModuleGeneral           = 0x01 << rcModuleShift
	rcModuleSQLite3           = 0x02 << rcModuleShift
	rcModuleNeighbor          = 0x03 << rcModuleShift
	rcModuleCClient           = 0x04 << rcModuleShift
	rcModuleConsensus         = 0x05 << rcModuleShift
	rcModuleCW                = 0x06 << rcModuleShift
	rcModuleExitProbabilities = 0x07 << rcModuleShift
	rcModuleSnapshot          = 0x08 << rcModuleShift
	rcModuleLedgerValidator   = 0x09 << rcModuleShift
	rcModuleTipSelector       = 0x0A << rcModuleShift
	rcModuleTangle            = 0x0B << rcModuleShift
	rcModuleUtils             = 0x0C << rcModuleShift
	rcModuleProcessor         = 0x0D << rcModuleShift
	rcModuleConf              = 0x0E << rcModuleShift
	rcModuleAPI               = 0x0F << rcModuleShift
	rcModuleNode              = 0x10 << rcModuleShift
	rcModuleMAM               = 0x11 << rcModuleShift
	rcModuleHelpers           = 0x12 << rcModuleShift
	rcModuleCrypto            = 0x13 << rcModuleShift
	rcModuleCommon            = 0x14 << rcModuleShift
	rcModuleHandle            = 0x15 << rcModuleShift
)

const (
	rcOK = 0
	// unknown error
	rcError = 0xFFFF
)

var ErrEntangledUnknown = errors.New("unknown error")

// general module
const (
	rcNullParam    = 0x01 | rcModuleGeneral | rcSeverityMajor
	rcInvalidParam = 0x02 | rcModuleGeneral | rcSeverityMajor
	rcOOM          = 0x03 | rcModuleGeneral | rcSeverityMajor
	rcStillRunning = 0x04 | rcModuleGeneral | rcSeverityModerate
)

var (
	ErrNullParam    = errors.New("null parameter")
	ErrInvalidParam = errors.New("invalid parameter")
	ErrOutOfMemory  = errors.New("out of memory")
	ErrStillRunning = errors.New("still running")
)

// SQLite 3
const (
	rcSQLite3FailedOpenDb            = 0x01 | rcModuleSQLite3 | rcSeverityFatal
	rcSQLite3FailedInsertDb          = 0x02 | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3NoPathForDbSpecified    = 0x03 | rcModuleSQLite3 | rcSeverityFatal
	rcSQLite3FailedNotImplemented    = 0x04 | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3FailedBinding           = 0x05 | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3FailedPreparedStatement = 0x06 | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3FailedFinalize          = 0x07 | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3FailedStep              = 0x08 | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3FailedBegin             = 0x09 | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3FailedEnd               = 0x0A | rcModuleSQLite3 | rcSeverityMajor
	rcSQLite3FailedRollback          = 0x0B | rcModuleSQLite3 | rcSeverityFatal
	rcSQLite3FailedConfig            = 0x0C | rcModuleSQLite3 | rcSeverityFatal
	rcSQLite3FailedInitialize        = 0x0D | rcModuleSQLite3 | rcSeverityFatal
	rcSQLite3FailedShutdown          = 0x0E | rcModuleSQLite3 | rcSeverityFatal
)

var (
	ErrSQLite3FailedOpenDb            = errors.New("failed to open SQLite3 db")
	ErrSQLite3FailedInsertDB          = errors.New("failed insert op. into SQLite3 db")
	ErrSQLite3NoPathForDbSpecified    = errors.New("no db path specified for SQLite 3 db")
	ErrSQLite3FailedNotImplemented    = errors.New("SQLite3 op. failed - not implemented")
	ErrSQLite3FailedBinding           = errors.New("failed SQLite3 binding")
	ErrSQLite3FailedPreparedStatement = errors.New("failed to create SQLite3 prepared statement")
	ErrSQLite3FailedFinalize          = errors.New("SQLite3 finalize op. failed")
	ErrSQlite3FailedStep              = errors.New("SQLite 3 step op. failed")
	ErrSQLite3FailedBegin             = errors.New("SQLite3 begin op. failed")
	ErrSQLite3FailedEnd               = errors.New("SQLite3 end op. failed")
	ErrSQLite3FailedRollback          = errors.New("SQLite3 rollback op. failed")
	ErrSQLite3FailedConfig            = errors.New("SQLite3 config op. failed")
	ErrSQLite3FailedInitialize        = errors.New("SQLite3 initialisation failed")
	ErrSQLite3FailedShutdown          = errors.New("SQLite3 shutdown op. failed")
)

// neighbor
const (
	rcNeighborFailedUriParsing      = 0x01 | rcModuleNeighbor | rcSeverityMajor
	rcNeighborInvalidProtocol       = 0x02 | rcModuleNeighbor | rcSeverityMajor
	rcNeighborInvalidHost           = 0x03 | rcModuleNeighbor | rcSeverityMajor
	rcNeighborFailedSend            = 0x04 | rcModuleNeighbor | rcSeverityModerate
	rcNeighborFailedEndpointInit    = 0x05 | rcModuleNeighbor | rcSeverityFatal
	rcNeighborFailedEndpointDestroy = 0x06 | rcModuleNeighbor | rcSeverityFatal
	rcNeighborAlreadyPaired         = 0x07 | rcModuleNeighbor | rcSeverityModerate
	rcNeighborNotPaired             = 0x08 | rcModuleNeighbor | rcSeverityModerate
)

var (
	ErrNeighborFailedURIParsing      = errors.New("failed to parse neighbor URI")
	ErrNeighborInvalidProtocol       = errors.New("invalid protocol given for neighbor URI")
	ErrNeighborInvalidHost           = errors.New("invalid host for neighbor URI")
	ErrNeighborFailedSend            = errors.New("send to neighbor failed")
	ErrNeighborFailedEndpointInit    = errors.New("failed to init. a new endpoint to the given neighbor")
	ErrNeighborFailedEndpointDestroy = errors.New("failed to cleanly destroy the neighbor endpoint")
	ErrNeighborAlreadyPaired         = errors.New("neighbor already paired")
	ErrNeighborNotPaired             = errors.New("neighbor not paired")
)

// C client
const (
	rcCClientJsonCreate          = 0x01 | rcModuleCClient | rcSeverityFatal
	rcCClientJsonParse           = 0x02 | rcModuleCClient | rcSeverityMajor
	rcCClientHttp                = 0x04 | rcModuleCClient | rcSeverityMajor
	rcCClientHttpReq             = 0x05 | rcModuleCClient | rcSeverityMajor
	rcCClientResError            = 0x07 | rcModuleCClient | rcSeverityModerate
	rcCClientJsonKey             = 0x08 | rcModuleCClient | rcSeverityMinor
	rcCClientFlexTrits           = 0x09 | rcModuleCClient | rcSeverityModerate
	rcCClientNullPtr             = 0x0A | rcModuleCClient | rcSeverityMajor
	rcCClientUnimplemented       = 0x0B | rcModuleCClient | rcSeverityMajor
	rcCClientInvalidSecurity     = 0x0C | rcModuleCClient | rcSeverityMinor
	rcCClientTxDeserializeFailed = 0x0E | rcModuleCClient | rcSeverityModerate
	rcCClientInsufficientBalance = 0x0F | rcModuleCClient | rcSeverityMinor
	rcCClientPowFailed           = 0x10 | rcModuleCClient | rcSeverityMinor
	rcCClientInvalidTransfer     = 0x11 | rcModuleCClient | rcSeverityModerate
	rcCClientInvalidTailHash     = 0x12 | rcModuleCClient | rcSeverityMajor
	rcCClientInvalidBundle       = 0x13 | rcModuleCClient | rcSeverityMinor
	rcCClientCheckBalance        = 0x14 | rcModuleCClient | rcSeverityMinor
	rcCClientNotPromotable       = 0x15 | rcModuleCClient | rcSeverityMinor
)

var (
	ErrCClientJSONCreate          = errors.New("error creating JSON")
	ErrCClientJSONParsing         = errors.New("error parsing JSON")
	ErrCClientHTTP                = errors.New("http error")
	ErrCClientHTTPReq             = errors.New("http request error")
	ErrCClientHTTPRes             = errors.New("http response error")
	ErrCClientJSONKey             = errors.New("error reading JSON key")
	ErrCClientFlexTrits           = errors.New("flex trits error")
	ErrCClientNullPointer         = errors.New("null pointer")
	ErrCClientUnimplemented       = errors.New("unimplemented")
	ErrCClientInvalidSecurity     = errors.New("invalid security")
	ErrCClientTxDeserializeFailed = errors.New("failed to deserialize tx")
	ErrCClientInsufficientBalance = errors.New("insufficient balance")
	ErrCClientPowFailed           = errors.New("proof-of-work failed")
	ErrCClientInvalidTransfer     = errors.New("invalid transfer")
	ErrCClientInvalidTailHash     = errors.New("invalid tail hash")
	ErrCClientInvalidBundle       = errors.New("invalid bundle")
	ErrCClientCheckBalance        = errors.New("error checking balance")
	ErrCClientNotPromotable       = errors.New("the given tx is not promotable")
)

// consensus
const (
	rcConsensusNotImplemented = 0x01 | rcModuleConsensus | rcSeverityMajor
)

var ErrConsensusNotImplemented = errors.New("consensus not implemented")

// consensus CW module
const (
	rcCWFailedInDfsFromDb = 0x01 | rcModuleCW | rcSeverityMajor
	rcCWFailedInLightDFS  = 0x02 | rcModuleCW | rcSeverityMajor
)

var (
	ErrCWFailedInDFSFromDB = errors.New("CW calculation failed in DFS from db")
	ErrCWFailedInLightDFS  = errors.New("CW calculation failed in light DFS")
)

// consensus exit probabilities
const (
	rcExitProbabilitiesInvalidEntrypoint = 0x01 | rcModuleExitProbabilities | rcSeverityMajor
	rcExitProbabilitiesMissingRating     = 0x02 | rcModuleExitProbabilities | rcSeverityModerate
	rcExitProbabilitiesNotImplemented    = 0x03 | rcModuleExitProbabilities | rcSeverityMajor
)

var (
	ErrExitProbabilitiesInvalidEntrypoint = errors.New("invalid entrypoint for exit probabilities")
	ErrExitProbabilitiesMissingRating     = errors.New("missing rating for exit probabilities")
	ErrExitProbabilitiesNotImplemented    = errors.New("exit probabilities not implemented")
)

// snapshot
const (
	rcSnapshotFileNotFound                  = 0x01 | rcModuleSnapshot | rcSeverityFatal
	rcSnapshotInvalidFile                   = 0x02 | rcModuleSnapshot | rcSeverityFatal
	rcSnapshotInvalidSupply                 = 0x03 | rcModuleSnapshot | rcSeverityFatal
	rcSnapshotInconsistentSnapshot          = 0x04 | rcModuleSnapshot | rcSeverityFatal
	rcSnapshotInconsistentPatch             = 0x05 | rcModuleSnapshot | rcSeverityMajor
	rcSnapshotBalanceNotFound               = 0x06 | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotInvalidSignature              = 0x07 | rcModuleSnapshot | rcSeverityFatal
	rcSnapshotFailedJsonParsing             = 0x08 | rcModuleSnapshot | rcSeverityFatal
	rcSnapshotServiceNotEnoughDepth         = 0x09 | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotServiceMilestoneNotLoaded     = 0x0A | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotServiceMilestoneNotSolid      = 0x0B | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotServiceMilestoneTooOld        = 0x0C | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotMetadataFailedDeserializing   = 0x0D | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotMetadataFailedSerializing     = 0x0E | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotStateDeltaFailedDeserializing = 0x0F | rcModuleSnapshot | rcSeverityModerate
	rcSnapshotMissingMilestoneTransaction   = 0x10 | rcModuleSnapshot | rcSeverityModerate
)

var (
	ErrSnapshotFileNotFound                  = errors.New("snapshot file not found")
	ErrSnapshotInvalidFile                   = errors.New("invalid snapshot file")
	ErrSnapshotInvalidSupply                 = errors.New("invalid supply in snapshot file")
	ErrSnapshotInconsistentSnapshot          = errors.New("inconsistent snapshot")
	ErrSnapshotInconsistentPatch             = errors.New("inconsistent patch")
	ErrSnapshotBalanceNotFound               = errors.New("balance not found")
	ErrSnapshotInvalidSignature              = errors.New("invalid signature")
	ErrSnapshotFailedJSONParsing             = errors.New("failed to parse JSOn")
	ErrSnapshotServiceNotEnoughDepth         = errors.New("service not enough depth")
	ErrSnapshotServiceMilestoneNotLoaded     = errors.New("milestone not loaded")
	ErrSnapshotServiceMilestoneNotSolid      = errors.New("milestone not solid")
	ErrSnapshotServiceMilestoneTooOld        = errors.New("milestone too old")
	ErrSnapshotMetadataFailedDeserializing   = errors.New("failed to deserialize metadata")
	ErrSnapshotMetadataFailedSerializing     = errors.New("failedto serialize metadata")
	ErrSnapshotStateDeltaFailedDeserializing = errors.New("failed to deserialize snapshot state delta")
	ErrSnapshotMissingMilestoneTransaction   = errors.New("missing milestone tx in snapshot")
)

// ledger validator
const (
	rcLedgerValidatorInvalidTransaction    = 0x01 | rcModuleLedgerValidator | rcSeverityMajor
	rcLedgerValidatorCouldNotLoadMilestone = 0x02 | rcModuleLedgerValidator | rcSeverityMajor
	rcLedgerValidatorInconsistentDelta     = 0x03 | rcModuleLedgerValidator | rcSeverityFatal
	rcLedgerValidatorTransactionNotSolid   = 0x04 | rcModuleLedgerValidator | rcSeverityFatal
)

var (
	ErrLedgerValidatorInvalidTransaction    = errors.New("invalid transaction")
	ErrLedgerValidatorCouldNotLoadMilestone = errors.New("couldn't load milestone")
	ErrLedgerValidatorInconsistentDelta     = errors.New("inconsistent delta")
	ErrLedgerValidatorTransactionNotSolid   = errors.New("transaction not solid")
)

// tip selector
const (
	rcTipSelectorTipsNotConsistent = 0x01 | rcModuleTipSelector | rcSeverityModerate
	rcTipSelectorReferenceTooOld   = 0x02 | rcModuleTipSelector | rcSeverityModerate
)

var (
	ErrTipSelectorTipsNotConsistent = errors.New("tips not consistent")
	ErrTipSelectorReferenceTooOld   = errors.New("reference too old")
)

// tangle
const (
	rcTangleTailNotFound = 0x01 | rcModuleTangle | rcSeverityModerate
	rcTangleNotATail     = 0x02 | rcModuleTangle | rcSeverityModerate
)

var (
	ErrTangleTailNotFound = errors.New("tail not found")
	ErrTangleNotATail     = errors.New("not a tail tx")
)

// utils
const (
	rcUtilsFailedRemoveFile        = 0x01 | rcModuleUtils | rcSeverityMajor
	rcUtilsFailedToCopyFile        = 0x02 | rcModuleUtils | rcSeverityMajor
	rcUtilsFailedToOpenFile        = 0x03 | rcModuleUtils | rcSeverityMajor
	rcUtilsInvalidSigFile          = 0x04 | rcModuleUtils | rcSeverityMajor
	rcUtilsInvalidLoggerVersion    = 0x05 | rcModuleUtils | rcSeverityMajor
	rcUtilsFailedWriteFile         = 0x06 | rcModuleUtils | rcSeverityMajor
	rcUtilsFailedReadFile          = 0x07 | rcModuleUtils | rcSeverityMajor
	rcUtilsFailedOpenSrcFile       = 0x08 | rcModuleUtils | rcSeverityMajor
	rcUtilsFailedOpenCreateDstFile = 0x09 | rcModuleUtils | rcSeverityMajor
	rcUtilsFailedCloseFile         = 0x0A | rcModuleUtils | rcSeverityMajor
	rcUtilsFileDoesNotExits        = 0x0B | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketTlsRng            = 0x0C | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketTlsCa             = 0x0D | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketTlsClientPem      = 0x0E | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketTlsClientPk       = 0x0F | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketTlsConf           = 0x10 | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketTlsAuthmode       = 0x11 | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketClientAuth        = 0x12 | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketTlsHandshake      = 0x13 | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketConnect           = 0x14 | rcModuleUtils | rcSeverityMajor
	rcUtilsSocketRecv              = 0x15 | rcModuleUtils | rcSeverityMinor
	rcUtilsSocketSend              = 0x16 | rcModuleUtils | rcSeverityMinor
)

var (
	ErrUtilsFailedRemoveFile        = errors.New("failed to remove file")
	ErrUtilsFailedToCopyFile        = errors.New("failed to copy file")
	ErrUtilsFailedToOpenFile        = errors.New("failed to open file")
	ErrUtilsInvalidSigFile          = errors.New("invalid signature file")
	ErrUtilsInvalidLoggerVersion    = errors.New("invalid logger version")
	ErrUtilsFailedWriteFile         = errors.New("failed writing file")
	ErrUtilsFailedReadFile          = errors.New("failed reading file")
	ErrUtilsFailedOpenSrcFile       = errors.New("failed opening source file")
	ErrUtilsFailedOpenCreateDstFile = errors.New("failed to create/open destination file")
	ErrUtilsFailedCloseFile         = errors.New("failed to close file")
	ErrUtilsFileDoesNotExist        = errors.New("file does not exist")
	ErrUtilsSocketTLSRNG            = errors.New("rng failed")
	ErrUtilsSocketTLSCA             = errors.New("CA invalid")
	ErrUtilsSocketTLSClientPEM      = errors.New("client PEM invalid")
	ErrUtilsSocketTLSClientPK       = errors.New("client public key invalid")
	ErrUtilsSocketTLSConf           = errors.New("configuration invalid")
	ErrUtilsSocketTLSAuthMode       = errors.New("auth mode invalid")
	ErrUtilsSocketTLSClientAuth     = errors.New("client auth failed")
	ErrUtilsSocketTLSHandshake      = errors.New("handshake failed")
	ErrUtilsSocketConnect           = errors.New("connect failed")
	ErrUtilsSocketReceive           = errors.New("receive failed")
	ErrUtilsSocketSend              = errors.New("send failed")
)

// processor component
const (
	rcProcessorInvalidTransaction = 0x01 | rcModuleProcessor | rcSeverityModerate
	rcProcessorInvalidRequest     = 0x02 | rcModuleProcessor | rcSeverityModerate
)

var (
	ErrProcessorInvalidTransaction = errors.New("invalid transaction")
	ErrProcessorInvalidRequest     = errors.New("invalid request")
)

// conf
const (
	rcConfInvalidArgument = 0x01 | rcModuleConf | rcSeverityFatal
	rcConfMissingArgument = 0x02 | rcModuleConf | rcSeverityFatal
	rcConfUnknownOption   = 0x03 | rcModuleConf | rcSeverityFatal
	rcConfParserError     = 0x04 | rcModuleConf | rcSeverityFatal
)

var (
	ErrConfInvalidArgument = errors.New("invalid argument")
	ErrConfMissingArgument = errors.New("missing argument")
	ErrConfUnknownOption   = errors.New("unknown option")
	ErrConfParserError     = errors.New("parses error occurred")
)

// api
const (
	rcAPIMaxGetTrytes                = 0x01 | rcModuleAPI | rcSeverityModerate
	rcAPIFindTransactionsNoInput     = 0x02 | rcModuleAPI | rcSeverityModerate
	rcAPIMaxFindTransactions         = 0x3 | rcModuleAPI | rcSeverityModerate
	rcAPIInvalidDepthInput           = 0x04 | rcModuleAPI | rcSeverityModerate
	rcAPIInvalidSubtangleStatus      = 0x05 | rcModuleAPI | rcSeverityModerate
	rcAPITailMissing                 = 0x06 | rcModuleAPI | rcSeverityModerate
	rcAPINotTail                     = 0x07 | rcModuleAPI | rcSeverityModerate
	rcAPIGetBalancesInvalidThreshold = 0x08 | rcModuleAPI | rcSeverityMinor
	rcAPIGetBalancesUnknownTip       = 0x09 | rcModuleAPI | rcSeverityMinor
	rcAPIGetBalancesInconsistentTip  = 0x0A | rcModuleAPI | rcSeverityMinor
)

var (
	ErrAPIMaxGetTrytes                = errors.New("max trytes exceeded")
	ErrAPIFindTransactionsNoInput     = errors.New("no input given for findTransactions")
	ErrAPIMaxFindTransactions         = errors.New("find transactions result limit exceeded")
	ErrAPIInvalidDepthInput           = errors.New("invalid depth input")
	ErrAPIInvalidSubtangleStatus      = errors.New("invalid subtangle status")
	ErrAPITailMissing                 = errors.New("tail missing")
	ErrAPINotTail                     = errors.New("not a tail tx")
	ErrAPIGetBalancesInvalidThreshold = errors.New("invalid threshold in getBalances command")
	ErrAPIGetBalancesUnknownTip       = errors.New("unknown tip in getBalances command")
	ErrAPIGetBalancesInconsistentTip  = errors.New("inconsistent tip given in getBalances command")
)

// node
const (
	rcNodeSetPacketTransactionFailed = 0x01 | rcModuleNode | rcSeverityModerate
	rcNodeSetPacketRequestFailed     = 0x02 | rcModuleNode | rcSeverityModerate
)

var (
	ErrNodeSetPacketTransactionFailed = errors.New("setting transaction packet failed")
	ErrnodeSetPacketRequestFailed     = errors.New("setting request packet failed")
)

// MAM
const (
	rcMAMBufferTooSmall             = 0x01 | rcModuleMAM | rcSeverityModerate
	rcMAMInvalidArgument            = 0x02 | rcModuleMAM | rcSeverityModerate
	rcMAMInvalidValue               = 0x03 | rcModuleMAM | rcSeverityModerate
	rcMAMNegativeValue              = 0x04 | rcModuleMAM | rcSeverityModerate
	rcMAMInternalError              = 0x05 | rcModuleMAM | rcSeverityModerate
	rcMAMNotImplemented             = 0x06 | rcModuleMAM | rcSeverityModerate
	rcMAMPB3EOF                     = 0x07 | rcModuleMAM | rcSeverityModerate
	rcMAMPB3BadOneof                = 0x08 | rcModuleMAM | rcSeverityModerate
	rcMAMPB3BadOptional             = 0x09 | rcModuleMAM | rcSeverityModerate
	rcMAMPB3BadRepeated             = 0x0A | rcModuleMAM | rcSeverityModerate
	rcMAMPB3BadMac                  = 0x0B | rcModuleMAM | rcSeverityModerate
	rcMAMPB3BadSig                  = 0x0C | rcModuleMAM | rcSeverityModerate
	rcMAMPB3BadEkey                 = 0x0D | rcModuleMAM | rcSeverityModerate
	rcMAMTritsSizeTNotSupported     = 0x0E | rcModuleMAM | rcSeverityModerate
	rcMAMChannelNotFound            = 0x0F | rcModuleMAM | rcSeverityModerate
	rcMAMEndpointNotFound           = 0x10 | rcModuleMAM | rcSeverityModerate
	rcMAMVersionNotSupported        = 0x11 | rcModuleMAM | rcSeverityModerate
	rcMAMChannelNotTrusted          = 0x12 | rcModuleMAM | rcSeverityModerate
	rcMAMEndpointNotTrusted         = 0x13 | rcModuleMAM | rcSeverityModerate
	rcMAMKeyloadIrrelevant          = 0x14 | rcModuleMAM | rcSeverityModerate
	rcMAMKeyloadOverloaded          = 0x15 | rcModuleMAM | rcSeverityModerate
	rcMAMBundleNotEmpty             = 0x16 | rcModuleMAM | rcSeverityModerate
	rcMAMBundleDoesNotContainHeader = 0x17 | rcModuleMAM | rcSeverityModerate
	rcMAMRecvCtxNotFound            = 0x18 | rcModuleMAM | rcSeverityModerate
	rcMAMSendCtxNotFound            = 0x19 | rcModuleMAM | rcSeverityModerate
	rcMAMMessageNotFound            = 0x1A | rcModuleMAM | rcSeverityModerate
	rcMAMBadPacketOrd               = 0x1B | rcModuleMAM | rcSeverityModerate
	rcMAMMSSExhausted               = 0x1C | rcModuleMAM | rcSeverityModerate
	rcMAMNTRUPolyFailed             = 0x1D | rcModuleMAM | rcSeverityModerate
	rcMAMAPIFailedCreateEndpoint    = 0x1E | rcModuleMAM | rcSeverityModerate
	rcMAMAPIFailedCreateChannel     = 0x1F | rcModuleMAM | rcSeverityModerate
	rcMAMMSSNotFound                = 0x20 | rcModuleMAM | rcSeverityModerate
)

var (
	ErrMAMBufferTooSmall             = errors.New("buffer too small")
	ErrMAMInvalidArgument            = errors.New("invalid argument")
	ErrMAMInvalidValue               = errors.New("invalid value")
	ErrMAMNegativeValue              = errors.New("negative value")
	ErrMAMInternalError              = errors.New("internal error")
	ErrMAMNotImplemented             = errors.New("not implemented")
	ErrMAMPB3EOF                     = errors.New("PB3 EOF")
	ErrMAMPB3BadOneOf                = errors.New("PB3 bad oneof")
	ErrMAMPB3BadOptional             = errors.New("PB3 bad optional")
	ErrMAMPB3BadRepeated             = errors.New("PB3 bad repeated")
	ErrMAMPB3BadMAC                  = errors.New("PB3 bad MAC")
	ErrMAMPB3BadSig                  = errors.New("PB3 bad signature")
	ErrMAMPB3BadEKey                 = errors.New("PB3 bad encryption key")
	ErrMAMTritsSizeTNotSupported     = errors.New("trits size_t not supported")
	ErrMAMChannelNotFound            = errors.New("channel not found")
	ErrMAMEndpointNotFound           = errors.New("endpoint not found")
	ErrMAMVersionNotSupported        = errors.New("version not supported")
	ErrMAMChannelNotTrusted          = errors.New("channel not trusted")
	ErrMAMEndpointNotTrusted         = errors.New("endpoint not trusted")
	ErrMAMKeyloadIrrelevant          = errors.New("keyload irrelevant")
	ErrMAMKeyloadOverloaded          = errors.New("keyload overloaded")
	ErrMAMBundleNotEmpty             = errors.New("bundle not empty")
	ErrMAMBundleDoesNotContainHeader = errors.New("bundle does not contain header")
	ErrMAMRecvCtxNotFound            = errors.New("receive context not found")
	ErrMAMSendCtxNotFound            = errors.New("send context not found")
	ErrMAMMessageNotFound            = errors.New("message not found")
	ErrMAMBadPacketOrdering          = errors.New("bad packet ordering")
	ErrMAMMSSExhausted               = errors.New("MSS exhausted")
	ErrMAMNTRUPolyFailed             = errors.New("NTRU poly failed")
	ErrMAMAPIFailedCreateEndpoint    = errors.New("failed to create endpoint")
	ErrMAMAPIFailedCreateChannel     = errors.New("failed to create channel")
	ErrMAMMSSNotFound                = errors.New("MSS not found")
)

// helpers
const (
	rcHelpersPowInvalidTx = 0x01 | rcModuleHelpers | rcSeverityModerate
)

var ErrHelpersPoWInvalidTx = errors.New("invalid tx for PoW")

// crypto
const (
	rcCryptoUnsupportedSpongeType = 0x01 | rcModuleCrypto | rcSeverityMajor
)

var ErrCryptoUnsupportedSpongeType = errors.New("unsupported sponge type")

// common
const (
	rcCommonBundleSign = 0x01 | rcModuleCommon | rcSeverityMinor
)

var ErrCommonBundleSign = errors.New("error signing bundle")

// handle
const (
	rcThreadCreate = 0x01 | rcModuleHandle | rcSeverityFatal
	rcThreadJoin   = 0x02 | rcModuleHandle | rcSeverityModerate
	rcCondInit     = 0x03 | rcModuleHandle | rcSeverityFatal
	rcCondSignal   = 0x04 | rcModuleHandle | rcSeverityFatal
	rcCondDestroy  = 0x05 | rcModuleHandle | rcSeverityFatal
	rcLockInit     = 0x06 | rcModuleHandle | rcSeverityFatal
	rcLockDestroy  = 0x07 | rcModuleHandle | rcSeverityFatal
)

var (
	ErrThreadCreate = errors.New("couldn't create therad")
	ErrThreadJoin   = errors.New("couldn't join to thread")
	ErrCondInit     = errors.New("cond init failed")
	ErrCondSignal   = errors.New("cond signal failed")
	ErrCondDestroy  = errors.New("cond destroy failed")
	ErrLockInit     = errors.New("lock init failed")
	ErrLockDestroy  = errors.New("lock destroy failed")
)

var rcToErrMap = map[int]error{
	rcNullParam:    ErrNullParam,
	rcInvalidParam: ErrInvalidParam,
	rcOOM:          ErrOutOfMemory,
	rcStillRunning: ErrStillRunning,
	//
	rcSQLite3FailedOpenDb:            ErrSQLite3FailedOpenDb,
	rcSQLite3FailedInsertDb:          ErrSQLite3FailedInsertDB,
	rcSQLite3NoPathForDbSpecified:    ErrSQLite3NoPathForDbSpecified,
	rcSQLite3FailedNotImplemented:    ErrSQLite3FailedNotImplemented,
	rcSQLite3FailedBinding:           ErrSQLite3FailedBinding,
	rcSQLite3FailedPreparedStatement: ErrSQLite3FailedPreparedStatement,
	rcSQLite3FailedFinalize:          ErrSQLite3FailedFinalize,
	rcSQLite3FailedStep:              ErrSQlite3FailedStep,
	rcSQLite3FailedBegin:             ErrSQLite3FailedBegin,
	rcSQLite3FailedEnd:               ErrSQLite3FailedEnd,
	rcSQLite3FailedRollback:          ErrSQLite3FailedRollback,
	rcSQLite3FailedConfig:            ErrSQLite3FailedConfig,
	rcSQLite3FailedInitialize:        ErrSQLite3FailedInitialize,
	rcSQLite3FailedShutdown:          ErrSQLite3FailedShutdown,
	//
	rcNeighborFailedUriParsing:      ErrNeighborFailedURIParsing,
	rcNeighborInvalidProtocol:       ErrNeighborInvalidProtocol,
	rcNeighborInvalidHost:           ErrNeighborInvalidHost,
	rcNeighborFailedSend:            ErrNeighborFailedSend,
	rcNeighborFailedEndpointInit:    ErrNeighborFailedEndpointInit,
	rcNeighborFailedEndpointDestroy: ErrNeighborFailedEndpointDestroy,
	rcNeighborAlreadyPaired:         ErrNeighborAlreadyPaired,
	rcNeighborNotPaired:             ErrNeighborNotPaired,
	//
	rcCClientJsonCreate:          ErrCClientJSONCreate,
	rcCClientJsonParse:           ErrCClientJSONParsing,
	rcCClientHttp:                ErrCClientHTTP,
	rcCClientHttpReq:             ErrCClientHTTPReq,
	rcCClientResError:            ErrCClientHTTPRes,
	rcCClientJsonKey:             ErrCClientJSONKey,
	rcCClientFlexTrits:           ErrCClientFlexTrits,
	rcCClientNullPtr:             ErrCClientNullPointer,
	rcCClientUnimplemented:       ErrCClientUnimplemented,
	rcCClientInvalidSecurity:     ErrCClientInvalidSecurity,
	rcCClientTxDeserializeFailed: ErrCClientTxDeserializeFailed,
	rcCClientInsufficientBalance: ErrCClientInsufficientBalance,
	rcCClientPowFailed:           ErrCClientPowFailed,
	rcCClientInvalidTransfer:     ErrCClientInvalidTransfer,
	rcCClientInvalidTailHash:     ErrCClientInvalidTailHash,
	rcCClientInvalidBundle:       ErrCClientInvalidBundle,
	rcCClientCheckBalance:        ErrCClientCheckBalance,
	rcCClientNotPromotable:       ErrCClientNotPromotable,
	//
	rcConsensusNotImplemented: ErrConsensusNotImplemented,
	//
	rcCWFailedInDfsFromDb: ErrCWFailedInDFSFromDB,
	rcCWFailedInLightDFS:  ErrCWFailedInLightDFS,
	//
	rcExitProbabilitiesInvalidEntrypoint: ErrExitProbabilitiesInvalidEntrypoint,
	rcExitProbabilitiesMissingRating:     ErrExitProbabilitiesMissingRating,
	rcExitProbabilitiesNotImplemented:    ErrExitProbabilitiesNotImplemented,
	//
	rcSnapshotFileNotFound:                  ErrSnapshotFileNotFound,
	rcSnapshotInvalidFile:                   ErrSnapshotInvalidFile,
	rcSnapshotInvalidSupply:                 ErrSnapshotInvalidSupply,
	rcSnapshotInconsistentSnapshot:          ErrSnapshotInconsistentSnapshot,
	rcSnapshotInconsistentPatch:             ErrSnapshotInconsistentPatch,
	rcSnapshotBalanceNotFound:               ErrSnapshotBalanceNotFound,
	rcSnapshotInvalidSignature:              ErrSnapshotInvalidSignature,
	rcSnapshotFailedJsonParsing:             ErrSnapshotFailedJSONParsing,
	rcSnapshotServiceNotEnoughDepth:         ErrSnapshotServiceNotEnoughDepth,
	rcSnapshotServiceMilestoneNotLoaded:     ErrSnapshotServiceMilestoneNotLoaded,
	rcSnapshotServiceMilestoneNotSolid:      ErrSnapshotServiceMilestoneNotSolid,
	rcSnapshotServiceMilestoneTooOld:        ErrSnapshotServiceMilestoneTooOld,
	rcSnapshotMetadataFailedDeserializing:   ErrSnapshotMetadataFailedDeserializing,
	rcSnapshotMetadataFailedSerializing:     ErrSnapshotMetadataFailedSerializing,
	rcSnapshotStateDeltaFailedDeserializing: ErrSnapshotStateDeltaFailedDeserializing,
	rcSnapshotMissingMilestoneTransaction:   ErrSnapshotMissingMilestoneTransaction,
	//
	rcLedgerValidatorInvalidTransaction:    ErrLedgerValidatorInvalidTransaction,
	rcLedgerValidatorCouldNotLoadMilestone: ErrLedgerValidatorCouldNotLoadMilestone,
	rcLedgerValidatorInconsistentDelta:     ErrLedgerValidatorInconsistentDelta,
	rcLedgerValidatorTransactionNotSolid:   ErrLedgerValidatorTransactionNotSolid,
	//
	rcTipSelectorTipsNotConsistent: ErrTipSelectorTipsNotConsistent,
	rcTipSelectorReferenceTooOld:   ErrTipSelectorReferenceTooOld,
	//
	rcTangleTailNotFound: ErrTangleTailNotFound,
	rcTangleNotATail:     ErrTangleNotATail,
	//
	rcUtilsFailedRemoveFile:        ErrUtilsFailedRemoveFile,
	rcUtilsFailedToCopyFile:        ErrUtilsFailedToCopyFile,
	rcUtilsFailedToOpenFile:        ErrUtilsFailedToOpenFile,
	rcUtilsInvalidSigFile:          ErrUtilsInvalidSigFile,
	rcUtilsInvalidLoggerVersion:    ErrUtilsInvalidLoggerVersion,
	rcUtilsFailedWriteFile:         ErrUtilsFailedWriteFile,
	rcUtilsFailedReadFile:          ErrUtilsFailedReadFile,
	rcUtilsFailedOpenSrcFile:       ErrUtilsFailedOpenSrcFile,
	rcUtilsFailedOpenCreateDstFile: ErrUtilsFailedOpenCreateDstFile,
	rcUtilsFailedCloseFile:         ErrUtilsFailedCloseFile,
	rcUtilsFileDoesNotExits:        ErrUtilsFileDoesNotExist,
	rcUtilsSocketTlsRng:            ErrUtilsSocketTLSRNG,
	rcUtilsSocketTlsCa:             ErrUtilsSocketTLSCA,
	rcUtilsSocketTlsClientPem:      ErrUtilsSocketTLSClientPEM,
	rcUtilsSocketTlsClientPk:       ErrUtilsSocketTLSClientPK,
	rcUtilsSocketTlsConf:           ErrUtilsSocketTLSConf,
	rcUtilsSocketTlsAuthmode:       ErrUtilsSocketTLSAuthMode,
	rcUtilsSocketClientAuth:        ErrUtilsSocketTLSClientAuth,
	rcUtilsSocketTlsHandshake:      ErrUtilsSocketTLSHandshake,
	rcUtilsSocketConnect:           ErrUtilsSocketConnect,
	rcUtilsSocketRecv:              ErrUtilsSocketReceive,
	rcUtilsSocketSend:              ErrUtilsSocketSend,
	//
	rcProcessorInvalidTransaction: ErrProcessorInvalidTransaction,
	rcProcessorInvalidRequest:     ErrProcessorInvalidRequest,
	//
	rcConfInvalidArgument: ErrConfInvalidArgument,
	rcConfMissingArgument: ErrConfMissingArgument,
	rcConfUnknownOption:   ErrConfUnknownOption,
	rcConfParserError:     ErrConfParserError,
	//
	rcAPIMaxGetTrytes:                ErrAPIMaxGetTrytes,
	rcAPIFindTransactionsNoInput:     ErrAPIFindTransactionsNoInput,
	rcAPIMaxFindTransactions:         ErrAPIMaxFindTransactions,
	rcAPIInvalidDepthInput:           ErrAPIInvalidDepthInput,
	rcAPIInvalidSubtangleStatus:      ErrAPIInvalidSubtangleStatus,
	rcAPITailMissing:                 ErrAPITailMissing,
	rcAPINotTail:                     ErrAPINotTail,
	rcAPIGetBalancesInvalidThreshold: ErrAPIGetBalancesInvalidThreshold,
	rcAPIGetBalancesUnknownTip:       ErrAPIGetBalancesUnknownTip,
	rcAPIGetBalancesInconsistentTip:  ErrAPIGetBalancesInconsistentTip,
	//
	rcNodeSetPacketTransactionFailed: ErrNodeSetPacketTransactionFailed,
	rcNodeSetPacketRequestFailed:     ErrnodeSetPacketRequestFailed,
	//
	rcMAMBufferTooSmall:             ErrMAMBufferTooSmall,
	rcMAMInvalidArgument:            ErrMAMInvalidArgument,
	rcMAMInvalidValue:               ErrMAMInvalidValue,
	rcMAMNegativeValue:              ErrMAMNegativeValue,
	rcMAMInternalError:              ErrMAMInternalError,
	rcMAMNotImplemented:             ErrMAMNotImplemented,
	rcMAMPB3EOF:                     ErrMAMPB3EOF,
	rcMAMPB3BadOneof:                ErrMAMPB3BadOneOf,
	rcMAMPB3BadOptional:             ErrMAMPB3BadOptional,
	rcMAMPB3BadRepeated:             ErrMAMPB3BadRepeated,
	rcMAMPB3BadMac:                  ErrMAMPB3BadMAC,
	rcMAMPB3BadSig:                  ErrMAMPB3BadSig,
	rcMAMPB3BadEkey:                 ErrMAMPB3BadEKey,
	rcMAMTritsSizeTNotSupported:     ErrMAMTritsSizeTNotSupported,
	rcMAMChannelNotFound:            ErrMAMChannelNotFound,
	rcMAMEndpointNotFound:           ErrMAMEndpointNotFound,
	rcMAMVersionNotSupported:        ErrMAMVersionNotSupported,
	rcMAMChannelNotTrusted:          ErrMAMChannelNotTrusted,
	rcMAMEndpointNotTrusted:         ErrMAMEndpointNotTrusted,
	rcMAMKeyloadIrrelevant:          ErrMAMKeyloadIrrelevant,
	rcMAMKeyloadOverloaded:          ErrMAMKeyloadOverloaded,
	rcMAMBundleNotEmpty:             ErrMAMBundleNotEmpty,
	rcMAMBundleDoesNotContainHeader: ErrMAMBundleDoesNotContainHeader,
	rcMAMRecvCtxNotFound:            ErrMAMRecvCtxNotFound,
	rcMAMSendCtxNotFound:            ErrMAMSendCtxNotFound,
	rcMAMMessageNotFound:            ErrMAMMessageNotFound,
	rcMAMBadPacketOrd:               ErrMAMBadPacketOrdering,
	rcMAMMSSExhausted:               ErrMAMMSSExhausted,
	rcMAMNTRUPolyFailed:             ErrMAMNTRUPolyFailed,
	rcMAMAPIFailedCreateEndpoint:    ErrMAMAPIFailedCreateEndpoint,
	rcMAMAPIFailedCreateChannel:     ErrMAMAPIFailedCreateChannel,
	rcMAMMSSNotFound:                ErrMAMMSSNotFound,
	//
	rcHelpersPowInvalidTx: ErrHelpersPoWInvalidTx,
	//
	rcCryptoUnsupportedSpongeType: ErrCryptoUnsupportedSpongeType,
	//
	rcCommonBundleSign: ErrCommonBundleSign,
	//
	rcThreadCreate: ErrThreadCreate,
	rcThreadJoin:   ErrThreadJoin,
	rcCondInit:     ErrCondInit,
	rcCondSignal:   ErrCondSignal,
	rcCondDestroy:  ErrCondDestroy,
	rcLockInit:     ErrLockInit,
	rcLockDestroy:  ErrLockDestroy,
}

func wrapError(code int) error {
	if code == rcOK {
		return nil
	}
	err, ok := rcToErrMap[code]
	if ok {
		return err
	}
	return ErrEntangledUnknown
}
