package telegram

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// SearchResult represents a parsed torrent search result.
type SearchResult struct {
	ID     int
	Name   string
	Size   int64
	Date   time.Time
	MsgID  int
	FileID string
}

// Search resolves the Search Bot, sends a search query, polls the chat history,
// and parses the response to return matching .torrent files.
func (w *ClientWrapper) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	botUsername := w.cfg.Telegram.SearchBot
	w.logger.Info("Resolving Search Bot username", zap.String("username", botUsername))

	resolved, err := w.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: botUsername,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve username %s: %w", botUsername, err)
	}

	var targetUser *tg.User
	for _, u := range resolved.Users {
		if tu, ok := u.(*tg.User); ok {
			if peerUser, ok := resolved.Peer.(*tg.PeerUser); ok && tu.ID == peerUser.UserID {
				targetUser = tu
				break
			}
		}
	}

	if targetUser == nil {
		return nil, fmt.Errorf("failed to locate user details in resolved users for %s", botUsername)
	}

	inputPeer := &tg.InputPeerUser{
		UserID:     targetUser.ID,
		AccessHash: targetUser.AccessHash,
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random message ID: %w", err)
	}
	randomID := n.Int64()

	w.logger.Info("Sending search query to bot", zap.String("query", query), zap.String("bot", botUsername))
	sentMsgClass, err := w.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  query,
		RandomID: randomID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message to bot %s: %w", botUsername, err)
	}

	queryMsgID := extractMessageID(sentMsgClass)
	w.logger.Debug("Initial extracted query message ID", zap.Int("msg_id", queryMsgID))

	// Polling configuration
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, errors.New("timeout waiting for search bot response")
		case <-ticker.C:
			history, err := w.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:  inputPeer,
				Limit: 20,
			})
			if err != nil {
				w.logger.Warn("Failed to get history from Telegram", zap.Error(err))
				continue
			}

			var messages []tg.MessageClass
			switch h := history.(type) {
			case *tg.MessagesMessages:
				messages = h.Messages
			case *tg.MessagesMessagesSlice:
				messages = h.Messages
			case *tg.MessagesChannelMessages:
				messages = h.Messages
			default:
				continue
			}

			// Fallback: if we could not extract queryMsgID from the send response, find it in history
			if queryMsgID == 0 {
				for _, m := range messages {
					if msg, ok := m.(*tg.Message); ok {
						if msg.Out && msg.Message == query {
							queryMsgID = msg.ID
							w.logger.Debug("Found query message ID in history", zap.Int("msg_id", queryMsgID))
							break
						}
					}
				}
			}

			if queryMsgID == 0 {
				w.logger.Debug("Query message not yet visible in history, waiting...")
				continue
			}

			var tempResults []SearchResult
			for i := len(messages) - 1; i >= 0; i-- {
				msg, ok := messages[i].(*tg.Message)
				if !ok {
					continue
				}
				// Look for bot response messages that are newer than our query message
				if !msg.Out && msg.ID > queryMsgID {
					if results, ok := parseTextSearchResults(msg); ok {
						tempResults = append(tempResults, results...)
					} else if res, ok := parseSearchResult(msg); ok {
						tempResults = append(tempResults, *res)
					}
				}
			}

			if len(tempResults) > 0 {
				for i := range tempResults {
					tempResults[i].ID = i + 1
				}
				w.logger.Info("Search results retrieved successfully", zap.Int("count", len(tempResults)))
				return tempResults, nil
			}
		}
	}
}

// extractMessageID attempts to extract the sent message ID from tg.UpdatesClass.
func extractMessageID(updates tg.UpdatesClass) int {
	switch u := updates.(type) {
	case *tg.Updates:
		for _, upd := range u.Updates {
			switch uu := upd.(type) {
			case *tg.UpdateMessageID:
				return uu.ID
			case *tg.UpdateNewMessage:
				if msg, ok := uu.Message.(*tg.Message); ok {
					return msg.ID
				}
			}
		}
	case *tg.UpdateShortMessage:
		return u.ID
	case *tg.UpdateShort:
		switch uu := u.Update.(type) {
		case *tg.UpdateMessageID:
			return uu.ID
		case *tg.UpdateNewMessage:
			if msg, ok := uu.Message.(*tg.Message); ok {
				return msg.ID
			}
		}
	}
	return 0
}

