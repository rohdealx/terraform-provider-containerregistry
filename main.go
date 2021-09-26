package main

import (
	"context"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func buildAuthenticator(auth map[string]interface{}) authn.Authenticator {
	config := authn.AuthConfig{
		Username:      auth["username"].(string),
		Password:      auth["password"].(string),
		Auth:          auth["auth"].(string),
		IdentityToken: auth["identity_token"].(string),
		RegistryToken: auth["registry_token"].(string),
	}
	return authn.FromConfig(config)
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	authenticators := make(map[string]authn.Authenticator)
	for _, e := range d.Get("auth").([]interface{}) {
		auth := e.(map[string]interface{})
		authenticators[auth["address"].(string)] = buildAuthenticator(auth)
	}
	return authenticators, nil
}

func dataSourceContainerRegistryImageRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ref, err := name.ParseReference(data.Get("name").(string))
	if err != nil {
		return diag.Errorf("Error parsing reference: %s", err)
	}

	opts := []remote.Option{remote.WithContext(ctx)}
	authenticators := meta.(map[string]authn.Authenticator)
	if auth, ok := data.GetOk("auth"); ok {
		opts = append(opts, remote.WithAuth(buildAuthenticator(auth.(map[string]interface{}))))
	} else if authenticator, ok := authenticators[ref.Context().RegistryStr()]; ok {
		opts = append(opts, remote.WithAuth(authenticator))
	} else {
		opts = append(opts, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	}

	img, err := remote.Image(ref, opts...)
	if err != nil {
		return diag.Errorf("Error querying image: %s", err)
	}
	hash, err := img.Digest()
	if err != nil {
		return diag.FromErr(err)
	}
	digest := hash.String()
	data.Set("digest", digest)
	data.SetId(digest)
	return nil
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Address of the registry",
						},
						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"auth": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"identity_token": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"registry_token": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		ConfigureContextFunc: providerConfigure,
		ResourcesMap:         map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"containerregistry_image": {
				ReadContext: dataSourceContainerRegistryImageRead,
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"digest": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"auth": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"username": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"auth": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
								"password": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
								"identity_token": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
								"registry_token": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: Provider})
}
