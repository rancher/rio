package resto

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/cli/cli/config"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	digest "github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	log "github.com/sirupsen/logrus"
)

// RegistryOptions contains optional configuration for Registry operations
type RegistryOptions struct {
	Username             string
	Password             string
	Insecure             bool
	CleartextCredentials bool
}

type unsupportedMediaType struct{}

func (u unsupportedMediaType) Error() string {
	return "Unsupported media type"
}

// ManifestAny is a manifest type for arbitrary configuration data
type ManifestAny struct {
	manifest.Versioned
	Payload string `json:"payload,omitempty"`
}

type parsedReference struct {
	domain string
	path   string
	tag    string
}

func parseRef(repoTag string) (parsedReference, error) {
	rawref, err := reference.ParseNormalizedNamed(repoTag)
	if err != nil {
		return parsedReference{}, err
	}
	ref, ok := rawref.(reference.Named)
	if !ok {
		return parseRef("docker.io/" + repoTag)
	}
	tag := "latest"
	if rt, ok := ref.(reference.Tagged); ok {
		tag = rt.Tag()
	}
	domain := reference.Domain(ref)
	if domain == "docker.io" {
		domain = "registry-1.docker.io"
	}
	return parsedReference{"https://" + domain, reference.Path(ref), tag}, nil
}

func getCredentials(domain string) (string, string, error) {
	cfg, err := config.Load("")
	if err != nil {
		return "", "", err
	}
	switch domain {
	case "https://registry-1.docker.io":
		domain = "https://index.docker.io/v1/"
	default:
		domain = strings.TrimPrefix(domain, "https://")
	}
	auth, err := cfg.GetAuthConfig(domain)
	if err != nil {
		return "", "", err
	}
	return auth.Username, auth.Password, nil
}

func makeTarGz(content map[string]string) ([]byte, digest.Digest, error) {
	buf := bytes.NewBuffer(nil)
	err := func() error {
		w := tar.NewWriter(buf)
		defer w.Close()
		for k, v := range content {
			if err := w.WriteHeader(&tar.Header{
				Typeflag: tar.TypeReg,
				Name:     k,
				Mode:     0600,
				Size:     int64(len(v)),
			}); err != nil {
				return err
			}
			if _, err := w.Write([]byte(v)); err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return nil, "", err
	}
	dgst := digest.SHA256.FromBytes(buf.Bytes())
	gzbuf := bytes.NewBuffer(nil)
	g := gzip.NewWriter(gzbuf)
	if _, err := g.Write(buf.Bytes()); err != nil {
		return nil, "", err
	}
	if err := g.Close(); err != nil {
		return nil, "", err
	}
	return gzbuf.Bytes(), dgst, nil
}

const maxRepositoryCount = 10000

// ListRepositories lists all the repositories in a registry
func ListRepositories(ctx context.Context, endpoint string, opts RegistryOptions) ([]string, error) {
	tr, err := NewTransportCatalog(endpoint, opts)
	if err != nil {
		return nil, err
	}
	registry, err := client.NewRegistry(endpoint, tr)
	if err != nil {
		return nil, err
	}
	entries := make([]string, maxRepositoryCount)
	count, err := registry.Repositories(ctx, entries, "")
	if err != nil && err != io.EOF {
		return nil, err
	}
	return entries[0:count], nil
}

// ListTags lists all the tags in a repository
func ListTags(ctx context.Context, reponame string, opts RegistryOptions) ([]string, error) {
	pr, err := parseRef(reponame)
	if err != nil {
		return nil, err
	}
	repo, err := NewRepository(ctx, pr.domain, pr.path, opts)
	if err != nil {
		return nil, err
	}
	tagService := repo.Tags(ctx)
	return tagService.All(ctx)
}

// PullConfig pulls a configuration file from a registry
func PullConfig(ctx context.Context, repoTag string, opts RegistryOptions) (string, error) {
	res, err := PullConfigMulti(ctx, repoTag, opts)
	if err != nil {
		return "", err
	}
	return res["config"], nil
}

// PullConfigMulti pulls a set of configuration files from a registry
func PullConfigMulti(ctx context.Context, repoTag string, opts RegistryOptions) (map[string]string, error) {
	pr, err := parseRef(repoTag)
	if err != nil {
		return nil, err
	}
	if opts.Username == "" {
		opts.Username, opts.Password, err = getCredentials(pr.domain)
		if err != nil {
			log.Debugf("failed to get credentials for %s: %s", pr.domain, err)
		}
	}
	repo, err := NewRepository(ctx, pr.domain, pr.path, opts)
	if err != nil {
		return nil, err
	}
	tagService := repo.Tags(ctx)
	dgst, err := tagService.Get(ctx, pr.tag)
	if err != nil {
		return nil, err
	}
	manifestService, err := repo.Manifests(ctx)
	if err != nil {
		return nil, err
	}
	manifest, err := manifestService.Get(ctx, dgst.Digest)
	if err != nil {
		return nil, err
	}
	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return nil, err
	}
	if mediaType == MediaTypeConfig {
		var ma ManifestAny
		if err := json.Unmarshal(payload, &ma); err != nil {
			return nil, err
		}
		res := make(map[string]string)
		err = json.Unmarshal([]byte(ma.Payload), &res)
		return res, err
	}
	// legacy image mode
	return pullConfigImage(ctx, manifest, repo)
}

