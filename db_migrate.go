//go:build migrate

/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// Command migrate copies data from the legacy xconf Cassandra schema
// (db_create_tables.cql) to the new multi-tenant sharded schema
// (db_create_tables_xconf.cql).
//
// Old schema tables use (key text, column1 text, value blob) with no tenant/shard
// partitioning. New schema tables add (tenant_id, shard_id) to the partition key
// and an updated timestamp column.
//
// Prerequisites:
//   - Source keyspace must contain old-schema tables (from db_create_tables.cql).
//   - Destination keyspace must already have new-schema tables created
//     (run db_create_tables_xconf.cql first).
//
// Usage:
//
//	migrate [-f <config_file>] \
//	        [-table <OldTableName>] \
//	        [-clear] \
//	        [-dry-run]
//
// Flags:
//
//	-f: Path to xconfwebconfig config file (default: /app/xconfwebconfig/xconfwebconfig.conf).
//	-table: Migrate only the specified legacy table name from tableMappings.
//	-clear: TRUNCATE each destination table before migrating so the run starts on a clean slate.
//	        Combine with -dry-run to log which tables would be truncated without touching the DB.
//	-dry-run: Read/count rows and log planned writes without writing destination rows.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
)

// migrationKind describes how a legacy table maps to its new-schema counterpart.
type migrationKind string

const (
	defaultConfigFile = "/app/xconfwebconfig/xconfwebconfig.conf"

	// kindSimple: (key, column1, value) → (tenant_id, shard_id, key, value, updated)
	// The column1 clustering key is dropped; if a key has multiple column1 rows the
	// last-written value wins (Cassandra LWT semantics are not needed here).
	kindSimple migrationKind = "simple"

	// kindTwoKey: (key, column1, value) → (tenant_id, shard_id, key, key2, value, updated)
	// Used for GenericXconfNamedList → generic_named_lists where key2 is secondary key.
	kindTwoKey migrationKind = "two-key"
)

// tableMapping describes one old→new migration unit.
type tableMapping struct {
	oldName string
	newName string
	kind    migrationKind
}

// tableMappings is the complete list of migrations performed by this tool.
// Tables omitted here (Tag, Logs2, TagMembersBucketed, TagBucketMetadata) require
// schema-specific custom logic and are not included in this automated migration.
var tableMappings = []tableMapping{
	// Simple entity tables -------------------------------------------------------
	{"DcmRule", db.TABLE_DCM_RULES, kindSimple},
	{"DeviceSettings2", db.TABLE_DEVICE_SETTINGS, kindSimple},
	{"Environment", db.TABLE_ENVIRONMENTS, kindSimple},
	{"FeatureControlRule2", db.TABLE_FEATURE_CONTROL_RULES, kindSimple},
	{"FirmwareConfig", db.TABLE_FIRMWARE_CONFIGS, kindSimple},
	{"FirmwareRule4", db.TABLE_FIRMWARE_RULES, kindSimple},
	{"FirmwareRuleTemplate", db.TABLE_FIRMWARE_RULE_TEMPLATES, kindSimple},
	{"LogFile", db.TABLE_LOG_FILES, kindSimple},
	{"LogFileList", db.TABLE_LOG_FILE_LISTS, kindSimple},
	{"LogFilesGroups", db.TABLE_LOG_FILE_GROUPS, kindSimple},
	{"LogUploadSettings2", db.TABLE_LOG_UPLOAD_SETTINGS, kindSimple},
	{"Model", db.TABLE_MODELS, kindSimple},
	{"PermanentTelemetry", db.TABLE_PERMANENT_TELEMETRY_PROFILES, kindSimple},
	{"SettingProfiles", db.TABLE_SETTING_PROFILES, kindSimple},
	{"SettingRules", db.TABLE_SETTING_RULES, kindSimple},
	{"SingletonFilterValue", db.TABLE_SINGLETON_FILTER_VALUES, kindSimple},
	{"Telemetry", db.TABLE_TELEMETRY_PROFILES, kindSimple},
	{"TelemetryRules", db.TABLE_TELEMETRY_RULES, kindSimple},
	{"UploadRepository", db.TABLE_UPLOAD_REPOSITORIES, kindSimple},
	{"VodSettings2", db.TABLE_VOD_SETTINGS, kindSimple},
	{"XconfApprovedChange", db.TABLE_TELEMETRY_APPROVED_CHANGES, kindSimple},
	{"XconfChange", db.TABLE_TELEMETRY_CHANGES, kindSimple},
	{"XconfFeature", db.TABLE_FEATURES, kindSimple},
	{"TelemetryTwoProfiles", db.TABLE_TELEMETRY_TWO_PROFILES, kindSimple},
	{"TelemetryTwoRules", db.TABLE_TELEMETRY_TWO_RULES, kindSimple},
	{"XconfTelemetryTwoChange", db.TABLE_TELEMETRY_TWO_CHANGES, kindSimple},
	{"XconfApprovedTelemetryTwoChange", db.TABLE_TELEMETRY_APPROVED_TWO_CHANGES, kindSimple},
	{"AppSettings", db.TABLE_APP_SETTINGS, kindSimple},
	{"ApplicationTypes", db.TABLE_APPLICATION_TYPES, kindSimple},

	// Two-key table --------------------------------------------------------------
	// column1 is renamed as key2, the second clustering column in the new schema.
	{"GenericXconfNamedList", db.TABLE_GENERIC_NS_LIST, kindTwoKey},
}

