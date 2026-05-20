package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/util"
)

const (
	PenetrationMetricsTable           = "PenetrationMetrics"
	PenetrationDataTable              = "penetration_data"
	TenantIdColumnName                = "tenant_id"
	EstbMacColumnName                 = "estb_mac"
	EcmMacColumnName                  = "ecm_mac"
	SerialNumberColumnName            = "serial_number"
	PartnerColumnName                 = "partner"
	ModelColumnName                   = "model"
	FwFilenameColumnName              = "fw_filename"
	FwVersionColumnName               = "fw_version"
	FwReportedVersionColumnName       = "fw_reported_version"
	FwAdditionalVersionInfoColumnName = "fw_additional_version_info"
	FwAppliedRuleColumnName           = "fw_applied_rule"
	FwTsColumnName                    = "fw_ts"
	ClientCertExpiryValue             = "client_cert_expiry"
	RecoveryCertExpiryValue           = "recovery_cert_expiry"
	RfcAppliedRulesColumnName         = "rfc_applied_rules"
	RfcAccountMgmtColumnName          = "rfc_account_mgmt"
	RfcFeaturesColumnName             = "rfc_features"
	RfcPartnerColumnName              = "rfc_partner"
	TitanPartnerColumnName            = "titan_partner"
	RfcModelColumnName                = "rfc_model"
	RfcFwReportedVersionColumnName    = "rfc_fw_reported_version"
	RfcAccountHashColumnName          = "rfc_account_hash"
	RfcAccountIdColumnName            = "rfc_account_id"
	TitanAccountIdColumnName          = "titan_account_id"
	RfcEnvColumnName                  = "rfc_env"
	RfcApplicationTypeColumnName      = "rfc_application_type"
	RfcExperienceColumnName           = "rfc_experience"
	RfcTimeZoneColumnName             = "rfc_time_zone"
	RfcConfigsetHashColumnName        = "rfc_configsethash"
	RfcQueryParamsColumnName          = "rfc_query_params"
	RfcTagsColumnName                 = "rfc_tags"
	RfcEstbIpColumnName               = "rfc_estb_ip"
	RfcTsColumnName                   = "rfc_ts"
	RfcPostProcColumnName             = "rfc_post_proc"
)

// PenetrationData struct
type FwPenetrationData struct {
	TenantId                string
	EstbMac                 string
	Partner                 string
	Model                   string
	FwFilename              string
	FwVersion               string
	FwReportedVersion       string
	FwAdditionalVersionInfo string
	FwAppliedRule           string
	FwTs                    int64
	ClientCertExpiry        string
	RecoveryCertExpiry      string
}

type RfcPenetrationData struct {
	TenantId             string
	EstbMac              string
	EcmMac               string
	SerialNum            string
	Partner              string
	Model                string
	RfcPartner           string
	TitanPartner         string
	RfcModel             string
	RfcFwReportedVersion string
	RfcAppliedRules      string
	RfcAccountMgmt       string
	RfcFeatures          string
	RfcTs                int64
	RfcAccountHash       string
	RfcAccountId         string
	TitanAccountId       string
	RfcEnv               string
	RfcApplicationType   string
	RfcExperience        string
	RfcTimeZone          string
	RfcConfigsetHash     string
	RfcQueryParams       string
	RfcTags              string
	RfcEstbIp            string
	RfcPostProc          string
	ClientCertExpiry     string
	RecoveryCertExpiry   string
}

type SecurityTokenDeviceInfo struct {
	Partner                 string
	Model                   string
	FwFilename              string
	FwVersion               string
	FwAdditionalVersionInfo string
	FwTs                    int64
}

var emptyValueSet = util.NewSet("", "unknown", "noaccount", "novalue", "nomatch", "na", "nomodel")

