package base

import "time"

const (
	ServiceName    = "inshorts-news-srv" // ServiceName
	PropSystem     = "system"            // PropSystem log property system
	HdrXRefNo      = "x_ref_number"      // Header attribute which holds x-ref-number value to trace in logs
	EnvLogLevel    = "LOG_LEVEL"         // Log level
	EnvPort        = "PORT"              // Service address e.g. 3000
	EnvContextPath = "CONTEXT_PATH"      // Context path
	EnvAPIKey      = "API_KEY"           // apikey
	EnvESUserName  = "ES_USERNAME"       // ES user name
	EnvESPassword  = "ES_PASSWORD"       // ES Password
	EnvESUrl       = "ES_URL"            // ES URL
	Env            = "ENV"               // ENV
	ESIndex        = "inshorts-news"     // ES index name
)

// response messages
const (
	ParsingRequestMsg         = "error while parsing body request data!" // ParsingRequestMsg error message
	SystemParamsMissingMsg    = "required system parameters missing!"    // SystemParamsMissingMsg error message
	ResourceNotFoundMsg       = "resource not found"                     // ResourceNotFoundMsg error msg
	ErrorProcessingRequestMsg = "error while processing request!"        // ErrorProcessingRequest failure message
	ParamsMissingMsg          = "parameters missing or wrong!"           // ParamsMissing which prevents fullfilling the request
	InvalidDataMsg            = "invalid data"                           // invalid data in payload
	DuplicateDataMsg          = "duplicate data could not be processed"  // duplicate data could not be processed
	InvalidOrMissingTokenMsg  = "invalid or missing token in env"        // Invalid or missing token in env
)

// service error codes
const (
	CodeInternalError = "10100" // Internal Error
	CodeInvalidData   = "10101" // Invalid data passed which could not be marshalled
	CodeMissingParams = "10102" // missing params
	CodeDataNotFound  = "10103" // entity not found
	CodeAccessDenied  = "10105" // access denied for this resource
)

const (
	DefaultPageSize        = 5
	MaxWindowSize          = 10000
	DefaultMinScore        = 0.7
	MaxArticlesForTrending = 1000
)

// Trending service const
const (
	DefaultLimit     = 5
	MaxLimit         = 50
	CacheTTL         = 2 * time.Minute
	MaxEventAge      = 48 * time.Hour
	EventRadiusKm    = 50
	EventInterval    = 5 * time.Second
	PruneInterval    = 1 * time.Minute
	RecencyHalfLifeH = 24
)