// parseSearchResult parses a message to extract a search result if it contains a valid .torrent document.
func parseSearchResult(msg *tg.Message) (*SearchResult, bool) {
	if msg.Out {
		return nil, false
	}
	if msg.Media == nil {
		return nil, false
	}
	mediaDoc, ok := msg.Media.(*tg.MessageMediaDocument)
	if !ok {
		return nil, false
	}
	doc, ok := mediaDoc.Document.(*tg.Document)
	if !ok {
		return nil, false
	}

	var filename string
	for _, attrClass := range doc.Attributes {
		if attr, ok := attrClass.(*tg.DocumentAttributeFilename); ok {
			filename = attr.FileName
			break
		}
	}

	if filename == "" {
		return nil, false
	}

	// Validate extension is exactly .torrent
	if !strings.HasSuffix(strings.ToLower(filename), ".torrent") {
		return nil, false
	}

	// Serialized FileID: id:access_hash:file_reference_hex
	fileID := fmt.Sprintf("%d:%d:%x", doc.ID, doc.AccessHash, doc.FileReference)

	return &SearchResult{
		Name:   filename,
		Size:   doc.Size,
		Date:   time.Unix(int64(doc.Date), 0),
		MsgID:  msg.ID,
		FileID: fileID,
	}, true
}

// parseTextSearchResults parses a text message containing potentially multiple game details and download commands.
func parseTextSearchResults(msg *tg.Message) ([]SearchResult, bool) {
	if msg.Out {
		return nil, false
	}
	text := msg.Message
	if text == "" {
		return nil, false
	}

	// The message must contain a download command
	if !strings.Contains(text, "/download") {
		return nil, false
	}

	var results []SearchResult
	lines := strings.Split(text, "\n")
	var currentBlock []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(currentBlock) > 0 {
				if res, ok := parseBlock(currentBlock, msg.Date, msg.ID); ok {
					results = append(results, res)
				}
				currentBlock = nil
			}
		} else {
			currentBlock = append(currentBlock, trimmed)
		}
	}
	if len(currentBlock) > 0 {
		if res, ok := parseBlock(currentBlock, msg.Date, msg.ID); ok {
			results = append(results, res)
		}
	}

	if len(results) == 0 {
		return nil, false
	}

	return results, true
}

func parseBlock(block []string, date int, msgID int) (SearchResult, bool) {
	var downloadCmd string
	var sizeBytes int64
	var title string

	hasDownload := false
	for _, line := range block {
		if strings.Contains(line, "/download") {
			words := strings.Fields(line)
			for _, w := range words {
				if strings.HasPrefix(w, "/download") {
					downloadCmd = w
					hasDownload = true
					break
				}
			}
		}
	}

	if !hasDownload || downloadCmd == "" {
		return SearchResult{}, false
	}

	title = strings.TrimSpace(block[0])

	for _, line := range block {
		if strings.HasPrefix(line, "Tamanho:") {
			sizeStr := strings.TrimSpace(strings.TrimPrefix(line, "Tamanho:"))
			sizeBytes = parseSizeToBytes(sizeStr)
			break
		}
	}

	if !strings.HasSuffix(strings.ToLower(title), ".torrent") {
		title = title + ".torrent"
	}

	return SearchResult{
		Name:   title,
		Size:   sizeBytes,
		Date:   time.Unix(int64(date), 0),
		MsgID:  msgID,
		FileID: downloadCmd,
	}, true
}

// parseSizeToBytes parses size strings (e.g. "13.88 GB", "500 MB") to bytes.
func parseSizeToBytes(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	var val float64
	var unit string
	n, _ := fmt.Sscanf(sizeStr, "%f %s", &val, &unit)
	if n < 2 {
		var valFallback float64
		var unitFallback string
		nFallback, _ := fmt.Sscanf(sizeStr, "%f%s", &valFallback, &unitFallback)
		if nFallback == 2 {
			val = valFallback
			unit = unitFallback
		} else if nFallback == 1 && n < 1 {
			val = valFallback
			unit = ""
		}
	}

	switch unit {
	case "GB", "G":
		return int64(val * 1024 * 1024 * 1024)
	case "MB", "M":
		return int64(val * 1024 * 1024)
	case "KB", "K":
		return int64(val * 1024)
	case "B":
		return int64(val)
	default:
		return int64(val)
	}
}
