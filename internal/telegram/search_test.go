package telegram

import (
	"testing"
	"time"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestParseSearchResult(t *testing.T) {
	t.Run("Valid torrent document", func(t *testing.T) {
		msg := &tg.Message{
			ID:  123,
			Out: false, // Sent by the bot
			Media: &tg.MessageMediaDocument{
				Document: &tg.Document{
					ID:            456,
					AccessHash:    789,
					FileReference: []byte("file_ref_token"),
					Size:          1024,
					Date:          1600000000,
					Attributes: []tg.DocumentAttributeClass{
						&tg.DocumentAttributeFilename{
							FileName: "Zelda_Tears_of_the_Kingdom.torrent",
						},
					},
				},
			},
		}

		result, ok := parseSearchResult(msg)
		assert.True(t, ok)
		assert.NotNil(t, result)
		assert.Equal(t, "Zelda_Tears_of_the_Kingdom.torrent", result.Name)
		assert.Equal(t, int64(1024), result.Size)
		assert.Equal(t, 123, result.MsgID)
		assert.Equal(t, time.Unix(1600000000, 0), result.Date)
		// 66696c655f7265665f746f6b656e is hex for "file_ref_token"
		assert.Equal(t, "456:789:66696c655f7265665f746f6b656e", result.FileID)
	})

	t.Run("Ignore non-torrent extensions", func(t *testing.T) {
		msg := &tg.Message{
			ID:  124,
			Out: false,
			Media: &tg.MessageMediaDocument{
				Document: &tg.Document{
					Attributes: []tg.DocumentAttributeClass{
						&tg.DocumentAttributeFilename{
							FileName: "Zelda.bin",
						},
					},
				},
			},
		}

		_, ok := parseSearchResult(msg)
		assert.False(t, ok)
	})

	t.Run("Ignore outgoing messages", func(t *testing.T) {
		msg := &tg.Message{
			ID:  125,
			Out: true, // Sent by us
			Media: &tg.MessageMediaDocument{
				Document: &tg.Document{
					Attributes: []tg.DocumentAttributeClass{
						&tg.DocumentAttributeFilename{
							FileName: "game.torrent",
						},
					},
				},
			},
		}

		_, ok := parseSearchResult(msg)
		assert.False(t, ok)
	})

	t.Run("No media in message", func(t *testing.T) {
		msg := &tg.Message{
			ID:    126,
			Out:   false,
			Media: nil,
		}

		_, ok := parseSearchResult(msg)
		assert.False(t, ok)
	})

	t.Run("No filename attribute", func(t *testing.T) {
		msg := &tg.Message{
			ID:  127,
			Out: false,
			Media: &tg.MessageMediaDocument{
				Document: &tg.Document{
					Attributes: []tg.DocumentAttributeClass{
						&tg.DocumentAttributeImageSize{
							W: 100,
							H: 100,
						},
					},
				},
			},
		}

		_, ok := parseSearchResult(msg)
		assert.False(t, ok)
	})
}

func TestExtractMessageID(t *testing.T) {
	t.Run("Updates type", func(t *testing.T) {
		upds := &tg.Updates{
			Updates: []tg.UpdateClass{
				&tg.UpdateMessageID{
					ID: 999,
				},
			},
		}
		id := extractMessageID(upds)
		assert.Equal(t, 999, id)
	})

	t.Run("UpdateShortMessage type", func(t *testing.T) {
		upd := &tg.UpdateShortMessage{
			ID: 888,
		}
		id := extractMessageID(upd)
		assert.Equal(t, 888, id)
	})

	t.Run("UpdateShort with UpdateNewMessage", func(t *testing.T) {
		upd := &tg.UpdateShort{
			Update: &tg.UpdateNewMessage{
				Message: &tg.Message{
					ID: 777,
				},
			},
		}
		id := extractMessageID(upd)
		assert.Equal(t, 777, id)
	})

	t.Run("Unknown updates type", func(t *testing.T) {
		id := extractMessageID(&tg.UpdateShortSentMessage{})
		assert.Equal(t, 0, id)
	})
}

func TestParseTextSearchResults(t *testing.T) {
	t.Run("Valid single game text search result", func(t *testing.T) {
		msg := &tg.Message{
			ID:   500,
			Out:  false,
			Date: 1700000000,
			Message: `Hatsune Miku Project DIVA Mega39's
Tamanho: 13.88 GB
Download: /download_12345 /cover_12345`,
		}

		res, ok := parseTextSearchResults(msg)
		assert.True(t, ok)
		assert.Len(t, res, 1)
		assert.Equal(t, "Hatsune Miku Project DIVA Mega39's.torrent", res[0].Name)
		assert.Equal(t, int64(14903536517), res[0].Size)
		assert.Equal(t, "/download_12345", res[0].FileID)
		assert.Equal(t, 500, res[0].MsgID)
	})

	t.Run("Valid multi-file torrent text search result", func(t *testing.T) {
		msg := &tg.Message{
			ID:   501,
			Out:  false,
			Date: 1700000000,
			Message: `Super Mario Bundle
No torrent: "mario1.nsp, mario2.nsp"
Tamanho: 500.5 MB
Download: /download_98765 /cover_98765`,
		}

		res, ok := parseTextSearchResults(msg)
		assert.True(t, ok)
		assert.Len(t, res, 1)
		assert.Equal(t, "Super Mario Bundle.torrent", res[0].Name)
		assert.Equal(t, int64(524812288), res[0].Size)
		assert.Equal(t, "/download_98765", res[0].FileID)
	})

	t.Run("Valid single message containing multiple games", func(t *testing.T) {
		msg := &tg.Message{
			ID:   503,
			Out:  false,
			Date: 1700000000,
			Message: `Para evitar problemas de instalação, use KEFIR!
Jogos com tag [🇧🇷MOD] têm mod de tradução PTBR na página "COVER".

✅Torrents encontrados para você: 10✅

Cadence of Hyrule - Crypt of the NecroDancer featuring The Legend of Zelda [🇧🇷MOD]
Tamanho: 2.66 GB
Download: /download411 /cover411

Ship of Harkinian (The Legend of Zelda - Ocarina of Time, native port) [🇧🇷MOD]
Tamanho: 1.17 GB
Download: /download5647 /cover5647`,
		}

		res, ok := parseTextSearchResults(msg)
		assert.True(t, ok)
		assert.Len(t, res, 2)

		assert.Equal(t, "Cadence of Hyrule - Crypt of the NecroDancer featuring The Legend of Zelda [🇧🇷MOD].torrent", res[0].Name)
		assert.Equal(t, int64(2856153251), res[0].Size)
		assert.Equal(t, "/download411", res[0].FileID)

		assert.Equal(t, "Ship of Harkinian (The Legend of Zelda - Ocarina of Time, native port) [🇧🇷MOD].torrent", res[1].Name)
		assert.Equal(t, int64(1256277934), res[1].Size)
		assert.Equal(t, "/download5647", res[1].FileID)
	})

	t.Run("Invalid message without download command", func(t *testing.T) {
		msg := &tg.Message{
			ID:   502,
			Out:  false,
			Message: `Some update notice
Please keep updated!`,
		}

		_, ok := parseTextSearchResults(msg)
		assert.False(t, ok)
	})
}

func TestParseSizeToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"13.88 GB", 14903536517},
		{"500 MB", 524288000},
		{"12.5KB", 12800},
		{"1024 B", 1024},
		{"10", 10},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res := parseSizeToBytes(tt.input)
			assert.Equal(t, tt.expected, res)
		})
	}
}
