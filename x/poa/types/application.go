package types

import "github.com/cosmos/cosmos-sdk/codec"

// NewApplication creates a new Application instance.
func NewApplication(validator Validator) Application {
	return Application{
		Validator: validator,
	}
}

// MustMarshalApplication marshals an application to bytes. It panics on error.
func MustMarshalApplication(cdc codec.BinaryCodec, v Application) []byte {
	return cdc.MustMarshal(&v)
}

// MustUnmarshalApplication unmarshals an application from bytes. It panics on error.
func MustUnmarshalApplication(cdc codec.BinaryCodec, value []byte) Application {
	application, err := UnmarshalApplication(cdc, value)
	if err != nil {
		panic(err)
	}

	return application
}

// UnmarshalApplication unmarshals an application from bytes.
func UnmarshalApplication(cdc codec.BinaryCodec, value []byte) (v Application, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}
