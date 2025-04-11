package feature_toggles

const mrMoToggle = "MRMO_IS_ACTIVE"

func MrMoIsActive() bool {
	return envVarIsSet(mrMoToggle)
}