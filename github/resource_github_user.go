package github

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v28/github"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGithubUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubUserCreate,
		Read:   resourceGithubUserRead,
		Update: nil,
		Delete: resourceGithubUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubUserCreate(d *schema.ResourceData, meta interface{}) error {
	return errors.New("The github_user resource must be imported, it cannot be created.")
}

func resourceGithubUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	var user *github.User
	var resp *github.Response
	var err error

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	if err != nil {
		// when the resource is new, it will have just been imported, and the Id
		// will be a string containing the username, not a numeric Id
		log.Printf("[DEBUG] Reading user: %s", d.Id())
		user, resp, err = client.Users.Get(ctx, d.Id())
	} else {
		// the resource is not new, so the username->Id transformation has already been
		// performed
		log.Printf("[DEBUG] Reading user: %d", id)
		user, resp, err = client.Users.GetByID(ctx, id)
	}

	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing user %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(strconv.FormatInt(*user.ID, 10))
	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("username", user.Login)

	return nil
}

func resourceGithubUserDelete(d *schema.ResourceData, meta interface{}) error {
	// this operation cannot be performed, but should be silently ignored
	return nil
}
