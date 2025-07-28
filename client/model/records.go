package model

type Record struct {
	IP         string `json:"ip"`
	DomainID   string `json:"domain_id"`
	RecordType string `json:"record_type"`
	Timestamp  int64  `json:"timestamp"`
}

type ARecord struct {
	DomainID  string `db:"domain_id" json:"domain_id" cql:"domain_id"`
	IP        string `db:"ip" json:"ip" cql:"ip"`
	ASN       int    `db:"asn" json:"asn" cql:"asn"`
	ASNName   string `db:"asn_name" json:"asn_name" cql:"asn_name"`
	Country   string `db:"country" json:"country" cql:"country"`
	City      string `db:"city" json:"city" cql:"city"`
	LatLong   string `db:"latlong" json:"latlong" cql:"latlong"`
	Timestamp int64  `db:"timestamp" json:"timestamp" cql:"timestamp"`
}

func (r ARecord) GetDomainID() string {
	return r.DomainID
}

type AAAARecord struct {
	DomainID  string `db:"domain_id" json:"domain_id" cql:"domain_id"`
	IP        string `db:"ip" json:"ip" cql:"ip"`
	ASN       int    `db:"asn" json:"asn" cql:"asn"`
	ASNName   string `db:"asn_name" json:"asn_name" cql:"asn_name"`
	Country   string `db:"country" json:"country" cql:"country"`
	City      string `db:"city" json:"city" cql:"city"`
	LatLong   string `db:"latlong" json:"latlong" cql:"latlong"`
	Timestamp int64  `db:"timestamp" json:"timestamp" cql:"timestamp"`
}

func (r AAAARecord) GetDomainID() string {
	return r.DomainID
}
