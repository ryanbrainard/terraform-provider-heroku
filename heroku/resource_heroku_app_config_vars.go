package heroku

import "github.com/hashicorp/terraform/helper/schema"

func resourceHerokuAppConfigVars() *schema.Resource {
	return &schema.Resource{
		Create: resourceHerokuAppConfigVarsCreate,
		Read:   resourceHerokuAppConfigVarsRead,
		Update: resourceHerokuAppConfigVarsUpdate,
		Delete: resourceHerokuAppConfigVarsDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"public": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},

			"private": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
		},
	}
}
