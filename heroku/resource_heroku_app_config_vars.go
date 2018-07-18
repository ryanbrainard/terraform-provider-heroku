package heroku

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/cyberdelia/heroku-go/v3"
	"log"
	"fmt"
	"context"
)

func resourceHerokuAppConfigVars() *schema.Resource {
	return &schema.Resource{
		Create: resourceHerokuAppConfigVarsCreate, // There is no CREATE endpoint for config-vars
		Read:   resourceHerokuAppConfigVarsRead,
		Update: resourceHerokuAppConfigVarsUpdate,
		Delete: resourceHerokuAppConfigVarsDelete,
		// TODO: should we handle scenario where a private var is in the public one?

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

func resourceHerokuAppConfigVarsCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*heroku.Service)

	// Get App Name
	appName := getAppName(d)

	// Define the Public & Private vars
	var publicVars, privateVars map[string]interface{}
	if v, ok := d.GetOk("public"); ok {
		publicVars = v.(map[string]interface{})
	}
	if v, ok := d.GetOk("private"); ok {
		privateVars = v.(map[string]interface{})
	}

	// Combine `public` & `private` config vars together and remove duplicates
	configVars := mergeMaps(publicVars, privateVars)

	// Create Config Vars for App
	log.Printf("[INFO] Creating %s's config vars: *%#v", appName, configVars)

	if _, err := client.ConfigVarUpdate(context.TODO(), appName, configVars); err != nil {
		return fmt.Errorf("[ERROR] Error creating %s's config vars: %s", appName, err)
	}

	return resourceHerokuAppConfigVarsRead(d, meta)
}

func resourceHerokuAppConfigVarsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*heroku.Service)

	// Get App Name
	appName := getAppName(d)

	// Get the App Id that we will use as this resource's Id
	appUuid := getAppUuid(appName, client)

	configVars, err := client.ConfigVarInfoForApp(context.TODO(), appName)
	if err != nil {
		return err
	}

	d.SetId(appUuid)
	if err := d.Set("config_vars", configVars); err != nil {
		log.Printf("[WARN] Error setting config vars: %s", err)
	}

	return nil
}

func resourceHerokuAppConfigVarsUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceHerokuAppConfigVarsDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func mergeMaps(maps ...map[string]interface{}) map[string]*string {
	result := make(map[string]*string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v.(*string)
		}
	}
	return result
}