func (c *CassandraClient) SetFwPenetrationData(pData *FwPenetrationData) error {
	// build the statement and avoid unnecessary fields/columns
	columns := []string{
		TenantIdColumnName,
		EstbMacColumnName,
		FwFilenameColumnName,
		FwVersionColumnName,
		FwReportedVersionColumnName,
		FwAdditionalVersionInfoColumnName,
		FwAppliedRuleColumnName,
		FwTsColumnName,
	}
	if isEmptyString(pData.FwAppliedRule) {
		pData.FwAppliedRule = ""
	}
	if isEmptyString(pData.FwFilename) {
		pData.FwFilename = ""
	}
	if isEmptyString(pData.FwVersion) {
		pData.FwVersion = ""
	}
	if isEmptyString(pData.FwReportedVersion) {
		pData.FwReportedVersion = ""
	}
	if isEmptyString(pData.FwAdditionalVersionInfo) {
		pData.FwAdditionalVersionInfo = ""
	}

	values := []any{
		pData.TenantId,
		pData.EstbMac,
		pData.FwFilename,
		pData.FwVersion,
		pData.FwReportedVersion,
		pData.FwAdditionalVersionInfo,
		pData.FwAppliedRule,
		pData.FwTs,
	}

	// XPC-18738 special handling for partner and model. We allow replacement but do not clean up if not found in input
	if !isEmptyString(pData.Partner) {
		columns = append(columns, PartnerColumnName)
		values = append(values, pData.Partner)
	}
	if !isEmptyString(pData.Model) {
		columns = append(columns, ModelColumnName)
		values = append(values, pData.Model)
	}
	if !isEmptyString(pData.ClientCertExpiry) {
		columns = append(columns, ClientCertExpiryValue)
		values = append(values, pData.ClientCertExpiry)
	}
	if !isEmptyString(pData.RecoveryCertExpiry) {
		columns = append(columns, RecoveryCertExpiryValue)
		values = append(values, pData.RecoveryCertExpiry)
	}

	return c.updatePenetrationData(columns, values)
}

func (c *CassandraClient) SetRfcPenetrationData(pData *RfcPenetrationData, isReturn304FromPrecook bool) error {
	// always write rfc_* values so pre-cook data is as close to what we're using in rule evaluation
	columns := []string{
		TenantIdColumnName,
		EstbMacColumnName,
		EcmMacColumnName,
		SerialNumberColumnName,
		RfcPartnerColumnName,
		TitanPartnerColumnName,
		RfcModelColumnName,
		RfcAccountHashColumnName,
		RfcAccountIdColumnName,
		RfcAccountMgmtColumnName,
		RfcFwReportedVersionColumnName,
		RfcEnvColumnName,
		RfcApplicationTypeColumnName,
		RfcExperienceColumnName,
		RfcTimeZoneColumnName,
		RfcConfigsetHashColumnName,
		RfcQueryParamsColumnName,
		RfcTagsColumnName,
		RfcEstbIpColumnName,
		RfcTsColumnName,
		RfcPostProcColumnName,
	}

	values := []any{
		pData.TenantId,
		pData.EstbMac,
		pData.EcmMac,
		pData.SerialNum,
		pData.RfcPartner,
		pData.TitanPartner,
		pData.RfcModel,
		pData.RfcAccountHash,
		pData.RfcAccountId,
		pData.RfcAccountMgmt,
		pData.RfcFwReportedVersion,
		pData.RfcEnv,
		pData.RfcApplicationType,
		pData.RfcExperience,
		pData.RfcTimeZone,
		pData.RfcConfigsetHash,
		pData.RfcQueryParams,
		pData.RfcTags,
		pData.RfcEstbIp,
		pData.RfcTs,
		pData.RfcPostProc,
	}

	// only write following values when they're non-empty for rfc penetratioin metrics
	if !isEmptyString(pData.Partner) {
		columns = append(columns, PartnerColumnName)
		values = append(values, pData.Partner)
	}
	if !isEmptyString(pData.Model) {
		columns = append(columns, ModelColumnName)
		values = append(values, pData.Model)
	}
	if !isEmptyString(pData.ClientCertExpiry) {
		columns = append(columns, ClientCertExpiryValue)
		values = append(values, pData.ClientCertExpiry)
	}
	if !isEmptyString(pData.RecoveryCertExpiry) {
		columns = append(columns, RecoveryCertExpiryValue)
		values = append(values, pData.RecoveryCertExpiry)
	}
	if isEmptyString(pData.RfcAppliedRules) {
		pData.RfcAppliedRules = ""
	}
	if isEmptyString(pData.RfcFeatures) {
		pData.RfcFeatures = ""
	}
	if !isEmptyString(pData.TitanAccountId) {
		columns = append(columns, TitanAccountIdColumnName)
		values = append(values, pData.TitanAccountId)
	}

	//if we return 304 based on precook data, we do not update features and applied_rules with empty string
	if !isReturn304FromPrecook {
		columns = append(columns, RfcFeaturesColumnName, RfcAppliedRulesColumnName)
		values = append(values, pData.RfcFeatures, pData.RfcAppliedRules)
	}

	return c.updatePenetrationData(columns, values)
}

