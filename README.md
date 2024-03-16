# Entx

Entx is my personal collection of useful snippet I use in any project involving using [Ent](https://entgo.io/). Currently only three feature is present.

## Usage

Create a new go file, we'll use it to generate our ent codebase. You need to tweak the path you pass to the `entc.Generate` function.

In this example you can notice I use some community extension and additional ent features, you can tweak it as you want, the only important part is the use of `entx.New(nil),` to instantiate the extension. You can use the nil value to enable all features of the extension of pass an option struct to customize which extension to enable.

```go
package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/hedwigz/entviz"
	"github.com/m-vinc/entx"
)

func main() {
	err := entc.Generate("../../pkg/ent/schema", &gen.Config{
		Features: []gen.Feature{gen.FeatureLock, gen.FeatureExecQuery, gen.FeatureUpsert},
	}, entc.Extensions(
        entviz.Extension{},
        entx.New(nil),
        // entx.New(&entx.Options{Tx: true}),
    ))
	if err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}

```

## Features

### Postgres

Postgres generate only simple function when using a postgresql database.

```go
db, err := ent.Postgres(&ent.PostgresConfiguration{
    Host:     configDatabase.Host,
    Port:     strconv.FormatUint(uint64(configDatabase.Port), 10),
    Username: configDatabase.Username,
    Password: configDatabase.Password,
    Database: configDatabase.Database,
    SSLMode:  configDatabase.SSLMode,
})

if err != nil {
    log.Fatal(err)
}
```

### Queryable

Queryable generate what you need to abstract a plain client or a transaction and create agnostic function, allowing you to build function which is able to query the database in any circumstance. Useful for validations function we'll see in this simple example :

```go
func IsNameUnique(queryable ent.Queryable, name string) error {
	ok, err := queryable.AccountClient().Query().
        Where(account.NameEQ(name)).
        Exist(queryable.Context())
	if err != nil {
		return err
	}

	if ok {
		return ErrNameDuplicate
	}

	return nil
}

ctx := context.Background()

// Use queryable with a simple ent client
var db *ent.Client
q := db.Queryable(ctx)

// Run with a simple client
IsNameUnique(q, "toto")



// Use queryable with a transaction
var tx *ent.Tx

q := tx.Queryable(ctx)

// Run within the transaction
IsNameUnique(q, "toto")
```

### Tx

The Tx feature extend the ent Client sturct and the Tx struct, adding two function `Acquire` to the ent `*Client` and a `Release` function on `*Tx`.

With these two function we can instantiate a transaction with `Acquire` which store a transaction in the given context and returning a new context. :

```go

ctx := context.Background()

var db *ent.Client

tx, ctx, err := db.Acquire(ctx)
if err != nil {
    log.Fatal("help !", err)
}
```

When a transaction is initated this way, you can release it by calling the `Release` function on the `*Tx` struct with the current context :

```go

err := errors.New("toto")
// this tx will be rollbacked because an error is passed to the release function
err = tx.Release(ctx, err) // Rollback

// However we can commit the transaction if no error is passed
err = tx.Release(ctx, nil) // Commit !
```

By using those two functions, we can create function using them and allowing a function to initate a transaction, use it, pass it to another function which use the current transaction if a transaction is present.

The original caller is "responsible" of the transaction, only him can `Commit` or `Rollback` the transaction by calling `Release`, if a child function call need to use the transaction, you'll need to pass the context created by the `Acquire` function, the child function will use the same pattern as the original caller but in that case no new transaction will be initiated and every call to `Release` or `Acquire` will not do anything other than forwarding the error to othe caller.

Let's demonstrate how this mechanism work, let's define two simple function :

```go
func GetUser(ctx context.Context, username string) (*ent.User, error) {
    tx, ctx, err := db.Acquire(ctx)
    if err != nil {
        return nil, err
    }

    user, err := tx.User.Query().Where(user.UsernameEQ(username)).First(ctx)
    if err != nil {
        return nil, tx.Release(ctx, err)
    }

    return user, tx.Release(ctx, nil)
}

func UpdateUsername(ctx context.Context, username stirng, newUsername string) error {
    tx, ctx, err := db.Acquire(ctx)
    if err != nil {
        return err
    }

    user, err := GetUser(ctx, username)
    if err != nil {
        return tx.Release(ctx, err)
    }

    _, err = tx.User.UpdateOne(user).SetUsername(newUsername).Save(ctx)
    if err != nil {
        return tx.Release(ctx, err)
    }

    return tx.Release(ctx, nil)
}
```

Then let's use them :

```go
var db *ent.Client

ctx := context.Backgound()

// In this case, UpdateUsername will initiate the transaction
// Pass it to GetUser which use it to retrieve a user or throw an error if not exist
// If an error occured GetUser call tx.Release but since the transaction is not owned by GetUser the error is forward to UpdateUsername and the tx.Release call is performed by UpdateUsername rollbacking the transaction or commiting it.
err = UpdateUsername(ctx, "toto", "tata")

// We can use GetUser without initating a transaction, the Acquire call will do it for you.
user, err := GetUser(ctx, "toto")

```