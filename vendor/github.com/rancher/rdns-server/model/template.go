package model

var CoreFileTmpl = `
. {
    {{- if and .CoreDNSDBFile .CoreDNSDBZone}}
    file {{.CoreDNSDBFile}} {{.CoreDNSDBZone}} {
        reload 0
    }
    {{- end}}
    rdns {{.Domain}} {
        path {{.EtcdPrefixPath}}
        endpoint {{.EtcdEndpoints}}
        upstream 8.8.8.8:53 8.8.4.4:53
        wildcardbound {{.WildCardBound}}
    }
    cache {{.TTL}} {{.Domain}}
    loadbalance
    forward . 8.8.8.8:53 8.8.4.4:53
    log stdout
    errors
}`

type CoreFile struct {
	CoreDNSDBFile  string
	CoreDNSDBZone  string
	Domain         string
	EtcdPrefixPath string
	EtcdEndpoints  string
	TTL            string
	WildCardBound  string
}
