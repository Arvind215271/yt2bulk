package youtube

// PlayerClient defines the specific player client to be used with yt-dlp.
type PlayerClient string

const (
	PlayerClientDefault   PlayerClient = "default"
	PlayerClientAndroid   PlayerClient = "android"
	PlayerClientAndroidVR PlayerClient = "android_vr"
	PlayerClientVisionOS  PlayerClient = "visionos"
	PlayerClientTVSamsung PlayerClient = "tv_samsung"
	PlayerClientAll       PlayerClient = "all"
	// Clients with known extraction issues should be avoided or used with caution
	PlayerClientWeb       PlayerClient = "web"
	PlayerClientWebSafari PlayerClient = "web_safari"
	PlayerClientMWeb      PlayerClient = "mweb"
	PlayerClientIOS       PlayerClient = "ios"
)

// IsValid checks if the player client is supported for reliable extraction.
func (p PlayerClient) IsValid() bool {
	switch p {
	case PlayerClientAndroid, PlayerClientAndroidVR, PlayerClientVisionOS, PlayerClientTVSamsung, PlayerClientDefault, PlayerClientAll:
		return true
	default:
		return false
	}
}
