package db

import (
	"fmt"
)

const (
	RfcPrecookHashColumnName    = "ref_id"
	RfcPrecookPayloadColumnName = "payload"
	VersionColumnName           = "version"
)

// GetPrecookDataFromXPC Get Precook data from XPC
func (c *CassandraClient) GetPrecookDataFromXPC(RfcPrecookHash string) ([]byte, string, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	stmt := fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = ? LIMIT 1", RfcPrecookPayloadColumnName, VersionColumnName, fmt.Sprintf("%s.%s", c.XpcKeyspace(), c.XpcPrecookTableName()), RfcPrecookHashColumnName)
	query := c.Query(stmt, RfcPrecookHash)
	var payload []byte
	var version string
	err := query.Scan(&payload, &version)
	if err != nil {
		return nil, "", err
	}

	return payload, version, nil
}

// setPrecookDataInXPC Set Precook data in XPC
func (c *CassandraClient) SetPrecookDataInXPC(RfcPrecookHash string, RfcPrecookPayload []byte) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	stmt := fmt.Sprintf("INSERT INTO %s.%s (%s, %s) VALUES (?, ?)", c.XpcKeyspace(), c.XpcPrecookTableName(), RfcPrecookHashColumnName, RfcPrecookPayloadColumnName)
	err := c.Query(stmt, RfcPrecookHash, RfcPrecookPayload).Exec()
	if err != nil {
		return err
	}

	return nil
}
