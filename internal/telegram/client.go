package telegram

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/matheuszin/telegram-tinfoil-downloader/internal/config"
	"go.uber.org/zap"
)

// ClientWrapper wraps the gotd Telegram client to orchestrate authentication and run lifecycle.
type ClientWrapper struct {
	client *telegram.Client
	cfg    *config.Config
	logger *zap.Logger
	stdin  io.Reader
}

// NewClientWrapper instantiates a ClientWrapper, ensuring the session storage path exists.
func NewClientWrapper(cfg *config.Config, logger *zap.Logger) (*ClientWrapper, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	sessionDir := filepath.Dir(cfg.Telegram.SessionPath)
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	storage := &session.FileStorage{
		Path: cfg.Telegram.SessionPath,
	}

	client := telegram.NewClient(cfg.Telegram.ApiID, cfg.Telegram.ApiHash, telegram.Options{
		SessionStorage: storage,
		Logger:         logger,
	})

	return &ClientWrapper{
		client: client,
		cfg:    cfg,
		logger: logger,
		stdin:  os.Stdin,
	}, nil
}

// Run starts the Telegram client connection, completes the authentication flow if necessary, and executes the callback f.
func (w *ClientWrapper) Run(ctx context.Context, f func(ctx context.Context) error) error {
	flow := auth.NewFlow(
		&interactiveAuthenticator{defaultPhone: w.cfg.Telegram.Phone, stdin: w.stdin},
		auth.SendCodeOptions{},
	)

	w.logger.Info("Starting Telegram client connection...")
	err := w.client.Run(ctx, func(ctx context.Context) error {
		w.logger.Info("Telegram client connected. Checking authentication status...")
		if err := w.client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		w.logger.Info("Authentication successful.")
		return f(ctx)
	})
	if err != nil {
		return fmt.Errorf("telegram client execution error: %w", err)
	}
	return nil
}

// API returns the MTProto RPC client. Only valid to call inside the Run callback.
func (w *ClientWrapper) API() *tg.Client {
	return w.client.API()
}

