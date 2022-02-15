# Basic connection pool for go-ldap

This little library use the [go-ldap](https://github.com/go-ldap/ldap) library and pool connections for you.

Don't hesitate to open issues or pull requests, I write that for my own need so that miss some features or tests :) 


## Features

- Customize the number of connections to keep alive.
- Custom timeout duration while awaiting a available connection
- Reconnect when a connection goes down or the server doesn't respond anymore.

## Limitation

- Fixed connections count.
- Some functions is missing, I only added Search, SearchWithPaging, Add, Modify, ModifyWithResult, ModifyDN, Del.
- Only anonymous or simple bind is available for the moment.


## How to use it ?

```go
p, err := ldappool.NewPool(context.Background(), &ldappool.PoolOptions{
    URL: "ldaps://ldap.example.fr",
    // Manage 10 connections
    ConnectionsCount: 10,
    // Every request is going to wait a available connection 5 seconds and return an error if there is no connections available
    ConnectionTimeout: time.Second * 5,
    // If a connection is marked as unavailable after a heartbeat, we try to connect every 5 seconds
    WakeupInterval: time.Second * 5,
    BindCredentials: &ldappool.BindCredentials{
        Username: "cn=admin,dc=example,dc=fr",
        Password: "toto",
    },
})

if err != nil {
    log.Fatal(err)
}

res, err := p.Search(ldap.NewSearchRequest(
    "dc=example,dc=fr",
    ldap.ScopeWholeSubtree,
    ldap.NeverDerefAliases, 0, 0, false,
    "(&(objectClass=posixAccount))",
    []string{"*"}, []ldap.Control{}),
)

log.Println(res, err)
```