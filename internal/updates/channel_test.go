package updates

import "testing"

func TestChannelString(t *testing.T) {
	tests := []struct {
		name     string
		channel  Channel
		expected string
	}{
		{"Dev Channel", ChannelDev, "https://github-build-artifacts.vitalvas.dev/vitalvas/oneauth/"},
		{"Prod Channel", ChannelProd, "https://oneauth-files.vitalvas.dev/release/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.channel.String()
			if result != test.expected {
				t.Errorf("Expected %s, but got %s", test.expected, result)
			}
		})
	}
}

func TestGetChannel(t *testing.T) {
	tests := []struct {
		version  string
		expected Channel
	}{
		{"v0.0.1", ChannelDev},
		{"v0.0.99", ChannelDev},
		{"v0.0.777777777", ChannelDev},
		{"v0.1.0", ChannelProd},
		{"v1.0.0", ChannelProd},
		{"v2.1.3", ChannelProd},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := getChannel(test.version)
			if result != test.expected {
				t.Errorf("Expected getChannel(%s) to return %v, but got %v", test.version, test.expected, result)
			}
		})
	}
}

func TestGetChannelName(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{"v0.0.1", "dev"},
		{"v0.0.1111111", "dev"},
		{"v0.1.0", "prod"},
		{"v1.0.0", "prod"},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := GetChannelName(test.version)
			if result != test.expected {
				t.Errorf("Expected GetChannelName(%s) to return '%s', but got '%s'", test.version, test.expected, result)
			}
		})
	}
}