func (c *CassandraClient) SetPenetrationData(kvmap map[string]string) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	columns := []string{}
	values := []any{}

	for k, v := range kvmap {
		columns = append(columns, k)
		values = append(values, v)
	}

	return c.updatePenetrationData(columns, values)
}

func (c *CassandraClient) GetPenetrationData(estbMac string) (map[string]any, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	tableName := PenetrationDataTable
	if IsDualWriteEnabled() {
		// When dual write is enabled, read from old PenetrationMetrics table for backward compatibility,
		// until penetration_data table is fully migrated
		tableName = c.getTableNameFromLogKeyspace(PenetrationMetricsTable)
	}

	dict := util.Dict{}
	stmt := fmt.Sprintf("SELECT * FROM %s WHERE %s=?", tableName, EstbMacColumnName)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)

	if err != nil {
		return dict, err
	}

	return dict, nil
}

func (c *CassandraClient) GetFwPenetrationData(estbMac string) (*FwPenetrationData, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	pData := &FwPenetrationData{}
	dict := util.Dict{}
	tableName := PenetrationDataTable
	if IsDualWriteEnabled() {
		// When dual write is enabled, read from old PenetrationMetrics table for backward compatibility,
		// until penetration_data table is fully migrated
		tableName = c.getTableNameFromLogKeyspace(PenetrationMetricsTable)
	}

	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE %s=?`, tableName, EstbMacColumnName)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)

	if err != nil {
		return pData, err
	}

	for k, v := range dict {
		switch k {
		case EstbMacColumnName:
			if itfvalue, ok := v.(string); ok {
				// NOTE we choose to interpret an empty string as a null string
				if len(itfvalue) > 0 {
					pData.EstbMac = itfvalue
				}
			}
		case PartnerColumnName:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pData.Partner = itfvalue
				}
			}
		case ModelColumnName:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pData.Model = itfvalue
				}
			}
		case FwFilenameColumnName:
			if itfvalue, ok := v.(string); ok {
				pData.FwFilename = itfvalue
			}
		case FwVersionColumnName:
			if itfvalue, ok := v.(string); ok {
				pData.FwVersion = itfvalue
			}
		case FwReportedVersionColumnName:
			if itfvalue, ok := v.(string); ok {
				pData.FwReportedVersion = itfvalue
			}
		case FwAdditionalVersionInfoColumnName:
			if itfvalue, ok := v.(string); ok {
				pData.FwAdditionalVersionInfo = itfvalue
			}
		case FwAppliedRuleColumnName:
			if itfvalue, ok := v.(string); ok {
				pData.FwAppliedRule = itfvalue
			}
		case FwTsColumnName:
			if itfvalue, ok := v.(time.Time); ok {
				pData.FwTs = itfvalue.Unix()
			} else if itfvalue, ok := v.(int64); ok {
				// fallback for existing int64 values
				pData.FwTs = itfvalue
			}
		}
	}

	return pData, nil
}

func (c *CassandraClient) GetRfcPenetrationData(estbMac string) (*RfcPenetrationData, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	pData := &RfcPenetrationData{}
	dict := util.Dict{}
	tableName := PenetrationDataTable
	if IsDualWriteEnabled() {
		// When dual write is enabled, read from old PenetrationMetrics table for backward compatibility,
		// until penetration_data table is fully migrated
		tableName = c.getTableNameFromLogKeyspace(PenetrationMetricsTable)
	}

	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE %s=?`, tableName, EstbMacColumnName)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)

	if err != nil {
		return pData, err
	}

	for k, v := range dict {
		switch k {
		case EstbMacColumnName:
			if itfvalue, ok := v.(string); ok {
				// NOTE we choose to interpret an empty string as a null string
				if len(itfvalue) > 0 {
					pData.EstbMac = itfvalue
				}
			}
		case PartnerColumnName:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pData.Partner = itfvalue
				}
			}
		case ModelColumnName:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pData.Model = itfvalue
				}
			}
		case RfcAppliedRulesColumnName:
			if itfvalue, ok := v.(string); ok {
				pData.RfcAppliedRules = itfvalue
			}
		case RfcFeaturesColumnName:
			if itfvalue, ok := v.(string); ok {
				pData.RfcFeatures = itfvalue
			}
		case RfcTsColumnName:
			if itfvalue, ok := v.(time.Time); ok {
				pData.RfcTs = itfvalue.Unix()
			} else if itfvalue, ok := v.(int64); ok {
				// fallback for existing int64 values
				pData.RfcTs = itfvalue
			}
		}
	}

	return pData, nil
}

