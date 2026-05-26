package switchutil

import (
	"testing"
)

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Phoenix Wright - Ace Attorney Trilogy [🇧🇷MOD].torrent",
			expected: "Phoenix Wright - Ace Attorney Trilogy.torrent",
		},
		{
			input:    "Cadence of Hyrule - Crypt of the NecroDancer featuring The Legend of Zelda [??MOD].torrent",
			expected: "Cadence of Hyrule - Crypt of the NecroDancer featuring The Legend of Zelda.torrent",
		},
		{
			input:    "Super Mario Odyssey [0100000000010000] [v0].torrent",
			expected: "Super Mario Odyssey.torrent",
		},
		{
			input:    "Some Game [Update] [USA] [v12345].torrent",
			expected: "Some Game.torrent",
		},
		{
			input:    "The Great Ace Attorney Chronicles.torrent",
			expected: "The Great Ace Attorney Chronicles.torrent",
		},
		{
			input:    "NoExtension [USA] [v0]",
			expected: "NoExtension",
		},
		{
			input:    "   Game   [Tag]   Name  [Tag2]  .torrent",
			expected: "Game Name.torrent",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			actual := CleanTitle(tc.input)
			if actual != tc.expected {
				t.Errorf("CleanTitle(%q) = %q; expected %q", tc.input, actual, tc.expected)
			}
		})
	}
}
