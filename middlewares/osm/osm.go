package valse

import (
	"fmt"
	"net/url"

	"github.com/xwinie/osm"
)

// DbConfig database orm config
type DbConfig struct {
	DbUser     string
	DbPassword string
	DbHost     string
	DbName     string
	DbType     string
	DbPort     string
	DbPath     []string
	DbCharset  string
}

// OsmConnect connect database return osm struct
func OsmConnect(config DbConfig) (*osm.Osm, error) {
	return osm.New(config.DbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&loc=%v&parseTime=true",
		config.DbUser,
		config.DbPassword,
		config.DbHost,
		config.DbPort,
		config.DbName,
		config.DbCharset,
		url.QueryEscape("Asia/Shanghai")), config.DbPath)
}