// DownloadFile downloads a document specified by the serialized FileID (or command string) and writes its contents to the provided writer.
func (w *ClientWrapper) DownloadFile(ctx context.Context, fileID string, writer io.Writer) error {
	if fileID == "" {
		return errors.New("file ID cannot be empty")
	}

	// Case 1: fileID is a command (e.g. /download_12345)
	if strings.HasPrefix(fileID, "/") {
		return w.downloadFileViaCommand(ctx, fileID, writer)
	}

	// Case 2: Standard fileID format (id:access_hash:file_ref)
	parts := strings.Split(fileID, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid file ID format: expected 3 parts, got %d", len(parts))
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid document ID in file ID: %w", err)
	}

	accessHash, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid access hash in file ID: %w", err)
	}

	fileRef, err := hex.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("invalid file reference hex in file ID: %w", err)
	}

	loc := &tg.InputDocumentFileLocation{
		ID:            id,
		AccessHash:    accessHash,
		FileReference: fileRef,
	}

	d := downloader.NewDownloader()

	_, err = d.Download(w.API(), loc).Stream(ctx, writer)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func (w *ClientWrapper) downloadFileViaCommand(ctx context.Context, cmd string, writer io.Writer) error {
	botUsername := w.cfg.Telegram.SearchBot
	w.logger.Info("Resolving Search Bot username for download", zap.String("username", botUsername))

	resolved, err := w.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: botUsername,
	})
	if err != nil {
		return fmt.Errorf("failed to resolve username %s: %w", botUsername, err)
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
		return fmt.Errorf("failed to locate user details in resolved users for %s", botUsername)
	}

	inputPeer := &tg.InputPeerUser{
		UserID:     targetUser.ID,
		AccessHash: targetUser.AccessHash,
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return fmt.Errorf("failed to generate random message ID: %w", err)
	}
	randomID := n.Int64()

	w.logger.Info("Sending download command to bot", zap.String("command", cmd), zap.String("bot", botUsername))
	sentMsgClass, err := w.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  cmd,
		RandomID: randomID,
	})
	if err != nil {
		return fmt.Errorf("failed to send command to bot %s: %w", botUsername, err)
	}

	cmdMsgID := extractMessageID(sentMsgClass)
	w.logger.Debug("Sent download command message ID", zap.Int("msg_id", cmdMsgID))

	// Polling for the document response from the bot
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return errors.New("timeout waiting for bot to reply with the torrent file")
		case <-ticker.C:
			history, err := w.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:  inputPeer,
				Limit: 10,
			})
			if err != nil {
				w.logger.Warn("Failed to get history from Telegram during download", zap.Error(err))
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

			if cmdMsgID == 0 {
				for _, m := range messages {
					if msg, ok := m.(*tg.Message); ok {
						if msg.Out && msg.Message == cmd {
							cmdMsgID = msg.ID
							break
						}
					}
				}
			}

			if cmdMsgID == 0 {
				continue
			}

			var targetDoc *tg.Document
			for _, m := range messages {
				msg, ok := m.(*tg.Message)
				if !ok {
					continue
				}
				if !msg.Out && msg.ID > cmdMsgID {
					if msg.Media != nil {
						if mediaDoc, ok := msg.Media.(*tg.MessageMediaDocument); ok {
							if doc, ok := mediaDoc.Document.(*tg.Document); ok {
								var filename string
								for _, attrClass := range doc.Attributes {
									if attr, ok := attrClass.(*tg.DocumentAttributeFilename); ok {
										filename = attr.FileName
										break
									}
								}
								if strings.HasSuffix(strings.ToLower(filename), ".torrent") {
									targetDoc = doc
									break
								}
							}
						}
					}
				}
			}

			if targetDoc != nil {
				w.logger.Info("Found torrent document in bot response. Starting download...")
				loc := &tg.InputDocumentFileLocation{
					ID:            targetDoc.ID,
					AccessHash:    targetDoc.AccessHash,
					FileReference: targetDoc.FileReference,
				}

				d := downloader.NewDownloader()
				_, err = d.Download(w.API(), loc).Stream(ctx, writer)
				if err != nil {
					return fmt.Errorf("failed to stream document: %w", err)
				}
				return nil
			}
		}
	}
}

// SetStdin sets the input source for the interactive authentication flow (primarily for testing).
func (w *ClientWrapper) SetStdin(stdin io.Reader) {
	w.stdin = stdin
}

// interactiveAuthenticator implements auth.UserAuthenticator to interactively read credentials from stdin.
type interactiveAuthenticator struct {
	defaultPhone string
	stdin        io.Reader
}

func readInput(prompt string, stdin io.Reader) (string, error) {
	fmt.Print(prompt)
	if stdin == nil {
		stdin = os.Stdin
	}
	reader := bufio.NewReader(stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func (a *interactiveAuthenticator) Phone(ctx context.Context) (string, error) {
	phone, err := readInput(fmt.Sprintf("Enter Telegram Phone Number (default: %s): ", a.defaultPhone), a.stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read phone number: %w", err)
	}
	if phone == "" {
		return a.defaultPhone, nil
	}
	return phone, nil
}

func (a *interactiveAuthenticator) Password(ctx context.Context) (string, error) {
	password, err := readInput("Enter Telegram 2FA Password: ", a.stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read 2FA password: %w", err)
	}
	return password, nil
}

func (a *interactiveAuthenticator) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	// Auto-accept terms of service
	return nil
}

func (a *interactiveAuthenticator) SignUp(ctx context.Context) (auth.UserInfo, error) {
	// Sign-up is out of scope for this utility
	return auth.UserInfo{}, errors.New("sign up not supported: please register your account on an official Telegram client first")
}

func (a *interactiveAuthenticator) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	code, err := readInput("Enter Telegram Verification Code: ", a.stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read verification code: %w", err)
	}
	return code, nil
}