func main() {
	configFile := flag.String("f", defaultConfigFile, "config file")
	tableName := flag.String("table", "", "Migrate only this old table name; omit to migrate all tables")
	clear := flag.Bool("clear", false, "TRUNCATE destination tables before migrating (allows re-running on existing data)")
	dryRun := flag.Bool("dry-run", false, "Count rows without writing to the destination")
	flag.Parse()

	// read new hocon config
	sc, err := common.NewServerConfig(*configFile)
	if err != nil {
		log.Fatalf("ERROR reading config file: %v", err)
	}

	dcc := db.DefaultCassandraConnection{Connection_type: "local"}
	dbClient, err := dcc.NewCassandraClient(sc.Config, false)
	if err != nil {
		log.Fatalf("ERROR cassandra db init error: %v", err)
	}
	defer dbClient.Close()

	// Select which tables to migrate.
	mappings := tableMappings
	if *tableName != "" {
		mappings = nil
		for _, m := range tableMappings {
			if m.oldName == *tableName {
				mappings = append(mappings, m)
			}
		}
		if len(mappings) == 0 {
			log.Fatalf("unknown old table %q\nKnown tables: %s", *tableName, knownOldNames())
		}
	}

	if *dryRun {
		log.Println("Dry run — no rows will be written.")
	}
	if *clear {
		log.Println("Clear mode — destination tables will be truncated before migration.")
	}

	// Validate all destination tables exist before starting migration
	log.Println("Validating destination tables...")
	dstKeyspace := dbClient.Keyspace
	if err := validateDestinationTables(dbClient.Session, dstKeyspace, mappings); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	log.Printf("All %d destination table(s) validated in keyspace %q.", len(mappings), dstKeyspace)

	// Optionally truncate destination tables so migration starts from a clean state.
	if *clear {
		if err := clearDestinationTables(dbClient.Session, dstKeyspace, mappings, *dryRun); err != nil {
			log.Fatalf("ERROR deleting data from destination tables: %v", err)
		}
	}

	tenantID := db.GetDefaultTenantId()
	srcKeyspace := dbClient.GetLogKeyspace() // old tables are in ApplicationsDiscoveryDataService keyspace

	var totalRows, totalErrors int
	for _, m := range mappings {
		rows, migrateErr := runMigration(dbClient.Session, m, srcKeyspace, tenantID, *dryRun)
		totalRows += rows
		if migrateErr != nil {
			log.Printf("ERROR %q → %q: %v", m.oldName, m.newName, migrateErr)
			totalErrors++
		}
	}

	action := "Migrated"
	if *dryRun {
		action = "Would migrate"
	}
	log.Printf("%s %d rows across %d table(s) with %d error(s).",
		action, totalRows, len(mappings), totalErrors)

	if totalErrors > 0 {
		os.Exit(1)
	}
}

// clearDestinationTables truncates every destination table listed in mappings so that
// the migration can be re-run on a clean database.  When dryRun is true the TRUNCATE
// statements are logged but not executed.
func clearDestinationTables(session *gocql.Session, keyspace string, mappings []tableMapping, dryRun bool) error {
	for _, m := range mappings {
		if dryRun {
			log.Printf("  dry-run: would TRUNCATE %q.%q", keyspace, m.newName)
			continue
		}
		log.Printf("  TRUNCATE %q.%q ...", keyspace, m.newName)
		stmt := fmt.Sprintf(`TRUNCATE "%s"."%s"`, keyspace, m.newName)
		if err := session.Query(stmt).Exec(); err != nil {
			return fmt.Errorf("truncate %q: %w", m.newName, err)
		}
	}
	return nil
}