func (c *CassandraClient) updatePenetrationData(columns []string, values []any) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	if len(columns) == 0 || len(values) == 0 || len(columns) != len(values) {
		return fmt.Errorf("invalid input for updating penetration data, columns: %v, values: %v", columns, values)
	}

	tables := []string{PenetrationDataTable}
	if IsDualWriteEnabled() {
		// Write to PenetrationMetrics table for backward compatibility, but PenetrationMetrics will be eventually removed
		tables = append(tables, c.getTableNameFromLogKeyspace(PenetrationMetricsTable))
	}
	for _, tableName := range tables {
		stmt := fmt.Sprintf(`INSERT INTO %s(%v) VALUES(%v)`, tableName, GetColumnsStr(columns), GetValuesStr(len(columns)))
		if err := c.Query(stmt, values...).Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CassandraClient) GetEstbIp(estbMac string) (string, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var estbIp string
	dict := util.Dict{}
	tableName := PenetrationDataTable
	if IsDualWriteEnabled() {
		// When dual write is enabled, read from old PenetrationMetrics table for backward compatibility,
		// until penetration_data table is fully migrated
		tableName = c.getTableNameFromLogKeyspace(PenetrationMetricsTable)
	}

	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE %s=?`, tableName, EstbMacColumnName)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)
	if err != nil {
		return estbIp, err
	}

	if itf, ok := dict["estb_ip"]; ok {
		estbIp = itf.(string)
	}

	return estbIp, nil
}

func (c *CassandraClient) GetSecurityTokenFields(estbMac string) (*SecurityTokenDeviceInfo, error) {
	securityTokenDeviceInfo := &SecurityTokenDeviceInfo{}
	dict := util.Dict{}
	columns := []string{
		PartnerColumnName,
		ModelColumnName,
		FwFilenameColumnName,
		FwVersionColumnName,
		FwAdditionalVersionInfoColumnName,
		FwTsColumnName,
	}
	tableName := PenetrationDataTable
	if IsDualWriteEnabled() {
		// When dual write is enabled, read from old PenetrationMetrics table for backward compatibility,
		// until penetration_data table is fully migrated
		tableName = c.getTableNameFromLogKeyspace(PenetrationMetricsTable)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s=?`, GetColumnsStr(columns), tableName, EstbMacColumnName)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)
	if err != nil {
		return securityTokenDeviceInfo, err
	}

	// Extract values from dict into struct
	if v, ok := dict[PartnerColumnName].(string); ok {
		securityTokenDeviceInfo.Partner = v
	}
	if v, ok := dict[ModelColumnName].(string); ok {
		securityTokenDeviceInfo.Model = v
	}
	if v, ok := dict[FwFilenameColumnName].(string); ok {
		securityTokenDeviceInfo.FwFilename = v
	}
	if v, ok := dict[FwVersionColumnName].(string); ok {
		securityTokenDeviceInfo.FwVersion = v
	}
	if v, ok := dict[FwAdditionalVersionInfoColumnName].(string); ok {
		securityTokenDeviceInfo.FwAdditionalVersionInfo = v
	}
	if v, ok := dict[FwTsColumnName].(time.Time); ok {
		securityTokenDeviceInfo.FwTs = v.Unix()
	} else if itfvalue, ok := dict[FwTsColumnName].(int64); ok {
		// fallback for existing int64 values
		securityTokenDeviceInfo.FwTs = itfvalue
	}

	return securityTokenDeviceInfo, nil
}

func isEmptyString(str string) bool {
	str = strings.TrimSpace(strings.ToLower(str))
	return emptyValueSet.Contains(str)
}
