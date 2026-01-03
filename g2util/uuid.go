package g2util

import (
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
)

// ShortUUID ...
func ShortUUID() string { return shortuuid.DefaultEncoder.Encode(uuid.Must(uuid.NewUUID())) }

// UUIDString ...
func UUIDString() string { return uuid.Must(uuid.NewUUID()).String() }