// validateDestinationTables checks that all destination tables exist before migration.
// Returns error listing any missing tables. Fetches all tables in one query for efficiency.
func validateDestinationTables(session *gocql.Session, keyspace string, mappings []tableMapping) error {
	// Fetch all table names in the keyspace with a single query
	query := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?`
	iter := session.Query(query, keyspace).Iter()

	existingTables := make(map[string]bool)
	var tableName string
	for iter.Scan(&tableName) {
		existingTables[tableName] = true
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("error fetching tables from keyspace %q: %w", keyspace, err)
	}

	// Check each destination table against the fetched set
	var missingTables []string
	for _, m := range mappings {
		if !existingTables[m.newName] {
			missingTables = append(missingTables, m.newName)
		}
	}

	if len(missingTables) > 0 {
		return fmt.Errorf("missing destination table(s): %s\nRun db_create_tables_xconf.cql first", strings.Join(missingTables, ", "))
	}
	return nil
}

// runMigration dispatches to the correct handler and returns the row count.
func runMigration(session *gocql.Session, m tableMapping, srcKeyspace, tenantID string, dryRun bool) (int, error) {
	switch m.kind {
	case kindSimple:
		return migrateSimple(session, m.oldName, m.newName, srcKeyspace, tenantID, dryRun)
	case kindTwoKey:
		return migrateTwoKey(session, m.oldName, m.newName, srcKeyspace, tenantID, dryRun)
	default:
		return 0, fmt.Errorf("unknown migration kind %q", m.kind)
	}
}

// migrateSimple reads every (key, value) row from the legacy (key, column1, value)
// table and writes it to the new (tenant_id, shard_id, key, value, updated) table.
// When a key has multiple column1 rows, each is written to the destination; the last
// write wins (Cassandra upsert semantics ensure idempotency on retries).
func migrateSimple(session *gocql.Session, oldTable, newTable, srcKeyspace, tenantID string, dryRun bool) (int, error) {
	readStmt := fmt.Sprintf(`SELECT key, value FROM "%s"."%s"`, srcKeyspace, oldTable)
	writeStmt := fmt.Sprintf(`INSERT INTO "%s" (tenant_id, shard_id, key, value, updated) VALUES (?, ?, ?, ?, ?)`, newTable)
	now := time.Now()

	iter := session.Query(readStmt).Iter()
	count, writeErrors := 0, 0

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		key, _ := row["key"].(string)
		if key == "" {
			continue
		}
		value, _ := row["value"].([]byte)
		count++

		if dryRun {
			log.Printf("    dry-run: would write key=%q to %q (tenant_id=%s, shard_id=%d)", key, newTable, tenantID, db.GetShardId(key))
		} else {
			if err := session.Query(writeStmt, tenantID, db.GetShardId(key), key, value, now).Exec(); err != nil {
				log.Printf("    warn: write key=%q to %q: %v", key, newTable, err)
				writeErrors++
			}
		}
	}
	if err := iter.Close(); err != nil {
		return count, fmt.Errorf("read %q: %w", oldTable, err)
	}

	logResult(dryRun, oldTable, newTable, count, writeErrors, "")
	return count, nil
}

// migrateTwoKey reads (key, column1, value) rows and writes them with the full
// (tenant_id, shard_id, key, key2, value, updated) primary key, preserving
// the column1 clustering column.  Used for GenericXconfNamedList → generic_named_lists.
func migrateTwoKey(session *gocql.Session, oldTable, newTable, srcKeyspace, tenantID string, dryRun bool) (int, error) {
	readStmt := fmt.Sprintf(`SELECT key, column1, value FROM "%s"."%s"`, srcKeyspace, oldTable)
	writeStmt := fmt.Sprintf(`INSERT INTO "%s" (tenant_id, shard_id, key, key2, value, updated) VALUES (?, ?, ?, ?, ?, ?)`, newTable)
	now := time.Now()

	iter := session.Query(readStmt).Iter()
	count, writeErrors := 0, 0

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		key, _ := row["key"].(string)
		if key == "" {
			continue
		}
		key2, _ := row["column1"].(string)
		value, _ := row["value"].([]byte)
		count++

		if dryRun {
			log.Printf("    dry-run: would write key=%q key2=%q to %q (tenant_id=%s, shard_id=%d)", key, key2, newTable, tenantID, db.GetShardId(key))
		} else {
			if err := session.Query(writeStmt, tenantID, db.GetShardId(key), key, key2, value, now).Exec(); err != nil {
				log.Printf("    warn: write key=%q key2=%q to %q: %v", key, key2, newTable, err)
				writeErrors++
			}
		}
	}
	if err := iter.Close(); err != nil {
		return count, fmt.Errorf("read %q: %w", oldTable, err)
	}

	logResult(dryRun, oldTable, newTable, count, writeErrors, "")
	return count, nil
}

// logResult prints a one-line summary for a completed table migration.
func logResult(dryRun bool, oldTable, newTable string, count, writeErrors int, extra string) {
	verb := "migrated"
	if dryRun {
		verb = "would migrate"
	}
	log.Printf("  %q → %q: %s %d rows, %d write errors%s",
		oldTable, newTable, verb, count, writeErrors, extra)
}

// knownOldNames returns a comma-separated string of all old table names.
func knownOldNames() string {
	names := make([]string, len(tableMappings))
	for i, m := range tableMappings {
		names[i] = m.oldName
	}
	return strings.Join(names, ", ")
}
