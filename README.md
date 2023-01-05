# Basic connection pool for go-ldap

This little library use the [go-ldap](https://github.com/go-ldap/ldap) library and pool connections for you.

Don't hesitate to open issues or pull requests, I write that for my own need so that miss some features or tests :) 


## Features

- Customize the number of connections to keep alive.
- Reconnect when a connection or the server goes down

## Limitation

- Fixed connections count.
- Some functions is missing, I only added Search, SearchWithPaging, Add, Modify, ModifyWithResult, ModifyDN, Del.
- Only anonymous or simple bind is available for the moment.


## How to use it ?

```go
p, err := ldappool.NewPool(&ldappool.PoolOptions{
    URL: "ldaps://ldap.example.fr",
    // Manage 10 connections
    ConnectionsCount: 10,
    // Your bind credentials if you want to bind as non-anonymous user
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
