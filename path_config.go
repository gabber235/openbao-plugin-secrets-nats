package natsbackend

import (
	"context"
	"net/http"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

const backendConfigPath = "config"

type backendConfigEntry struct {
	ImplicitRevocationsEnabled bool `json:"implicit_revocations_enabled"`
}

func defaultBackendConfig() *backendConfigEntry {
	return &backendConfigEntry{
		ImplicitRevocationsEnabled: false,
	}
}

func (backendConfigEntry) configPath() string {
	return backendConfigPath
}

func (b *backend) Config(ctx context.Context, s logical.Storage) (*backendConfigEntry, error) {
	cfg := defaultBackendConfig()

	var stored *backendConfigEntry
	if err := get(ctx, s, backendConfigPath, &stored); err != nil {
		return nil, err
	}
	if stored != nil {
		cfg.ImplicitRevocationsEnabled = stored.ImplicitRevocationsEnabled
	}

	return cfg, nil
}

func pathConfigBackend(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: backendConfigPath + "$",
			Fields: map[string]*framework.FieldSchema{
				"implicit_revocations_enabled": {
					Type:        framework.TypeBool,
					Description: "Whether revoking an OpenBao lease for nats/creds should create a NATS account revocation. Defaults to false.",
					Required:    false,
					Default:     false,
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathConfigWrite,
					Responses: map[int][]framework.Response{
						http.StatusOK: {{Description: "OK"}},
					},
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathConfigWrite,
					Responses: map[int][]framework.Response{
						http.StatusOK: {{Description: "OK"}},
					},
				},
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathConfigRead,
					Responses: map[int][]framework.Response{
						http.StatusOK: {{
							Description: "OK",
							Fields: map[string]*framework.FieldSchema{
								"implicit_revocations_enabled": {
									Type:        framework.TypeBool,
									Description: "Whether revoking an OpenBao lease for nats/creds creates a NATS account revocation.",
									Required:    true,
								},
							},
						}},
					},
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathConfigDelete,
					Responses: map[int][]framework.Response{
						http.StatusNoContent: {{Description: "No Content"}},
					},
				},
			},
			HelpSynopsis:    "Configure global behavior for the NATS secrets engine.",
			HelpDescription: "Configure plugin-wide behavior, including whether OpenBao lease revocation of user credentials should create NATS account revocations.",
		},
	}
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	cfg, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if v, ok := d.GetOk("implicit_revocations_enabled"); ok {
		cfg.ImplicitRevocationsEnabled = v.(bool)
	}

	if err := storeInStorage(ctx, req.Storage, backendConfigPath, cfg); err != nil {
		return nil, err
	}

	return &logical.Response{}, nil
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	cfg, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	return &logical.Response{Data: map[string]any{
		"implicit_revocations_enabled": cfg.ImplicitRevocationsEnabled,
	}}, nil
}

func (b *backend) pathConfigDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, backendConfigPath); err != nil {
		return nil, err
	}

	return &logical.Response{}, nil
}
