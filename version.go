package custd

// Version is this SDK's released version. Custd learns which release a caller
// runs from the X-Custd-Sdk header, so a consumer on an old version becomes
// visible instead of being found by grepping repositories.
//
// scripts/check-sdk-version-identity.sh asserts it equals VERSION and
// scripts/bump-version.sh updates it, so a release cannot leave it behind.
const Version = "2.4.0"

// sdkIdentity is the X-Custd-Sdk value for this SDK.
const sdkIdentity = "go/" + Version
