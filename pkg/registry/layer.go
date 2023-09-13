package registry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	distribution "github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/ocischema"
	"github.com/distribution/distribution/v3/registry/client/auth"
	digest "github.com/opencontainers/go-digest"
	"github.com/openshift/library-go/pkg/image/reference"
	"github.com/openshift/library-go/pkg/image/registryclient"
)

type Client struct {
	credentials auth.CredentialStore
	ref         reference.DockerImageReference

	rwContext         *registryclient.Context
	roContext         *registryclient.Context
	transport         http.RoundTripper
	insecureTransport http.RoundTripper

	insecure bool
	userPass userPass
}

type Option func(*Client)

func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) {
		c.transport = t
	}
}

func WithInsecureTransport(t http.RoundTripper) Option {
	return func(c *Client) {
		c.insecureTransport = t
	}
}

func WithInsecure(insecure bool) Option {
	return func(c *Client) {
		c.insecure = insecure
	}
}

func WithUserPass(username string, password string) Option {
	return func(c *Client) {
		c.userPass = userPass{
			Username: username,
			Password: password,
		}
	}
}

func NewClient(image string, opts ...Option) (*Client, error) {
	ref, err := reference.Parse(image)
	if err != nil {
		return nil, err
	}

	c := &Client{
		ref:               ref,
		transport:         http.DefaultTransport,
		insecureTransport: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(c)
	}
	credentialStore, err := newBasicCredentials(c.userPass, ref.Registry, c.insecure)
	if err != nil {
		return nil, err
	}
	c.credentials = credentialStore
	return c, nil
}

func (c *Client) rw() *registryclient.Context {
	if c.rwContext == nil {
		c.rwContext = registryclient.NewContext(c.transport, c.insecureTransport).
			WithCredentials(c.credentials).
			WithActions("pull", "push")
	}
	return c.rwContext
}

func (c *Client) ro() *registryclient.Context {
	if c.roContext == nil {
		c.roContext = registryclient.NewContext(c.transport, c.insecureTransport).
			WithCredentials(c.credentials)
	}
	return c.roContext
}

func (c *Client) Put(ctx context.Context, r io.Reader) error {
	repo, err := c.rw().RepositoryForRef(ctx, c.ref, c.insecure)
	if err != nil {
		return err
	}

	blobs := repo.Blobs(ctx)

	writer, err := blobs.Create(ctx)
	if err != nil {
		return err
	}

	dgstr := digest.Canonical.Digester()
	n, err := io.Copy(writer, io.TeeReader(r, dgstr.Hash()))
	if err != nil {
		return err
	}

	desc := distribution.Descriptor{
		Size:   n,
		Digest: dgstr.Digest(),
	}

	desc, err = writer.Commit(ctx, desc)
	if err != nil {
		return err
	}

	data, err := json.Marshal(ocischema.Manifest{
		Versioned: ocischema.SchemaVersion,
		Config:    desc,
	})
	if err != nil {
		return err
	}

	m, err := ocischema.NewManifestBuilder(blobs, data, nil).Build(ctx)
	if err != nil {
		return err
	}

	ms, err := repo.Manifests(ctx)
	if err != nil {
		return err
	}

	_, err = ms.Put(ctx, m, distribution.WithTag(c.ref.Tag))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Get(ctx context.Context) (io.ReadSeekCloser, error) {
	repo, err := c.ro().RepositoryForRef(ctx, c.ref, c.insecure)
	if err != nil {
		return nil, err
	}

	ms, err := repo.Manifests(ctx)
	if err != nil {
		return nil, err
	}

	dm, err := ms.Get(ctx, "", distribution.WithTag(c.ref.Tag))
	if err != nil {
		return nil, err
	}

	refs := dm.References()[0]

	blobs := repo.Blobs(ctx)

	blob, err := blobs.Get(ctx, refs.Digest)
	if err != nil {
		return nil, err
	}

	m := ocischema.Manifest{}
	err = json.Unmarshal(blob, &m)
	if err != nil {
		return nil, err
	}

	target := m.Target()

	return blobs.Open(ctx, target.Digest)
}
