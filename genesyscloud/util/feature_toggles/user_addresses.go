package feature_toggles

import "os"

const newUserAddressesLogicEnvToggle = "USE_NEW_USER_ADDRESS_LOGIC"

func NewUserAddressesLogicToggleName() string {
	return newUserAddressesLogicEnvToggle
}

func NewUserAddressesLogicExists() bool {
	var exists bool
	_, exists = os.LookupEnv(newUserAddressesLogicEnvToggle)
	return exists
}
