package handlers

const (
	permissionWrite     = "com.ecadlabs.users.write"
	permissionRead      = "com.ecadlabs.users.read"
	permissionReadSelf  = "com.ecadlabs.users.read_self"
	permissionWriteSelf = "com.ecadlabs.users.write_self"
	permissionFull      = "com.ecadlabs.users.full_control"
	permissionLogs      = "com.ecadlabs.users.read_logs"

	permissionDelegatePrefix = "com.ecadlabs.users.delegate:"

	permissionTenantsFull       = "com.ecadlabs.tenants.full_control"
	permissionTenantsRead       = "com.ecadlabs.tenants.read"
	permissionTenantsWrite      = "com.ecadlabs.tenants.write"
	permissionTenantsReadOwned  = "com.ecadlabs.tenants.read_owned"
	permissionTenantsWriteOwned = "com.ecadlabs.tenants.write_owned"
	permissionTenantsCreate     = "com.ecadlabs.tenants.create"

	permissionServiceWrite = "com.ecadlabs.service_accounts.write"
	permissionServiceRead  = "com.ecadlabs.service_accounts.read"
	permissionServiceFull  = "com.ecadlabs.service_accounts.full_control"
)
