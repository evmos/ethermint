package types

// AppContext provides the ability for the application to pass around and
// obtain immutable objects easily. More importantly, it allows for the
// utilization of the object-capability model in which components gain access
// to other components for which they truly need.
type AppContext struct {
}
