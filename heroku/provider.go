package heroku

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-homedir"
	"os"
	"runtime"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_EMAIL", nil),
			},

			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_API_KEY", nil),
			},
			"headers": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_HEADERS", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"heroku_addon":                             resourceHerokuAddon(),
			"heroku_addon_attachment":                  resourceHerokuAddonAttachment(),
			"heroku_app":                               resourceHerokuApp(),
			"heroku_app_feature":                       resourceHerokuAppFeature(),
			"heroku_app_release":                       resourceHerokuAppRelease(),
			"heroku_cert":                              resourceHerokuCert(),
			"heroku_domain":                            resourceHerokuDomain(),
			"heroku_drain":                             resourceHerokuDrain(),
			"heroku_formation":                         resourceHerokuFormation(),
			"heroku_pipeline":                          resourceHerokuPipeline(),
			"heroku_pipeline_coupling":                 resourceHerokuPipelineCoupling(),
			"heroku_space":                             resourceHerokuSpace(),
			"heroku_space_inbound_ruleset":             resourceHerokuSpaceInboundRuleset(),
			"heroku_space_app_access":                  resourceHerokuSpaceAppAccess(),
			"heroku_space_peering_connection_accepter": resourceHerokuSpacePeeringConnectionAccepter(),
			"heroku_space_vpn_connection":              resourceHerokuSpaceVPNConnection(),
			"heroku_team_collaborator":                 resourceHerokuTeamCollaborator(),
			"heroku_team_member":                       resourceHerokuTeamMember(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"heroku_space":              dataSourceHerokuSpace(),
			"heroku_space_peering_info": dataSourceHerokuSpacePeeringInfo(),
			"heroku_app":                dataSourceHerokuApp(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{}

	headers := make(map[string]string)
	if h := d.Get("headers").(string); h != "" {
		if err := json.Unmarshal([]byte(h), &headers); err != nil {
			return nil, err
		}
	}

	h := make(http.Header)
	for k, v := range headers {
		h.Set(k, v)
	}

	// Read from netrc file first. If not available fall back to original authentication
	err := readNetrcFile(&config, h)
	if err != nil {
		log.Println("Unable to read from netrc. Failing back to default authentication.")

		config = Config{
			Email:   d.Get("email").(string),
			APIKey:  d.Get("api_key").(string),
			Headers: h,
		}
	}

	log.Println("[INFO] Initializing Heroku client")
	return &config, nil
}

func buildCompositeID(a, b string) string {
	return fmt.Sprintf("%s:%s", a, b)
}

func parseCompositeID(id string) (string, string) {
	parts := strings.SplitN(id, ":", 2)
	return parts[0], parts[1]
}

func readNetrcFile(config *Config, headers http.Header) error {
	// Get the netrc file path. If path not shown, then fall back to default netrc path value
	path := os.Getenv("NETRC_PATH")

	if path == "" {
		filename := ".netrc"
		if runtime.GOOS == "windows" {
			filename = "_netrc"
		}

		var err error
		path, err = homedir.Expand("~/" + filename)
		if err != nil {
			return err
		}
	}

	// If the file is not a file, then do nothing
	if fi, err := os.Stat(path); err != nil {
		// File doesn't exist, do nothing
		if os.IsNotExist(err) {
			return nil
		}

		// Some other error!
		return err
	} else if fi.IsDir() {
		// File is directory, ignore
		return nil
	}

	// Load up the netrc file
	net, err := netrc.ParseFile(path)
	if err != nil {
		return fmt.Errorf("error parsing netrc file at %q: %s", path, err)
	}

	machine := net.FindMachine("api.heroku.com")
	if machine == nil {
		// Machine not found, no problem
		return nil
	}

	// Set the user/api key/headers
	config.Email = machine.Login
	config.APIKey = machine.Password
	config.Headers = headers

	return nil
}
