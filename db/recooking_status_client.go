package db

import (
	"fmt"
	"time"
)

const (
	XcrpModuleNameColumnName     = "module"
	XcrpPartionIdColumnName      = "partition_id"
	XcrpStateColumnName          = "state"
	XcrpUpdatedTimeColumnName    = "updated_time"
	XcrpRecookingStatusTableName = "RecookingStatus"
	XcrpAppNameColumnName        = "app_name"
	XcrpShardIdColumnName        = "shard_id"
)

const (
	PrecookInitialized = iota
	PrecookPending
	PrecookComplete
)

type RecookingStatus struct {
	AppName     string    `json:"appName"`
	PartitionId string    `json:"partitionId"`
	State       int       `json:"state"`
	UpdatedTime time.Time `json:"updatedTime"`
}

func (c *CassandraClient) GetRecookingStatus(tenantId string, moduleName string, partitionId string) (int, time.Time, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	if tenantId == "" {
		tenantId = GetDefaultTenantId()
	}

	shardId := GetShardId(moduleName)
	/*
		stmt := fmt.Sprintf(`SELECT %s, %s FROM "%s" WHERE %s = ? AND %s = ? LIMIT 1`, XcrpStateColumnName, XcrpUpdatedTimeColumnName, c.xconfRecookingStatusTableName, XcrpModuleNameColumnName, XcrpPartionIdColumnName)
		query := c.Query(stmt, moduleName, partitionId)
	*/

	stmt := fmt.Sprintf(`SELECT %s, %s FROM "%s" WHERE %s = ? AND %s = ? AND %s = ? AND %s = ? LIMIT 1`, XcrpStateColumnName, XcrpUpdatedTimeColumnName, c.xconfRecookingStatusTableName, TenantIdColumnName, XcrpShardIdColumnName, XcrpModuleNameColumnName, XcrpPartionIdColumnName)
	query := c.Query(stmt, tenantId, shardId, moduleName, partitionId)
	var state int
	var updatedTime time.Time
	err := query.Scan(&state, &updatedTime)
	if err != nil {
		return 0, time.Time{}, err
	}

	return state, updatedTime, nil
}

func (c *CassandraClient) CheckFinalRecookingStatus(tenantId string, moduleName string) (bool, time.Time, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	if tenantId == "" {
		tenantId = GetDefaultTenantId()
	}
	shardId := GetShardId(moduleName)

	var partitionID string
	var state int
	var updatedTime time.Time
	stmt := fmt.Sprintf(`SELECT %s, %s, %s FROM "%s" WHERE %s = ? AND %s = ? AND %s = ?`, XcrpPartionIdColumnName, XcrpStateColumnName, XcrpUpdatedTimeColumnName, c.XconfRecookingStatusTableName(), TenantIdColumnName, XcrpShardIdColumnName, XcrpModuleNameColumnName)
	query := c.Session.Query(stmt, tenantId, shardId, moduleName)
	iter := query.Iter()
	defer iter.Close()

	allComplete := true
	var latestUpdatedTime time.Time

	for iter.Scan(&partitionID, &state, &updatedTime) {
		if state != PrecookComplete {
			allComplete = false
		}
		if updatedTime.After(latestUpdatedTime) {
			latestUpdatedTime = updatedTime
		}
	}

	if err := iter.Close(); err != nil {
		return false, time.Time{}, err
	}

	return allComplete, latestUpdatedTime, nil
}

func (c *CassandraClient) SetRecookingStatus(tenantId string, moduleName string, partitionId string, state int) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	if tenantId == "" {
		tenantId = GetDefaultTenantId()
	}
	shardId := GetShardId(moduleName)

	stmt := fmt.Sprintf(`INSERT INTO "%s" (%s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?)`, c.XconfRecookingStatusTableName(), TenantIdColumnName, XcrpShardIdColumnName, XcrpModuleNameColumnName, XcrpPartionIdColumnName, XcrpStateColumnName, XcrpUpdatedTimeColumnName)
	updatedTime := time.Now()
	err := c.Query(stmt, tenantId, shardId, moduleName, partitionId, state, updatedTime).Exec()
	if err != nil {
		return err
	}

	return nil

}

func (c *CassandraClient) GetRecookingStatusDetails(tenantId, module string) ([]RecookingStatus, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	if tenantId == "" {
		tenantId = GetDefaultTenantId()
	}

	stmt := fmt.Sprintf(`SELECT %s, %s, %s, %s FROM "%s" WHERE %s = ? AND %s IN ?`,
		XcrpAppNameColumnName, XcrpPartionIdColumnName, XcrpStateColumnName, XcrpUpdatedTimeColumnName, XcrpRecookingStatusTableName, TenantIdColumnName, XcrpShardIdColumnName)
	query := c.Session.Query(stmt, tenantId, GetShardIds())
	iter := query.Iter()
	defer iter.Close()
	var statuses []RecookingStatus
	var status RecookingStatus
	for iter.Scan(&status.AppName, &status.PartitionId, &status.State, &status.UpdatedTime) {
		statuses = append(statuses, status)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return statuses, nil
}
