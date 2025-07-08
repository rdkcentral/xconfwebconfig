package db

import (
	"fmt"
	"strings"

	"xconfwebconfig/util"
)

const (
	PenetrationMetricsTable            = "PenetrationMetrics"
	EstbMacColumnValue                 = "estb_mac"
	EcmMacColumnValue                  = "ecm_mac"
	SerialNumberColumnValue            = "serial_number"
	PartnerColumnValue                 = "partner"
	ModelColumnValue                   = "model"
	FwVersionColumnValue               = "fw_version"
	FwReportedVersionColumnValue       = "fw_reported_version"
	FwAdditionalVersionInfoColumnValue = "fw_additional_version_info"
	FwAppliedRuleColumnValue           = "fw_applied_rule"
	FwTsColumnValue                    = "fw_ts"
	ClientCertExpiryValue              = "client_cert_expiry"
	RecoveryCertExpiryValue            = "recovery_cert_expiry"
	RfcAppliedRulesColumnValue         = "rfc_applied_rules"
	RfcAccountMgmtColumnValue          = "rfc_account_mgmt"
	RfcFeaturesColumnValue             = "rfc_features"
	RfcPartnerColumnValue              = "rfc_partner"
	TitanPartnerColumnValue            = "titan_partner"
	RfcModelColumnValue                = "rfc_model"
	RfcFwReportedVersionColumnValue    = "rfc_fw_reported_version"
	RfcAccountHashColumnValue          = "rfc_account_hash"
	RfcAccountIdColumnValue            = "rfc_account_id"
	TitanAccountIdColumnValue          = "titan_account_id"
	RfcEnvColumnValue                  = "rfc_env"
	RfcApplicationTypeColumnValue      = "rfc_application_type"
	RfcExperienceColumnValue           = "rfc_experience"
	RfcTimeZoneColumnValue             = "rfc_time_zone"
	RfcConfigsetHashColumnValue        = "rfc_configsethash"
	RfcQueryParamsColumnValue          = "rfc_query_params"
	RfcTagsColumnValue                 = "rfc_tags"
	RfcEstbIpColumnValue               = "rfc_estb_ip"
	RfcTsColumnValue                   = "rfc_ts"
	RfcPostProcColumnValue             = "rfc_post_proc"
)

// PenetrationMetrics struct
type FwPenetrationMetrics struct {
	EstbMac                 string
	Partner                 string
	Model                   string
	FwVersion               string
	FwReportedVersion       string
	FwAdditionalVersionInfo string
	FwAppliedRule           string
	FwTs                    int64
	ClientCertExpiry        string
	RecoveryCertExpiry      string
}

