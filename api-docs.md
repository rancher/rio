<p>Packages:</p>
<ul>
<li>
<a href="#autoscale.rio.cattle.io">autoscale.rio.cattle.io</a>
</li>
<li>
<a href="#git.rio.cattle.io">git.rio.cattle.io</a>
</li>
<li>
<a href="#project.rio.cattle.io">project.rio.cattle.io</a>
</li>
<li>
<a href="#rio.cattle.io">rio.cattle.io</a>
</li>
<li>
<a href="#webhookinator.rio.cattle.io">webhookinator.rio.cattle.io</a>
</li>
</ul>
<h2 id="autoscale.rio.cattle.io">autoscale.rio.cattle.io</h2>
<p>
</p>
Resource Types:
<ul><li>
<a href="#ServiceScaleRecommendation">ServiceScaleRecommendation</a>
</li></ul>
<h3 id="ServiceScaleRecommendation">ServiceScaleRecommendation
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
autoscale.rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>ServiceScaleRecommendation</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#ServiceScaleRecommendationSpec">
ServiceScaleRecommendationSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>serviceNameToRead</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>zeroScaleService</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>minScale</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>maxScale</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>concurrency</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>prometheusURL</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>selector</code></br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#ServiceScaleRecommendationStatus">
ServiceScaleRecommendationStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceScaleRecommendationSpec">ServiceScaleRecommendationSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceScaleRecommendation">ServiceScaleRecommendation</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>serviceNameToRead</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>zeroScaleService</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>minScale</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>maxScale</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>concurrency</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>prometheusURL</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>selector</code></br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceScaleRecommendationStatus">ServiceScaleRecommendationStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceScaleRecommendation">ServiceScaleRecommendation</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>desiredScale</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<hr/>
<h2 id="git.rio.cattle.io">git.rio.cattle.io</h2>
<p>
</p>
Resource Types:
<ul><li>
<a href="#GitModule">GitModule</a>
</li></ul>
<h3 id="GitModule">GitModule
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
git.rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>GitModule</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#GitModuleSpec">
GitModuleSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>serviceName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serviceNamespace</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repo</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>secret</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>branch</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#GitModuleStatus">
GitModuleStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="GitModuleSpec">GitModuleSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#GitModule">GitModule</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>serviceName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serviceNamespace</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repo</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>secret</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>branch</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="GitModuleStatus">GitModuleStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#GitModule">GitModule</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>lastRevision</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<hr/>
<h2 id="project.rio.cattle.io">project.rio.cattle.io</h2>
<p>
</p>
Resource Types:
<ul><li>
<a href="#ClusterDomain">ClusterDomain</a>
</li><li>
<a href="#Feature">Feature</a>
</li></ul>
<h3 id="ClusterDomain">ClusterDomain
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
project.rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>ClusterDomain</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#ClusterDomainSpec">
ClusterDomainSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>SecretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>addresses</code></br>
<em>
<a href="#Address">
[]Address
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>subdomains</code></br>
<em>
<a href="#Subdomain">
[]Subdomain
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#ClusterDomainStatus">
ClusterDomainStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Feature">Feature
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
project.rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Feature</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#FeatureSpec">
FeatureSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>description</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>enable</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>questions</code></br>
<em>
<a href="#Question">
[]Question
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>answers</code></br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>features</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#FeatureStatus">
FeatureStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Address">Address
</h3>
<p>
(<em>Appears on:</em>
<a href="#ClusterDomainSpec">ClusterDomainSpec</a>, 
<a href="#Subdomain">Subdomain</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ip</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ClusterDomainSpec">ClusterDomainSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#ClusterDomain">ClusterDomain</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>SecretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>addresses</code></br>
<em>
<a href="#Address">
[]Address
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>subdomains</code></br>
<em>
<a href="#Subdomain">
[]Subdomain
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ClusterDomainStatus">ClusterDomainStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#ClusterDomain">ClusterDomain</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>httpsSupported</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>domain</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="FeatureSpec">FeatureSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#Feature">Feature</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>description</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>enable</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>questions</code></br>
<em>
<a href="#Question">
[]Question
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>answers</code></br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>features</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="FeatureStatus">FeatureStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#Feature">Feature</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>enableOverride</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Subdomain">Subdomain
</h3>
<p>
(<em>Appears on:</em>
<a href="#ClusterDomainSpec">ClusterDomainSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>addresses</code></br>
<em>
<a href="#Address">
[]Address
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<hr/>
<h2 id="rio.cattle.io">rio.cattle.io</h2>
<p>
</p>
Resource Types:
<ul><li>
<a href="#App">App</a>
</li><li>
<a href="#ExternalService">ExternalService</a>
</li><li>
<a href="#PublicDomain">PublicDomain</a>
</li><li>
<a href="#Router">Router</a>
</li><li>
<a href="#Service">Service</a>
</li></ul>
<h3 id="App">App
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>App</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#AppSpec">
AppSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>revisions</code></br>
<em>
<a href="#Revision">
[]Revision
</a>
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#AppStatus">
AppStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ExternalService">ExternalService
</h3>
<p>
ExternalService creates a DNS record and route rules for any Service outside of the cluster, can be IPs or FQDN outside the mesh
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>ExternalService</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#ExternalServiceSpec">
ExternalServiceSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>ipAddresses</code></br>
<em>
[]string
</em>
</td>
<td>
External service located outside mesh, represented by IPs
</td>
</tr>
<tr>
<td>
<code>fqdn</code></br>
<em>
string
</em>
</td>
<td>
External service located outside mesh, represented by DNS
</td>
</tr>
<tr>
<td>
<code>service</code></br>
<em>
string
</em>
</td>
<td>
In-Mesh service name in another namespace
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#ExternalServiceStatus">
ExternalServiceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="PublicDomain">PublicDomain
</h3>
<p>
PublicDomain is a top-level resource to allow user to its own public domain for the services inside cluster. It can be pointed to
Router or Service. It is user's responsibility to setup a CNAME or A record to the clusterDomain or ingress IP.
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>PublicDomain</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#PublicDomainSpec">
PublicDomainSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>spec</code> are embedded into this type.)
</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
SecretRef reference the secret that contains key and certs for TLS configuration. By default it is configured to use Letsencrypt
</td>
</tr>
<tr>
<td>
<code>disableLetsencrypt</code></br>
<em>
bool
</em>
</td>
<td>
Whether to disable Letsencrypt certificates.
</td>
</tr>
<tr>
<td>
<code>targetServiceName</code></br>
<em>
string
</em>
</td>
<td>
Target Service Name in the same Namespace
</td>
</tr>
<tr>
<td>
<code>domainName</code></br>
<em>
string
</em>
</td>
<td>
PublicDomain name
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#PublicDomainStatus">
PublicDomainStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>status</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="Router">Router
</h3>
<p>
Router is a top level resource to create L7 routing to different services. It will create VirtualService, ServiceEntry and DestinationRules
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Router</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#RouterSpec">
RouterSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>routes</code></br>
<em>
<a href="#RouteSpec">
[]RouteSpec
</a>
</em>
</td>
<td>
An ordered list of route rules for HTTP traffic. The first rule matching an incoming request is used.
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#RouterStatus">
RouterStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Service">Service
</h3>
<p>
Service acts as a top level resource for a container and its sidecarsm and routing resources.
Each service represents an individual revision, group by Spec.App(defaults to Service.Name), and Spec.Version(defaults to v0)
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Service</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>ServiceScale</code></br>
<em>
<a href="#ServiceScale">
ServiceScale
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>ServiceRevision</code></br>
<em>
<a href="#ServiceRevision">
ServiceRevision
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>AutoscaleConfig</code></br>
<em>
<a href="#AutoscaleConfig">
AutoscaleConfig
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>RolloutConfig</code></br>
<em>
<a href="#RolloutConfig">
RolloutConfig
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>PodConfig</code></br>
<em>
<a href="#PodConfig">
PodConfig
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>disableServiceMesh</code></br>
<em>
bool
</em>
</td>
<td>
Whether to disable ServiceMesh for Service. If true, no mesh sidecar will be deployed along with the Service
</td>
</tr>
<tr>
<td>
<code>permissions</code></br>
<em>
<a href="#Permission">
[]Permission
</a>
</em>
</td>
<td>
Permissions to the Services. It will create corresponding ServiceAccounts, Roles and RoleBinding.
</td>
</tr>
<tr>
<td>
<code>globalPermissions</code></br>
<em>
<a href="#Permission">
[]Permission
</a>
</em>
</td>
<td>
GlobalPermissions to the Services. It will create corresponding ServiceAccounts, ClusterRoles and ClusterRoleBinding.
</td>
</tr>
<tr>
<td>
<code>systemSpec</code></br>
<em>
<a href="#SystemServiceSpec">
SystemServiceSpec
</a>
</em>
</td>
<td>
System Field Spec
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#ServiceStatus">
ServiceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Abort">Abort
</h3>
<p>
(<em>Appears on:</em>
<a href="#Fault">Fault</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>httpStatus</code></br>
<em>
int
</em>
</td>
<td>
REQUIRED. HTTP status code to use to abort the Http request.
</td>
</tr>
</tbody>
</table>
<h3 id="AppSpec">AppSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#App">App</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>revisions</code></br>
<em>
<a href="#Revision">
[]Revision
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="AppStatus">AppStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#App">App</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>publicDomains</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>endpoints</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>revisionWeight</code></br>
<em>
<a href="#ServiceObservedWeight">
map[string]github.com/rancher/rio/pkg/apis/rio.cattle.io/v1.ServiceObservedWeight
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="AutoscaleConfig">AutoscaleConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceSpec">ServiceSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>concurrency</code></br>
<em>
int
</em>
</td>
<td>
ContainerConcurrency specifies the maximum allowed in-flight (concurrent) requests per container of the Revision. Defaults to 0 which means unlimited concurrency.
This field replaces ConcurrencyModel. A value of 1 is equivalent to Single and 0 is equivalent to Multi.
</td>
</tr>
<tr>
<td>
<code>minScale</code></br>
<em>
int
</em>
</td>
<td>
The minimal scale Service can be scaled
</td>
</tr>
<tr>
<td>
<code>maxScale</code></br>
<em>
int
</em>
</td>
<td>
The maximum scale Service can be scaled
</td>
</tr>
</tbody>
</table>
<h3 id="Container">Container
</h3>
<p>
(<em>Appears on:</em>
<a href="#NamedContainer">NamedContainer</a>, 
<a href="#PodConfig">PodConfig</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code></br>
<em>
string
</em>
</td>
<td>
Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.
</td>
</tr>
<tr>
<td>
<code>build</code></br>
<em>
<a href="#ImageBuild">
ImageBuild
</a>
</em>
</td>
<td>
ImageBuild Specify the build parameter
</td>
</tr>
<tr>
<td>
<code>command</code></br>
<em>
[]string
</em>
</td>
<td>
Entrypoint array. Not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided.
Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged.
The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
</td>
</tr>
<tr>
<td>
<code>args</code></br>
<em>
[]string
</em>
</td>
<td>
Arguments to the entrypoint. The docker image's CMD is used if this is not provided.
Variable references $(VAR_NAME) are expanded using the container's environment.
If a variable cannot be resolved, the reference in the input string will be unchanged.
The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
</td>
</tr>
<tr>
<td>
<code>workingDir</code></br>
<em>
string
</em>
</td>
<td>
Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.
</td>
</tr>
<tr>
<td>
<code>ports</code></br>
<em>
<a href="#ContainerPort">
[]ContainerPort
</a>
</em>
</td>
<td>
List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed.
Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network. Cannot be updated.
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
<a href="#EnvVar">
[]EnvVar
</a>
</em>
</td>
<td>
List of environment variables to set in the container. Cannot be updated.
</td>
</tr>
<tr>
<td>
<code>cpus</code></br>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
CPU, in cores. (500m = .5 cores)
</td>
</tr>
<tr>
<td>
<code>memory</code></br>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
</td>
</tr>
<tr>
<td>
<code>secrets</code></br>
<em>
<a href="#DataMount">
[]DataMount
</a>
</em>
</td>
<td>
Secrets Mounts
</td>
</tr>
<tr>
<td>
<code>configs</code></br>
<em>
<a href="#DataMount">
[]DataMount
</a>
</em>
</td>
<td>
Configmaps Mounts
</td>
</tr>
<tr>
<td>
<code>livenessProbe</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
</td>
</tr>
<tr>
<td>
<code>readinessProbe</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
</td>
</tr>
<tr>
<td>
<code>stdin</code></br>
<em>
bool
</em>
</td>
<td>
Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.
</td>
</tr>
<tr>
<td>
<code>stdinOnce</code></br>
<em>
bool
</em>
</td>
<td>
Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false
</td>
</tr>
<tr>
<td>
<code>tty</code></br>
<em>
bool
</em>
</td>
<td>
Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="#Volume">
[]Volume
</a>
</em>
</td>
<td>
Pod volumes to mount into the container's filesystem. Cannot be updated.
</td>
</tr>
<tr>
<td>
<code>ContainerSecurityContext</code></br>
<em>
<a href="#ContainerSecurityContext">
ContainerSecurityContext
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ContainerPort">ContainerPort
</h3>
<p>
(<em>Appears on:</em>
<a href="#Container">Container</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>internalOnly</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>protocol</code></br>
<em>
<a href="#Protocol">
Protocol
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>targetPort</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ContainerSecurityContext">ContainerSecurityContext
</h3>
<p>
(<em>Appears on:</em>
<a href="#Container">Container</a>)
</p>
<p>
ContainerSecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty. See type description for default values of each field.
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>runAsUser</code></br>
<em>
int64
</em>
</td>
<td>
The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext.
If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container
</td>
</tr>
<tr>
<td>
<code>runAsGroup</code></br>
<em>
int64
</em>
</td>
<td>
The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in SecurityContext.
If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.
</td>
</tr>
<tr>
<td>
<code>readOnlyRootFilesystem</code></br>
<em>
bool
</em>
</td>
<td>
Whether this container has a read-only root filesystem. Default is false.
</td>
</tr>
</tbody>
</table>
<h3 id="DataMount">DataMount
</h3>
<p>
(<em>Appears on:</em>
<a href="#Container">Container</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>directory</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>file</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>key</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Destination">Destination
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteTraffic">RouteTraffic</a>, 
<a href="#WeightedDestination">WeightedDestination</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>service</code></br>
<em>
string
</em>
</td>
<td>
Destination Service
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
Destination Namespace
</td>
</tr>
<tr>
<td>
<code>revision</code></br>
<em>
string
</em>
</td>
<td>
Destination Revision
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
uint32
</em>
</td>
<td>
Destination Port
</td>
</tr>
</tbody>
</table>
<h3 id="EnvVar">EnvVar
</h3>
<p>
(<em>Appears on:</em>
<a href="#Container">Container</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>value</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>secretName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>configMapName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>key</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ExternalServiceSpec">ExternalServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#ExternalService">ExternalService</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ipAddresses</code></br>
<em>
[]string
</em>
</td>
<td>
External service located outside mesh, represented by IPs
</td>
</tr>
<tr>
<td>
<code>fqdn</code></br>
<em>
string
</em>
</td>
<td>
External service located outside mesh, represented by DNS
</td>
</tr>
<tr>
<td>
<code>service</code></br>
<em>
string
</em>
</td>
<td>
In-Mesh service name in another namespace
</td>
</tr>
</tbody>
</table>
<h3 id="ExternalServiceStatus">ExternalServiceStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#ExternalService">ExternalService</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
Represents the latest available observations of a ExternalService's current state.
</td>
</tr>
</tbody>
</table>
<h3 id="Fault">Fault
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteTraffic">RouteTraffic</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>percentage</code></br>
<em>
int
</em>
</td>
<td>
Percentage of requests on which the delay will be injected.
</td>
</tr>
<tr>
<td>
<code>delayMillis</code></br>
<em>
int
</em>
</td>
<td>
REQUIRED. Add a fixed delay before forwarding the request. Units: milliseconds
</td>
</tr>
<tr>
<td>
<code>abort</code></br>
<em>
<a href="#Abort">
Abort
</a>
</em>
</td>
<td>
Abort Http request attempts and return error codes back to downstream service, giving the impression that the upstream service is faulty.
</td>
</tr>
</tbody>
</table>
<h3 id="ImageBuild">ImageBuild
</h3>
<p>
(<em>Appears on:</em>
<a href="#Container">Container</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>repo</code></br>
<em>
string
</em>
</td>
<td>
Repository url
</td>
</tr>
<tr>
<td>
<code>revision</code></br>
<em>
string
</em>
</td>
<td>
Repo Revision. Can be a git commit or tag
</td>
</tr>
<tr>
<td>
<code>branch</code></br>
<em>
string
</em>
</td>
<td>
Repo Branch. If specified, a gitmodule will be created to watch the repo and creating new revision if new commit or tag is pushed.
</td>
</tr>
<tr>
<td>
<code>stageOnly</code></br>
<em>
bool
</em>
</td>
<td>
Whether to only stage the new revision. If true, the new created service will not be allocating any traffic automatically.
</td>
</tr>
<tr>
<td>
<code>dockerFile</code></br>
<em>
string
</em>
</td>
<td>
Specify the name Of the Dockerfile in the Repo. Defaults to `Dockerfile`.
</td>
</tr>
<tr>
<td>
<code>template</code></br>
<em>
string
</em>
</td>
<td>
Specify the build template. Defaults to `buildkit`.
</td>
</tr>
<tr>
<td>
<code>secret</code></br>
<em>
string
</em>
</td>
<td>
Specify the secret name. If specified, it will register a webhook and only creates new revision if webhook is triggered.
</td>
</tr>
</tbody>
</table>
<h3 id="Match">Match
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteSpec">RouteSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>path</code></br>
<em>
<a href="#StringMatch">
StringMatch
</a>
</em>
</td>
<td>
URI to match values are case-sensitive and formatted as follows:<br/><br/>   exact: "value" for exact string match<br/><br/>   prefix: "value" for prefix-based match<br/><br/>   regex: "value" for ECMAscript style regex-based match
</td>
</tr>
<tr>
<td>
<code>scheme</code></br>
<em>
<a href="#StringMatch">
StringMatch
</a>
</em>
</td>
<td>
URI Scheme values are case-sensitive and formatted as follows:<br/><br/>   exact: "value" for exact string match<br/><br/>   prefix: "value" for prefix-based match<br/><br/>   regex: "value" for ECMAscript style regex-based match
</td>
</tr>
<tr>
<td>
<code>method</code></br>
<em>
<a href="#StringMatch">
StringMatch
</a>
</em>
</td>
<td>
HTTP Method values are case-sensitive and formatted as follows:<br/><br/>   exact: "value" for exact string match<br/><br/>   prefix: "value" for prefix-based match<br/><br/>   regex: "value" for ECMAscript style regex-based match
</td>
</tr>
<tr>
<td>
<code>headers</code></br>
<em>
<a href="#StringMatch">
map[string]github.com/rancher/rio/pkg/apis/rio.cattle.io/v1.StringMatch
</a>
</em>
</td>
<td>
The header keys must be lowercase and use hyphen as the separator, e.g. x-request-id.<br/><br/>Header values are case-sensitive and formatted as follows:<br/><br/>   exact: "value" for exact string match<br/><br/>   prefix: "value" for prefix-based match<br/><br/>   regex: "value" for ECMAscript style regex-based match<br/><br/>Note: The keys uri, scheme, method, and authority will be ignored.
</td>
</tr>
<tr>
<td>
<code>cookies</code></br>
<em>
<a href="#StringMatch">
map[string]github.com/rancher/rio/pkg/apis/rio.cattle.io/v1.StringMatch
</a>
</em>
</td>
<td>
Cookies must be lowercase and use hyphen as the separator, e.g. x-request-id.<br/><br/>Header values are case-sensitive and formatted as follows:<br/><br/>   exact: "value" for exact string match<br/><br/>   prefix: "value" for prefix-based match<br/><br/>   regex: "value" for ECMAscript style regex-based match<br/><br/>Note: The keys uri, scheme, method, and authority will be ignored.
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
int
</em>
</td>
<td>
Specifies the ports on the host that is being addressed. Many services only expose a single port or label ports with the protocols they support, in these cases it is not required to explicitly select the port.
</td>
</tr>
<tr>
<td>
<code>from</code></br>
<em>
<a href="#ServiceSource">
ServiceSource
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="NamedContainer">NamedContainer
</h3>
<p>
(<em>Appears on:</em>
<a href="#PodConfig">PodConfig</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
The name of the container
</td>
</tr>
<tr>
<td>
<code>init</code></br>
<em>
bool
</em>
</td>
<td>
List of initialization containers belonging to the pod.
Init containers are executed in order prior to containers being started.
If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy.
The name for an init container or normal container must be unique among all containers.
Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes.
The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers.
Limits are applied to init containers in a similar fashion. Init containers cannot currently be added or removed. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
</td>
</tr>
<tr>
<td>
<code>Container</code></br>
<em>
<a href="#Container">
Container
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Permission">Permission
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceSpec">ServiceSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>role</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>verbs</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>apiGroup</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>resource</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>url</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>resourceName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="PodConfig">PodConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceSpec">ServiceSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>containers</code></br>
<em>
<a href="#NamedContainer">
[]NamedContainer
</a>
</em>
</td>
<td>
List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated.
</td>
</tr>
<tr>
<td>
<code>dnsPolicy</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#dnspolicy-v1-core">
Kubernetes core/v1.DNSPolicy
</a>
</em>
</td>
<td>
Set DNS policy for the pod. Defaults to "ClusterFirst". Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'.
</td>
</tr>
<tr>
<td>
<code>hostname</code></br>
<em>
string
</em>
</td>
<td>
Specifies the hostname of the Pod If not specified, the pod's hostname will be set to a system-defined value.
</td>
</tr>
<tr>
<td>
<code>hostAliases</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#hostalias-v1-core">
[]Kubernetes core/v1.HostAlias
</a>
</em>
</td>
<td>
HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file if specified. This is only valid for non-hostNetwork pods.
</td>
</tr>
<tr>
<td>
<code>PodDNSConfig</code></br>
<em>
<a href="#PodDNSConfig">
PodDNSConfig
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>Container</code></br>
<em>
<a href="#Container">
Container
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="PodDNSConfig">PodDNSConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#PodConfig">PodConfig</a>)
</p>
<p>
PodDNSConfig Specifies the DNS parameters of a pod. Parameters specified here will be merged to the generated DNS configuration based on DNSPolicy.
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>dnsNameservers</code></br>
<em>
[]string
</em>
</td>
<td>
A list of DNS name server IP addresses. This will be appended to the base nameservers generated from DNSPolicy. Duplicated nameservers will be removed.
</td>
</tr>
<tr>
<td>
<code>dnsSearches</code></br>
<em>
[]string
</em>
</td>
<td>
A list of DNS search domains for host-name lookup. This will be appended to the base search paths generated from DNSPolicy. Duplicated search paths will be removed.
</td>
</tr>
<tr>
<td>
<code>dnsOptions</code></br>
<em>
<a href="#PodDNSConfigOption">
[]PodDNSConfigOption
</a>
</em>
</td>
<td>
A list of DNS resolver options. This will be merged with the base options generated from DNSPolicy.
Duplicated entries will be removed. Resolution options given in Options will override those that appear in the base DNSPolicy.
</td>
</tr>
</tbody>
</table>
<h3 id="PodDNSConfigOption">PodDNSConfigOption
</h3>
<p>
(<em>Appears on:</em>
<a href="#PodDNSConfig">PodDNSConfig</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>value</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Protocol">Protocol
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#ContainerPort">ContainerPort</a>)
</p>
<p>
</p>
<h3 id="PublicDomainSpec">PublicDomainSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#PublicDomain">PublicDomain</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
SecretRef reference the secret that contains key and certs for TLS configuration. By default it is configured to use Letsencrypt
</td>
</tr>
<tr>
<td>
<code>disableLetsencrypt</code></br>
<em>
bool
</em>
</td>
<td>
Whether to disable Letsencrypt certificates.
</td>
</tr>
<tr>
<td>
<code>targetServiceName</code></br>
<em>
string
</em>
</td>
<td>
Target Service Name in the same Namespace
</td>
</tr>
<tr>
<td>
<code>domainName</code></br>
<em>
string
</em>
</td>
<td>
PublicDomain name
</td>
</tr>
</tbody>
</table>
<h3 id="PublicDomainStatus">PublicDomainStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#PublicDomain">PublicDomain</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>httpsSupported</code></br>
<em>
bool
</em>
</td>
<td>
Whether HTTP is supported in the Domain
</td>
</tr>
<tr>
<td>
<code>endpoint</code></br>
<em>
string
</em>
</td>
<td>
Endpoint to access this Domain
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
Represents the latest available observations of a PublicDomain's current state.
</td>
</tr>
</tbody>
</table>
<h3 id="Question">Question
</h3>
<p>
(<em>Appears on:</em>
<a href="#FeatureSpec">FeatureSpec</a>, 
<a href="#TemplateMeta">TemplateMeta</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>variable</code></br>
<em>
string
</em>
</td>
<td>
The variable name to reference using ${...} syntax
</td>
</tr>
<tr>
<td>
<code>label</code></br>
<em>
string
</em>
</td>
<td>
A friend name for the question
</td>
</tr>
<tr>
<td>
<code>description</code></br>
<em>
string
</em>
</td>
<td>
A longer description of the question
</td>
</tr>
<tr>
<td>
<code>type</code></br>
<em>
string
</em>
</td>
<td>
The field type: string, int, bool, enum. default is string
</td>
</tr>
<tr>
<td>
<code>required</code></br>
<em>
bool
</em>
</td>
<td>
The answer can not be blank
</td>
</tr>
<tr>
<td>
<code>default</code></br>
<em>
string
</em>
</td>
<td>
Default value of the answer if not specified by the user
</td>
</tr>
<tr>
<td>
<code>group</code></br>
<em>
string
</em>
</td>
<td>
Group the question with questions in the same group (Most used by UI)
</td>
</tr>
<tr>
<td>
<code>minLength</code></br>
<em>
int
</em>
</td>
<td>
Minimum length of the answer
</td>
</tr>
<tr>
<td>
<code>maxLength</code></br>
<em>
int
</em>
</td>
<td>
Maximum length of the answer
</td>
</tr>
<tr>
<td>
<code>min</code></br>
<em>
int
</em>
</td>
<td>
Minimum value of an int answer
</td>
</tr>
<tr>
<td>
<code>max</code></br>
<em>
int
</em>
</td>
<td>
Maximum value of an int answer
</td>
</tr>
<tr>
<td>
<code>options</code></br>
<em>
[]string
</em>
</td>
<td>
An array of valid answers for type enum questions
</td>
</tr>
<tr>
<td>
<code>validChars</code></br>
<em>
string
</em>
</td>
<td>
Answer must be composed of only these characters
</td>
</tr>
<tr>
<td>
<code>invalidChars</code></br>
<em>
string
</em>
</td>
<td>
Answer must not have any of these characters
</td>
</tr>
<tr>
<td>
<code>subquestions</code></br>
<em>
<a href="#SubQuestion">
[]SubQuestion
</a>
</em>
</td>
<td>
A list of questions that are considered child questions
</td>
</tr>
<tr>
<td>
<code>showIf</code></br>
<em>
string
</em>
</td>
<td>
Ask question only if this evaluates to true, more info on syntax below
</td>
</tr>
<tr>
<td>
<code>showSubquestionIf</code></br>
<em>
string
</em>
</td>
<td>
Ask subquestions if this evaluates to true
</td>
</tr>
</tbody>
</table>
<h3 id="Redirect">Redirect
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteSpec">RouteSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>host</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Retry">Retry
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteTraffic">RouteTraffic</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>attempts</code></br>
<em>
int
</em>
</td>
<td>
REQUIRED. Number of retries for a given request. The interval between retries will be determined automatically (25ms+).
Actual number of retries attempted depends on the httpReqTimeout.
</td>
</tr>
<tr>
<td>
<code>timeoutMillis</code></br>
<em>
int
</em>
</td>
<td>
Timeout per retry attempt for a given request. Units: milliseconds
</td>
</tr>
</tbody>
</table>
<h3 id="Revision">Revision
</h3>
<p>
(<em>Appears on:</em>
<a href="#AppSpec">AppSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>public</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serviceName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>Version</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>adjustedWeight</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>weight</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>scale</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>scaleStatus</code></br>
<em>
<a href="#ScaleStatus">
ScaleStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deploymentReady</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>RolloutConfig</code></br>
<em>
<a href="#RolloutConfig">
RolloutConfig
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Rewrite">Rewrite
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteSpec">RouteSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>host</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="RolloutConfig">RolloutConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#Revision">Revision</a>, 
<a href="#ServiceSpec">ServiceSpec</a>)
</p>
<p>
RolloutConfig specifies the configuration when promoting a new revision
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>rollout</code></br>
<em>
bool
</em>
</td>
<td>
Whether to turn on Rollout(changing the weight gradually)
</td>
</tr>
<tr>
<td>
<code>rolloutIncrement</code></br>
<em>
int
</em>
</td>
<td>
Increment Value each Rollout can scale up or down
</td>
</tr>
<tr>
<td>
<code>rolloutInterval</code></br>
<em>
int
</em>
</td>
<td>
Increment Interval between each Rollout
</td>
</tr>
</tbody>
</table>
<h3 id="RouteSpec">RouteSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouterSpec">RouterSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>matches</code></br>
<em>
<a href="#Match">
[]Match
</a>
</em>
</td>
<td>
Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics.
The rule is matched if any one of the match blocks succeed.
</td>
</tr>
<tr>
<td>
<code>to</code></br>
<em>
<a href="#WeightedDestination">
[]WeightedDestination
</a>
</em>
</td>
<td>
A http rule can either redirect or forward (default) traffic. The forwarding target can be one of several versions of a service (see glossary in beginning of document).
Weights associated with the service version determine the proportion of traffic it receives.
</td>
</tr>
<tr>
<td>
<code>redirect</code></br>
<em>
<a href="#Redirect">
Redirect
</a>
</em>
</td>
<td>
A http rule can either redirect or forward (default) traffic. If traffic passthrough option is specified in the rule, route/redirect will be ignored.
The redirect primitive can be used to send a HTTP 301 redirect to a different URI or Authority.
</td>
</tr>
<tr>
<td>
<code>rewrite</code></br>
<em>
<a href="#Rewrite">
Rewrite
</a>
</em>
</td>
<td>
Rewrite HTTP URIs and Authority headers. Rewrite cannot be used with Redirect primitive. Rewrite will be performed before forwarding.
</td>
</tr>
<tr>
<td>
<code>addHeaders</code></br>
<em>
[]string
</em>
</td>
<td>
Header manipulation rules
</td>
</tr>
<tr>
<td>
<code>RouteTraffic</code></br>
<em>
<a href="#RouteTraffic">
RouteTraffic
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="RouteTraffic">RouteTraffic
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteSpec">RouteSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>fault</code></br>
<em>
<a href="#Fault">
Fault
</a>
</em>
</td>
<td>
Fault injection policy to apply on HTTP traffic at the client side. Note that timeouts or retries will not be enabled when faults are enabled on the client side.
</td>
</tr>
<tr>
<td>
<code>mirror</code></br>
<em>
<a href="#Destination">
Destination
</a>
</em>
</td>
<td>
Mirror HTTP traffic to a another destination in addition to forwarding the requests to the intended destination.
Mirrored traffic is on a best effort basis where the sidecar/gateway will not wait for the mirrored cluster to respond before returning the response from the original destination.
Statistics will be generated for the mirrored destination.
</td>
</tr>
<tr>
<td>
<code>timeoutMillis</code></br>
<em>
int
</em>
</td>
<td>
Timeout for HTTP requests.
</td>
</tr>
<tr>
<td>
<code>retry</code></br>
<em>
<a href="#Retry">
Retry
</a>
</em>
</td>
<td>
Retry policy for HTTP requests.
</td>
</tr>
</tbody>
</table>
<h3 id="RouterSpec">RouterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#Router">Router</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>routes</code></br>
<em>
<a href="#RouteSpec">
[]RouteSpec
</a>
</em>
</td>
<td>
An ordered list of route rules for HTTP traffic. The first rule matching an incoming request is used.
</td>
</tr>
</tbody>
</table>
<h3 id="RouterStatus">RouterStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#Router">Router</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>publicDomains</code></br>
<em>
[]string
</em>
</td>
<td>
The list of publicedomains pointing to the router
</td>
</tr>
<tr>
<td>
<code>endpoint</code></br>
<em>
[]string
</em>
</td>
<td>
The endpoint to access the router
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
Represents the latest available observations of a PublicDomain's current state.
</td>
</tr>
</tbody>
</table>
<h3 id="ScaleStatus">ScaleStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#Revision">Revision</a>, 
<a href="#ServiceStatus">ServiceStatus</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ready</code></br>
<em>
int
</em>
</td>
<td>
Total number of ready pods targeted by this deployment.
</td>
</tr>
<tr>
<td>
<code>unavailable</code></br>
<em>
int
</em>
</td>
<td>
Total number of unavailable pods targeted by this deployment. This is the total number of pods that are still required for the deployment to have 100% available capacity.
They may either be pods that are running but not yet available or pods that still have not been created.
</td>
</tr>
<tr>
<td>
<code>available</code></br>
<em>
int
</em>
</td>
<td>
Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.
</td>
</tr>
<tr>
<td>
<code>updated</code></br>
<em>
int
</em>
</td>
<td>
Total number of non-terminated pods targeted by this deployment that have the desired template spec.
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceObservedWeight">ServiceObservedWeight
</h3>
<p>
(<em>Appears on:</em>
<a href="#AppStatus">AppStatus</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>lastWrite</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>weight</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serviceName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceRevision">ServiceRevision
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceSpec">ServiceSpec</a>)
</p>
<p>
ServiceRevision speficies the APP name, Version and Weight to uniquely identify each Revision
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
Revision Version
</td>
</tr>
<tr>
<td>
<code>weight</code></br>
<em>
int
</em>
</td>
<td>
Revision Weight
</td>
</tr>
<tr>
<td>
<code>app</code></br>
<em>
string
</em>
</td>
<td>
Revision App name
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceScale">ServiceScale
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceSpec">ServiceSpec</a>)
</p>
<p>
ServiceScale Specifies the scale parameters for Service
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>scale</code></br>
<em>
int
</em>
</td>
<td>
Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.
</td>
</tr>
<tr>
<td>
<code>updateBatchSize</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
The maximum number of pods that can be scheduled above the desired number of pods. Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
This can not be 0 if MaxUnavailable is 0. Absolute number is calculated from percentage by rounding up.
Defaults to 25%. Example: when this is set to 30%, the new ReplicaSet can be scaled up immediately when the rolling update starts, such that the total number of old and new pods do not exceed 130% of desired pods.
Once old pods have been killed, new ReplicaSet can be scaled up further, ensuring that total number of pods running at any time during the update is at most 130% of desired pods.
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceSource">ServiceSource
</h3>
<p>
(<em>Appears on:</em>
<a href="#Match">Match</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>service</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>stack</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>revision</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceSpec">ServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#Service">Service</a>)
</p>
<p>
ServiceSpec represents spec for Service
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ServiceScale</code></br>
<em>
<a href="#ServiceScale">
ServiceScale
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>ServiceRevision</code></br>
<em>
<a href="#ServiceRevision">
ServiceRevision
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>AutoscaleConfig</code></br>
<em>
<a href="#AutoscaleConfig">
AutoscaleConfig
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>RolloutConfig</code></br>
<em>
<a href="#RolloutConfig">
RolloutConfig
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>PodConfig</code></br>
<em>
<a href="#PodConfig">
PodConfig
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>disableServiceMesh</code></br>
<em>
bool
</em>
</td>
<td>
Whether to disable ServiceMesh for Service. If true, no mesh sidecar will be deployed along with the Service
</td>
</tr>
<tr>
<td>
<code>permissions</code></br>
<em>
<a href="#Permission">
[]Permission
</a>
</em>
</td>
<td>
Permissions to the Services. It will create corresponding ServiceAccounts, Roles and RoleBinding.
</td>
</tr>
<tr>
<td>
<code>globalPermissions</code></br>
<em>
<a href="#Permission">
[]Permission
</a>
</em>
</td>
<td>
GlobalPermissions to the Services. It will create corresponding ServiceAccounts, ClusterRoles and ClusterRoleBinding.
</td>
</tr>
<tr>
<td>
<code>systemSpec</code></br>
<em>
<a href="#SystemServiceSpec">
SystemServiceSpec
</a>
</em>
</td>
<td>
System Field Spec
</td>
</tr>
</tbody>
</table>
<h3 id="ServiceStatus">ServiceStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#Service">Service</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>deploymentStatus</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#deploymentstatus-v1-apps">
Kubernetes apps/v1.DeploymentStatus
</a>
</em>
</td>
<td>
Most recently observed status of the Deployment.
</td>
</tr>
<tr>
<td>
<code>scaleStatus</code></br>
<em>
<a href="#ScaleStatus">
ScaleStatus
</a>
</em>
</td>
<td>
ScaleStatus for the Service
</td>
</tr>
<tr>
<td>
<code>scaleFromZeroTimestamp</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
Last timestamp scaled from zero replica
</td>
</tr>
<tr>
<td>
<code>observedScale</code></br>
<em>
int
</em>
</td>
<td>
ObservedScale is calcaluted from autoscaling component to make sure pod has the desired load
</td>
</tr>
<tr>
<td>
<code>weightOverride</code></br>
<em>
int
</em>
</td>
<td>
WeightOverride is the weight calculated from serviceset revision
</td>
</tr>
<tr>
<td>
<code>containerImages</code></br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
[]github.com/rancher/wrangler/pkg/genericcondition.GenericCondition
</em>
</td>
<td>
Represents the latest available observations of a deployment's current state.
</td>
</tr>
<tr>
<td>
<code>endpoints</code></br>
<em>
[]string
</em>
</td>
<td>
The Endpoints to access the service
</td>
</tr>
<tr>
<td>
<code>publicDomains</code></br>
<em>
[]string
</em>
</td>
<td>
The list of publicdomains pointing to the service
</td>
</tr>
</tbody>
</table>
<h3 id="StringMatch">StringMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#Match">Match</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>exact</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>prefix</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>regexp</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="SubQuestion">SubQuestion
</h3>
<p>
(<em>Appears on:</em>
<a href="#Question">Question</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>variable</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>label</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>description</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>type</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>required</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>default</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>group</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>minLength</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>maxLength</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>min</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>max</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>options</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>validChars</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>invalidChars</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>showIf</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="SystemServiceSpec">SystemServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#ServiceSpec">ServiceSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>updateOrder</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>updateStrategy</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deploymentStrategy</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>global</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>volumeClaimTemplates</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#persistentvolumeclaim-v1-core">
[]Kubernetes core/v1.PersistentVolumeClaim
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>podSpec</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#podspec-v1-core">
Kubernetes core/v1.PodSpec
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="TemplateMeta">TemplateMeta
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>iconUrl</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>readme</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>questions</code></br>
<em>
<a href="#Question">
[]Question
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Volume">Volume
</h3>
<p>
(<em>Appears on:</em>
<a href="#Container">Container</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>Name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>Path</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="WeightedDestination">WeightedDestination
</h3>
<p>
(<em>Appears on:</em>
<a href="#RouteSpec">RouteSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>Destination</code></br>
<em>
<a href="#Destination">
Destination
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>weight</code></br>
<em>
int
</em>
</td>
<td>
Weight for the Destination
</td>
</tr>
</tbody>
</table>
<hr/>
<h2 id="webhookinator.rio.cattle.io">webhookinator.rio.cattle.io</h2>
<p>
</p>
Resource Types:
<ul><li>
<a href="#GitWebHookExecution">GitWebHookExecution</a>
</li><li>
<a href="#GitWebHookReceiver">GitWebHookReceiver</a>
</li></ul>
<h3 id="GitWebHookExecution">GitWebHookExecution
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
webhookinator.rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>GitWebHookExecution</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#GitWebHookExecutionSpec">
GitWebHookExecutionSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>payload</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>gitWebHookReceiverName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>commit</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>branch</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tag</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pr</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>sourceLink</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repositoryUrl</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>title</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>author</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>authorEmail</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>authorAvatar</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#GitWebHookExecutionStatus">
GitWebHookExecutionStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="GitWebHookReceiver">GitWebHookReceiver
</h3>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
webhookinator.rio.cattle.io/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>GitWebHookReceiver</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#GitWebHookReceiverSpec">
GitWebHookReceiverSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>repositoryUrl</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repositoryCredentialSecretName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>provider</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>push</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pr</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tag</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>executionLabels</code></br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#GitWebHookReceiverStatus">
GitWebHookReceiverStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="Condition">Condition
</h3>
<p>
(<em>Appears on:</em>
<a href="#GitWebHookExecutionStatus">GitWebHookExecutionStatus</a>, 
<a href="#GitWebHookReceiverStatus">GitWebHookReceiverStatus</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
string
</em>
</td>
<td>
Type of the condition.
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
Status of the condition, one of True, False, Unknown.
</td>
</tr>
<tr>
<td>
<code>lastUpdateTime</code></br>
<em>
string
</em>
</td>
<td>
The last time this condition was updated.
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
string
</em>
</td>
<td>
Last time the condition transitioned from one status to another.
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
The reason for the condition's last transition.
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
Human-readable message indicating details about last transition
</td>
</tr>
</tbody>
</table>
<h3 id="GitWebHookExecutionSpec">GitWebHookExecutionSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#GitWebHookExecution">GitWebHookExecution</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>payload</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>gitWebHookReceiverName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>commit</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>branch</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tag</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pr</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>sourceLink</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repositoryUrl</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>title</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>author</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>authorEmail</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>authorAvatar</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="GitWebHookExecutionStatus">GitWebHookExecutionStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#GitWebHookExecution">GitWebHookExecution</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#Condition">
[]Condition
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>statusUrl</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>appliedStatus</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="GitWebHookReceiverSpec">GitWebHookReceiverSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#GitWebHookReceiver">GitWebHookReceiver</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>repositoryUrl</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repositoryCredentialSecretName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>provider</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>push</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pr</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tag</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>executionLabels</code></br>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="GitWebHookReceiverStatus">GitWebHookReceiverStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#GitWebHookReceiver">GitWebHookReceiver</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#Condition">
[]Condition
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>token</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>hookId</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>b41e9f57</code>.
</em></p>
