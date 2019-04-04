# ECAD Labs auth daemon

`auth` is an authentication and authorization daemon that issues JWT tokens.

It's features include;

* Authentication of user credentials for JWT tokens
* User management API to create/invite/modify/delete users
* User Roles that allow assignment of permissions to users 
* Multi-tenant allowing "teams" of users. Suitable for supporting multiple
    tenant customers, each of which have users as members in their tenant
* Roles are assigned to users membership in a tenant, allowing a user to have
    a different set of permissions depending on the tenant they are a member
    of
* User creation life-cycle using invitation tokens via email
* User password reset functionality.
* Role Based Access Control (RBAC) based on collections (roles) of permission
    properties.

The project also includes a set of Angular components for managing
users/tenants/roles that can be used as is, or forked and used in other
projects. 

# Roles 

Roles are collections of permission properties. A user can have multiple roles
assigned to her. When a user logs in, the auth daemon will return all
permission properties that are assigned to the user via the users roles. These
roles are included in the JWT payload. 

## Permission Properties

Permission properties are simply strings that are consumed by other services.

Each property has a consumer that will enforce some sort of behaviour based on
the presence or absence of permission properties. It's up to the service
operator to define their own permission properties, and in turn, enforce rules
based on the presence or absence of properties.

Permission properties are structured as follows:

`<namespace>`.`<resource_name>`.`<action_verb>`

The auth daemon is itself a consumer of permission properties, but specifically
for properties that applies to resources that auth controls. These are
`users`, `tenants` and `service_accounts`, all of which are namespaced under
the string `com.ecadlabs.`

Other systems that will consume JWT tokens will see permission properties that
they have no interest in.

## Defining new permission properties

If your using the `auth` daemon, then you will likely want to define
permissions that apply to your systems domain. To illustrate the definition of
a new permission set, we will imagine a service named `pinger`. It's job is to
send pings, and record the sending of pings. 

As an example, we will use a namespace based on the domain name `example.net`.
This protects us from the possibility of different system using `ping` as a
resource name.

Our permission properties will be:

`net.example.ping.read`: Allows caller to view all ping records
`net.example.net.create`: Allows caller to send a ping

We add these definitions to our `auth` daemon via the `rbac.yaml` definition
file. The auth daemon has no intelligence around how these permissions will be
used. In your `pinger` service, you will decode and validate the JWT token, and
within the token. When your `pinger` service receives a request to list
`pings`, your service should assert the presence of the `net.example.ping.read`
permission property. If it is not present, your service should reject the
request with a HTTP code such as `403 - Forbidden`. If your service does find
the appropriate permission property, then it can service the call accordingly.

# Service Accounts and API Keys

Auth supports "Service Accounts" which are a special type of account designed
for "Machine to Machine" integrations. An admin can create a service account,
and generate an API key/token which can then be used to interact with the
other services that validate the JWT tokens. 

Additionally, a service account can be configured to use an "Allow List" of IP
CIDR ranges. If the calling parties IP address falls within an Allow List CIDR
range, the caller will be issued a JWT token.