type RfcPenetrationMetrics struct {
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

var emptyValueSet = util.NewSet("", "unknown", "noaccount", "novalue", "nomatch", "na", "nomodel")

func (c *CassandraClient) SetFwPenetrationMetrics(pMetrics *FwPenetrationMetrics) error {
	// build the statement and avoid unnecessary fields/columns

	columns := []string{
		EstbMacColumnValue,
		FwVersionColumnValue,
		FwReportedVersionColumnValue,
		FwAdditionalVersionInfoColumnValue,
		FwAppliedRuleColumnValue,
		FwTsColumnValue,
	}
	if isEmptyString(pMetrics.FwAppliedRule) {
		pMetrics.FwAppliedRule = ""
	}
	if isEmptyString(pMetrics.FwVersion) {
		pMetrics.FwVersion = ""
	}
	if isEmptyString(pMetrics.FwReportedVersion) {
		pMetrics.FwReportedVersion = ""
	}
	if isEmptyString(pMetrics.FwAdditionalVersionInfo) {
		pMetrics.FwAdditionalVersionInfo = ""
	}

	values := []interface{}{
		pMetrics.EstbMac,
		pMetrics.FwVersion,
		pMetrics.FwReportedVersion,
		pMetrics.FwAdditionalVersionInfo,
		pMetrics.FwAppliedRule,
		pMetrics.FwTs,
	}

	// XPC-18738 special handling for partner and model. We allow replacement but do not clean up if not found in input
	if !isEmptyString(pMetrics.Partner) {
		columns = append(columns, PartnerColumnValue)
		values = append(values, pMetrics.Partner)
	}
	if !isEmptyString(pMetrics.Model) {
		columns = append(columns, ModelColumnValue)
		values = append(values, pMetrics.Model)
	}

	if !isEmptyString(pMetrics.ClientCertExpiry) {
		columns = append(columns, ClientCertExpiryValue)
		values = append(values, pMetrics.ClientCertExpiry)
	}

	if !isEmptyString(pMetrics.RecoveryCertExpiry) {
		columns = append(columns, RecoveryCertExpiryValue)
		values = append(values, pMetrics.RecoveryCertExpiry)
	}

	stmt := fmt.Sprintf(`INSERT INTO "%s"(%v) VALUES(%v)`, PenetrationMetricsTable, GetColumnsStr(columns), GetValuesStr(len(columns)))

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()
	qry := c.Query(stmt, values...)
	err := qry.Exec()

	if err != nil {
		return err
	}
	return nil
}

func (c *CassandraClient) SetRfcPenetrationMetrics(pMetrics *RfcPenetrationMetrics) error {
	// always write rfc_* values so pre-cook data is as close to what we're using in rule evaluation
	columns := []string{
		EstbMacColumnValue,
		EcmMacColumnValue,
		SerialNumberColumnValue,
		RfcPartnerColumnValue,
		TitanPartnerColumnValue,
		RfcModelColumnValue,
		RfcAccountHashColumnValue,
		RfcAccountIdColumnValue,
		RfcAccountMgmtColumnValue,
		TitanAccountIdColumnValue,
		RfcFwReportedVersionColumnValue,
		RfcFeaturesColumnValue,
		RfcAppliedRulesColumnValue,
		RfcEnvColumnValue,
		RfcApplicationTypeColumnValue,
		RfcExperienceColumnValue,
		RfcTimeZoneColumnValue,
		RfcConfigsetHashColumnValue,
		RfcQueryParamsColumnValue,
		RfcTagsColumnValue,
		RfcEstbIpColumnValue,
		RfcTsColumnValue,
		RfcPostProcColumnValue,
	}

	if isEmptyString(pMetrics.RfcAppliedRules) {
		pMetrics.RfcAppliedRules = ""
	}
	if isEmptyString(pMetrics.RfcFeatures) {
		pMetrics.RfcFeatures = ""
	}

	values := []interface{}{
		pMetrics.EstbMac,
		pMetrics.EcmMac,
		pMetrics.SerialNum,
		pMetrics.RfcPartner,
		pMetrics.TitanPartner,
		pMetrics.RfcModel,
		pMetrics.RfcAccountHash,
		pMetrics.RfcAccountId,
		pMetrics.RfcAccountMgmt,
		pMetrics.TitanAccountId,
		pMetrics.RfcFwReportedVersion,
		pMetrics.RfcFeatures,
		pMetrics.RfcAppliedRules,
		pMetrics.RfcEnv,
		pMetrics.RfcApplicationType,
		pMetrics.RfcExperience,
		pMetrics.RfcTimeZone,
		pMetrics.RfcConfigsetHash,
		pMetrics.RfcQueryParams,
		pMetrics.RfcTags,
		pMetrics.RfcEstbIp,
		pMetrics.RfcTs,
		pMetrics.RfcPostProc,
	}

	// only write following values when they're non-empty for metrics
	if !isEmptyString(pMetrics.Partner) {
		columns = append(columns, PartnerColumnValue)
		values = append(values, pMetrics.Partner)
	}
	if !isEmptyString(pMetrics.Model) {
		columns = append(columns, ModelColumnValue)
		values = append(values, pMetrics.Model)
	}
	if !isEmptyString(pMetrics.ClientCertExpiry) {
		columns = append(columns, ClientCertExpiryValue)
		values = append(values, pMetrics.ClientCertExpiry)
	}

	if !isEmptyString(pMetrics.RecoveryCertExpiry) {
		columns = append(columns, RecoveryCertExpiryValue)
		values = append(values, pMetrics.RecoveryCertExpiry)
	}

	stmt := fmt.Sprintf(`INSERT INTO "%s"(%v) VALUES(%v)`, PenetrationMetricsTable, GetColumnsStr(columns), GetValuesStr(len(columns)))

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()
	qry := c.Query(stmt, values...)
	err := qry.Exec()

	if err != nil {
		return err
	}
	return nil
}

func (c *CassandraClient) GetFwPenetrationMetrics(estbMac string) (*FwPenetrationMetrics, error) {
	pMetrics := &FwPenetrationMetrics{}
	dict := util.Dict{}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()
	stmt := fmt.Sprintf(`SELECT * FROM "%s" WHERE %s=?`, PenetrationMetricsTable, EstbMacColumnValue)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)

	if err != nil {
		return pMetrics, err
	}

	for k, v := range dict {
		switch k {
		case EstbMacColumnValue:
			if itfvalue, ok := v.(string); ok {
				// NOTE we choose to interpret an empty string as a null string
				if len(itfvalue) > 0 {
					pMetrics.EstbMac = itfvalue
				}
			}
		case PartnerColumnValue:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pMetrics.Partner = itfvalue
				}
			}
		case ModelColumnValue:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pMetrics.Model = itfvalue
				}
			}
		case FwVersionColumnValue:
			if itfvalue, ok := v.(string); ok {
				pMetrics.FwVersion = itfvalue
			}
		case FwReportedVersionColumnValue:
			if itfvalue, ok := v.(string); ok {
				pMetrics.FwReportedVersion = itfvalue
			}
		case FwAdditionalVersionInfoColumnValue:
			if itfvalue, ok := v.(string); ok {
				pMetrics.FwAdditionalVersionInfo = itfvalue
			}
		case FwAppliedRuleColumnValue:
			if itfvalue, ok := v.(string); ok {
				pMetrics.FwAppliedRule = itfvalue
			}
		case FwTsColumnValue:
			if itfvalue, ok := v.(int64); ok {
				pMetrics.FwTs = itfvalue
			}
		}
	}

	return pMetrics, nil
}

