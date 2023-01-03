package testcase

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Login struct {
	Username string
	Password string
}

type PointerSlice struct {
	Logins []*Login
}

type ID struct {
	ObjectID primitive.ObjectID
}
