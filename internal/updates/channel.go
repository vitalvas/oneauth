package updates

import "golang.org/x/mod/semver"

type Channel string

func (c *Channel) String() string {
	return string(*c)
}

const (
	ChannelDev  Channel = "https://github-build-artifacts.vitalvas.dev/vitalvas/oneauth/"
	ChannelProd Channel = "https://oneauth-files.vitalvas.dev/release/"
)

func getChannel(version string) Channel {
	if semver.MajorMinor(version) != "v0.0" {
		return ChannelProd
	}

	return ChannelDev
}

func GetChannelName(version string) string {
	if getChannel(version) == ChannelProd {
		return "prod"
	}

	return "dev"
}