func (c *CassandraClient) GetRfcPenetrationMetrics(estbMac string) (*RfcPenetrationMetrics, error) {
	pMetrics := &RfcPenetrationMetrics{}
	dict := util.Dict{}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()
	stmt := fmt.Sprintf(`SELECT * FROM "%s" WHERE %s=?`, PenetrationMetricsTable, EstbMacColumnValue)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)

	if err != nil {
		return pMetrics, err
	}

	for k, v := range dict {
		switch k {
		case EstbMacColumnValue:
			if itfvalue, ok := v.(string); ok {
				// NOTE we choose to interpret an empty string as a null string
				if len(itfvalue) > 0 {
					pMetrics.EstbMac = itfvalue
				}
			}
		case PartnerColumnValue:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pMetrics.Partner = itfvalue
				}
			}
		case ModelColumnValue:
			if itfvalue, ok := v.(string); ok {
				if len(itfvalue) > 0 {
					pMetrics.Model = itfvalue
				}
			}
		case RfcAppliedRulesColumnValue:
			if itfvalue, ok := v.(string); ok {
				pMetrics.RfcAppliedRules = itfvalue
			}
		case RfcFeaturesColumnValue:
			if itfvalue, ok := v.(string); ok {
				pMetrics.RfcFeatures = itfvalue
			}
		case RfcTsColumnValue:
			if itfvalue, ok := v.(int64); ok {
				pMetrics.RfcTs = itfvalue
			}
		}
	}

	return pMetrics, nil
}

func isEmptyString(str string) bool {
	str = strings.TrimSpace(strings.ToLower(str))
	return emptyValueSet.Contains(str)
}

func (c *CassandraClient) UpdateFwPenetrationMetrics(kvmap map[string]string) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	columns := []string{}
	values := []interface{}{}

	for k, v := range kvmap {
		columns = append(columns, k)
		values = append(values, v)
	}

	stmt := fmt.Sprintf(`INSERT INTO "%v"(%v) VALUES(%v)`, PenetrationMetricsTable, GetColumnsStr(columns), GetValuesStr(len(columns)))
	if err := c.Query(stmt, values...).Exec(); err != nil {
		return err
	}
	return nil
}

func (c *CassandraClient) GetEstbIp(estbMac string) (string, error) {
	dict := util.Dict{}
	var estbIp string

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()
	stmt := fmt.Sprintf(`SELECT * FROM "%s" WHERE %s=?`, PenetrationMetricsTable, EstbMacColumnValue)
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
