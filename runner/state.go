package main

type PanelResult struct {
	Exception   interface{}    `json:"exception"`
	Value       *[]interface{} `json:"value"`
	Preview     string         `json:"preview"`
	Stdout      string         `json:"stdout"`
	Shape       Shape          `json:"shape"`
	ArrayCount  *float64       `json:"float64"`
	Size        *float64       `json:"size"`
	ContentType string         `json:"contentType"`
	Elapsed     *float64       `json:"elapsed"`
}

var defaultPanelResult = PanelResult{
	Stdout:      "",
	Shape:       defaultShape,
	Preview:     "",
	Size:        nil,
	ContentType: "unknown",
	Value:       nil,
	Exception:   nil,
	ArrayCount:  nil,
}

type Encrypt struct {
	Value     string `json:"value"`
	Encrypted bool   `json:"encrypted"`
}

type ServerInfoType string

const (
	SSHAgent      = "ssh-agent"
	SSHPassword   = "password"
	SSHPrivateKey = "private-key"
)

type ServerInfo struct {
	Name           string         `json:"name"`
	Address        string         `json:"address"`
	Port           float64        `json:"port"`
	Type           ServerInfoType `json:"type"`
	Username       string         `json:"username"`
	Password       Encrypt        `json:"password_encrypt"`
	PrivateKeyFile string         `json:"privateKeyFile"`
	Passphrase     Encrypt        `json:"passphrase_encrypt"`
	Id             string         `json:"id"`
}

var defaultServerInfo = ServerInfo{
	Type:           SSHPrivateKey,
	Name:           "Untitled Server",
	Address:        "",
	Port:           22,
	Username:       "",
	Password:       Encrypt{},
	PrivateKeyFile: "~/.ssh/id_rsa",
	Passphrase:     Encrypt{},
}

type ContentTypeInfo struct {
	Type             string `json:"type"`
	CustomLineRegexp string `json:"customLineRegexp"`
}

var defaultContentTypeInfo = ContentTypeInfo{}

type PanelInfoType string

const (
	HttpPanel     = "http"
	ProgramPanel  = "program"
	LiteralPanel  = "literal"
	FilePanel     = "file"
	FilaggPanel   = "filagg"
	DatabasePanel = "database"
)

type PanelInfo struct {
	Content    string        `json:"content"`
	Type       PanelInfoType `json:"type"`
	Name       string        `json:"name"`
	Id         string        `json:"id"`
	ServerId   string        `json:"serverId"`
	ResultMeta PanelResult   `json:"resultMeta"`
	*ProgramPanelInfo
	*FilePanelInfo
	*LiteralPanelInfo
	*DatabasePanelInfo
	*HttpPanelInfo
}

type SupportedLanguages string

const (
	Python     SupportedLanguages = "python"
	JavaScript                    = "javascript"
	Ruby                          = "ruby"
	R                             = "r"
	Julia                         = "julia"
	SQL                           = "sql"
)

type ProgramPanelInfo struct {
	Program struct {
		Type SupportedLanguages `json:"type"`
	} `json:"program"`
}

type FilePanelInfo struct {
	File struct {
		ContentTypeInfo ContentTypeInfo `json:"contentTypeInfo"`
		Name            string          `json:"name"`
	} `json:"file"`
}

type LiteralPanelInfo struct {
	Literal struct {
		ContentTypeInfo ContentTypeInfo `json:"contentTypeInfo"`
	} `json:"literal"`
}

type HttpPanelInfo struct {
	Http HttpConnectorInfo `json:"http"`
}

type DatabasePanelInfoDatabase struct {
	ConnectorId string      `json:"connectorId"`
	Range       interface{} `json:"range"` // TODO: support these
	Table       string      `json:"table"`
	Step        float64     `json:"step"`
}

type DatabasePanelInfo struct {
	Database DatabasePanelInfoDatabase `json:"database"`
}

type ConnectorInfoType string

const (
	DatabaseConnector ConnectorInfoType = "database"
	HTTPConnector                       = "http"
)

type ConnectorInfo struct {
	Name     string            `json:"name"`
	Type     ConnectorInfoType `json:"type"`
	Id       string            `json:"id"`
	ServerId string            `json:"serverId"`
	*DatabaseConnectorInfo
}

type DatabaseConnectorInfoType string

const (
	PostgresDatabase      DatabaseConnectorInfoType = "postgres"
	MySQLDatabase                                   = "mysql"
	SQLiteDatabase                                  = "sqlite"
	OracleDatabase                                  = "oracle"
	SQLServerDatabase                               = "sqlserver"
	PrestoDatabase                                  = "presto"
	ClickhouseDatabase                              = "clickhouse"
	SnowflakeDatabase                               = "snowflake"
	CassandraDatabase                               = "cassandra"
	ElasticsearchDatabase                           = "elasticsearch"
	SplunkDatabase                                  = "splunk"
	PrometheusDatabase                              = "prometheus"
	InfluxDatabase                                  = "influx"
)

type DatabaseConnectorInfoDatabase struct {
	Type     DatabaseConnectorInfoType `json:"type"`
	Database string                    `json:"database"`
	Username string                    `json:"username"`
	Password Encrypt                   `json:"password_encrypt"`
	Address  string                    `json:"address"`
	ApiKey   Encrypt                   `json:"apiKey_encrypt"`
	Extra    map[string]string         `json:"extra"`
}

type DatabaseConnectorInfo struct {
	Database DatabaseConnectorInfoDatabase `json:"database"`
}

type HttpConnectorInfoHttp struct {
	Method          string          `json:"method"`
	Url             string          `json:"url"`
	ContentTypeInfo ContentTypeInfo `json:"contentTypeInfo"`
	Headers         [][]string      `json:"headers"`
}

type HttpConnectorInfo struct {
	Http HttpConnectorInfoHttp `json:"http"`
}

type ProjectPage struct {
	Panels    []PanelInfo   `json:"panels"`
	Schedules []interface{} `json:"schedules"`
	Name      string        `json:"name"`
	Id        string        `json:"id"`
}

type ProjectState struct {
	Pages           []ProjectPage   `json:"pages"`
	Connectors      []ConnectorInfo `json:"connectors"`
	ProjectName     string          `json:"projectName"`
	Id              string          `json:"id"`
	OriginalVersion string          `json:"originalVersion"`
	LastVersion     string          `json:"lastVersion"`
}