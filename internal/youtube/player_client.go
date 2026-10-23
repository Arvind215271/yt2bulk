package youtube

// PlayerClient defines the specific player client to be used with yt-dlp.
type PlayerClient string

const (
	PlayerClientDefault        PlayerClient = "default"
	PlayerClientWeb            PlayerClient = "web"
	PlayerClientWebSafari      PlayerClient = "web_safari"
	PlayerClientWebEmbedded    PlayerClient = "web_embedded"
	PlayerClientMWeb           PlayerClient = "mweb"
	PlayerClientIOS            PlayerClient = "ios"
	PlayerClientAndroid        PlayerClient = "android"
	PlayerClientAndroidVR      PlayerClient = "android_vr"
	PlayerClientAndroidSDKLess PlayerClient = "android_sdkless"
	PlayerClientTV             PlayerClient = "tv"
	PlayerClientTVSimply       PlayerClient = "tv_simply"
	PlayerClientTVEmbedded     PlayerClient = "tv_embedded"
	PlayerClientVisionOS       PlayerClient = "visionos"
	PlayerClientTVSamsung      PlayerClient = "tv_samsung"
	PlayerClientAll            PlayerClient = "all"
)

// IsValid checks if the player client is supported for reliable extraction.
func (p PlayerClient) IsValid() bool {
	switch p {
	case PlayerClientDefault,
		PlayerClientWeb, PlayerClientWebSafari, PlayerClientWebEmbedded,
		PlayerClientMWeb, PlayerClientIOS,
		PlayerClientAndroid, PlayerClientAndroidVR, PlayerClientAndroidSDKLess,
		PlayerClientTV, PlayerClientTVSimply, PlayerClientTVEmbedded,
		PlayerClientVisionOS, PlayerClientTVSamsung,
		PlayerClientAll:
		return true
	default:
		return false
	}
}
