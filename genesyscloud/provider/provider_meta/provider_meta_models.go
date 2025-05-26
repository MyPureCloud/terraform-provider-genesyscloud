package provider_meta

//
//import "github.com/mypurecloud/platform-client-sdk-go/v152/platformclientv2"
//
//type GenesysCloudProvider struct {
//	Version       string
//	SdkClientPool SDKClientPool
//
//	AttributeEnvValues     *ProviderEnvVars
//	TokenPoolSize          int32
//	LogStackTraces         bool
//	LogStackTracesFilePath string
//
//	AuthDetails  *AuthInfo
//	SdkDebugInfo *SdkDebugInfo
//	Proxy        *Proxy
//	Gateway      *Gateway
//}
//
//type AuthInfo struct {
//	AccessToken  string
//	ClientId     string
//	ClientSecret string
//	Region       string
//}
//
//type SdkDebugInfo struct {
//	DebugEnabled bool
//	Format       string
//	FilePath     string
//}
//
//type Proxy struct {
//	Port     string
//	Host     string
//	Protocol string
//	Auth     *Auth
//}
//
//type Auth struct {
//	Username string
//	Password string
//}
//
//type Gateway struct {
//	Port       string
//	Host       string
//	Protocol   string
//	PathParams []PathParam
//	Auth       *Auth
//}
//
//type PathParam struct {
//	PathName  string
//	PathValue string
//}
//
//// SDKClientPool holds a Pool of client configs for the Genesys Cloud SDK. One should be
//// acquired at the beginning of any resource operation and released on completion.
//// This has the benefit of ensuring we don't issue too many concurrent requests and also
//// increases throughput as each token will have its own rate limit.
//// (duplicate)
//type SDKClientPool struct {
//	Pool chan *platformclientv2.Configuration
//}
//
//type ProviderEnvVars struct {
//	accessToken            string
//	region                 string
//	clientId               string
//	clientSecret           string
//	sdkDebug               string
//	sdkDebugFilePath       string
//	sdkDebugFormat         string
//	tokenPoolSize          string
//	logStackTraces         string
//	logStackTracesFilePath string
//	gatewayHost            string
//	gatewayPort            string
//	gatewayProtocol        string
//	gatewayAuthUsername    string
//	gatewayAuthPassword    string
//	gatewayPathParamsName  string
//	gatewayPathParamsValue string
//	proxyHost              string
//	proxyPort              string
//	proxyProtocol          string
//	proxyAuthUsername      string
//	proxyAuthPassword      string
//}
//
//
