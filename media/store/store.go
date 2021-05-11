package store

import (
	"encoding/hex"
	"io"
	"strconv"
	"strings"

	"github.com/timemore/foundation/errors"
	"github.com/timemore/foundation/media"
	"golang.org/x/crypto/blake2b"
)

type Store struct {
	config        Config
	serviceClient Service
}

func New(config Config) (*Store, error) {
	if len(config.Modules) == 0 {
		return nil, errors.ArgMsg("config.Modules", "empty")
	}
	if config.StoreService == "" {
		return nil, errors.ArgMsg("config.StoreService", "empty")
	}

	modCfg := config.Modules[config.StoreService]
	if modCfg == nil {
		return nil, errors.ArgMsg("config.StoreService", config.StoreService+" not configured")
	}
	serviceClient, err := NewServiceClient(config.StoreService, modCfg)
	if err != nil {
		return nil, errors.ArgWrap("config.StoreService", config.StoreService+" initialization failed", err)
	}

	return &Store{
		config:        config,
		serviceClient: serviceClient,
	}, nil
}

func (mediaStore *Store) Upload(mediaName string, contentSource io.Reader, mediaType media.MediaType) (publicURL string, err error) {
	objectURL, err := mediaStore.serviceClient.PutObject(mediaName, contentSource)
	if err != nil {
		return "", errors.Wrap("putting object", err)
	}

	if mediaType == media.MediaType_IMAGE {
		if mediaStore.config.ImagesBaseURL != "" {
			publicURL = strings.TrimRight(mediaStore.config.ImagesBaseURL, "/") + "/" + mediaName
		} else {
			publicURL = objectURL
		}
	} else {
		publicURL = objectURL
	}
	return publicURL, nil
}

const nameGenHashLength = 16

const nameGenKeyDefault = "N0k3y"

func (mediaStore *Store) GenerateName(stream io.Reader) string {
	keyBytes := []byte(mediaStore.config.NameGenerationKey)
	if len(keyBytes) == 0 {
		keyBytes = []byte(nameGenKeyDefault)
	}
	hasher, err := blake2b.New(nameGenHashLength, keyBytes)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 1*1024*1024)
	dataSize := 0
	for {
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		dataSize += n
		hasher.Write(buf)
		if err == io.EOF || n == 0 {
			break
		}
	}

	hashBytes := hasher.Sum(nil)

	encodedHash := hex.EncodeToString(hashBytes) + "K" + hex.EncodeToString(keyBytes[:4]) + "N" + strconv.FormatInt(int64(dataSize), 16)

	return encodedHash
}

func ConfigSkeleton() Config {
	return Config{
		Modules: ModuleConfigSkeletons(),
	}
}

func ContentTypeInList(contentType string, contentTypeList []string) bool {
	for _, ct := range contentTypeList {
		if ct == contentType {
			return true
		}
	}
	return false
}