func pullConfigImage(ctx context.Context, manifest distribution.Manifest, repo distribution.Repository) (map[string]string, error) {
	refs := manifest.References()
	if len(refs) != 2 {
		return nil, fmt.Errorf("expected 2 references, found %v", len(refs))
	}
	// assume second element is the layer (first being the image config)
	r := refs[1]
	rdgst := r.Digest
	blobsService := repo.Blobs(ctx)
	payloadGz, err := blobsService.Get(ctx, rdgst)
	if err != nil {
		return nil, err
	}
	payloadBuf := bytes.NewBuffer(payloadGz)
	gzf, err := gzip.NewReader(payloadBuf)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(gzf)
	return tarContent(tarReader)
}

func tarContent(tarReader *tar.Reader) (map[string]string, error) {
	res := make(map[string]string)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, err
		}
		if header.Typeflag == tar.TypeReg {
			content := bytes.NewBuffer(nil)
			io.Copy(content, tarReader)
			res[header.Name] = content.String()
		}
	}
	return res, nil
}

// PushConfig pushes a configuration file to a registry and returns its digest
func PushConfig(ctx context.Context, payload, repoTag string, opts RegistryOptions, labels map[string]string) (string, error) {
	return PushConfigMulti(ctx, map[string]string{
		"config": payload,
	}, repoTag, opts, labels)
}

// PushConfigMulti pushes a set of configuration files to a registry and returns its digest
func PushConfigMulti(ctx context.Context, payload map[string]string, repoTag string, opts RegistryOptions, labels map[string]string) (string, error) {
	pr, err := parseRef(repoTag)
	if err != nil {
		return "", err
	}
	if opts.Username == "" {
		opts.Username, opts.Password, err = getCredentials(pr.domain)
		if err != nil {
			log.Debugf("failed to get credentials for %s: %s", pr.domain, err)
		}
	}
	repo, err := NewRepository(ctx, pr.domain, pr.path, opts)
	if err != nil {
		return "", err
	}
	digest, err := pushConfigMediaType(ctx, payload, pr, repo)
	if err == nil {
		return digest, err
	}
	if _, ok := err.(unsupportedMediaType); ok {
		return pushConfigLegacy(ctx, payload, pr, repo, labels)
	}
	return digest, err
}

func pushConfigMediaType(ctx context.Context, payload map[string]string, pr parsedReference, repo distribution.Repository) (string, error) {
	j, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	manifestAny := ManifestAny{
		Versioned: manifest.Versioned{
			SchemaVersion: 2,
			MediaType:     MediaTypeConfig,
		},
		Payload: string(j),
	}
	raw, err := json.Marshal(manifestAny)
	if err != nil {
		return "", err
	}
	manifestService, err := repo.Manifests(ctx)
	if err != nil {
		return "", err
	}
	manifest := NewConfigManifest(MediaTypeConfig, raw)
	dgst, err := manifestService.Put(ctx, manifest, distribution.WithTag(pr.tag))
	if err == nil {
		return dgst.String(), nil
	}
	switch {
	case strings.Contains(err.Error(), "manifest invalid"):
		return "", unsupportedMediaType{}
	case strings.Contains(err.Error(), "manifest Unknown"):
		return "", unsupportedMediaType{}
	default:
		return "", err
	}
}

func pushConfigLegacy(ctx context.Context, payload map[string]string, pr parsedReference, repo distribution.Repository, labels map[string]string) (string, error) {
	manifestService, err := repo.Manifests(ctx)
	if err != nil {
		return "", err
	}
	// try legacy mode
	// create payload
	payloadGz, payloadUncompressedDigest, err := makeTarGz(payload)
	if err != nil {
		return "", err
	}
	blobsService := repo.Blobs(ctx)
	payloadDesc, err := blobsService.Put(ctx, schema2.MediaTypeLayer, payloadGz)
	if err != nil {
		return "", err
	}
	payloadDesc.MediaType = schema2.MediaTypeLayer
	// create dummy image config
	now := time.Now()
	imageConfig := ociv1.Image{
		Created:      &now,
		Architecture: "config",
		OS:           "config",
		Config: ociv1.ImageConfig{
			Labels: labels,
		},
		RootFS: ociv1.RootFS{
			Type:    "layers",
			DiffIDs: []digest.Digest{payloadUncompressedDigest},
		},
		History: []ociv1.History{
			{CreatedBy: "COPY configfile /"},
		},
	}
	icm, err := json.Marshal(imageConfig)
	if err != nil {
		return "", err
	}
	icDesc, err := blobsService.Put(ctx, schema2.MediaTypeImageConfig, icm)
	if err != nil {
		return "", err
	}
	icDesc.MediaType = schema2.MediaTypeImageConfig
	man := schema2.Manifest{
		Versioned: schema2.SchemaVersion,
		Config:    icDesc,
		Layers:    []distribution.Descriptor{payloadDesc},
	}
	dman, err := schema2.FromStruct(man)
	if err != nil {
		return "", err
	}
	dgst, err := manifestService.Put(ctx, dman, distribution.WithTag(pr.tag))
	return dgst.String(), err
}
